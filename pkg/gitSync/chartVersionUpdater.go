package gitSync

import (
	"context"
	"coveros.com/api/v1alpha1"
	"coveros.com/pkg"
	"coveros.com/pkg/factories/git"
	v3 "coveros.com/pkg/helm/v3"
	"coveros.com/pkg/utils"
	"errors"
	"fmt"
	"github.com/agill17/go-scm/scm"
	scmFactory "github.com/agill17/go-scm/scm/factory"
	"github.com/go-logr/logr"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
	"strings"
	"sync"
	"time"
)

// This defines how a chart version is updated in a git project

type ChartVersionSync struct {
	Client ctrlClient.Client
	Logger logr.Logger
	HelmV3 *v3.HelmV3
}

var wg sync.WaitGroup

func (c ChartVersionSync) StartSync() {
	for {

		time.Sleep(time.Minute * 2)

		// list releases
		releases := v1alpha1.ReleaseList{}
		if err := c.Client.List(context.TODO(), &releases); err != nil {
			c.Logger.Error(err, "Failed to list releases")
			return
		}

		wg.Add(len(releases.Items))

		// re-index cache for all helm repos
		if err := c.HelmV3.RefreshRepoIndex("all"); err != nil {
			c.Logger.Error(err, "Failed to refresh index")
			return
		}

		for _, release := range releases.Items {
			if _, ok := release.GetAnnotations()[pkg.SkipChartSync]; !ok {
				go c.StartVersionSync(&release)
			}
		}
		wg.Wait()
		c.Logger.Info("Successfully synced chart version updates for all releases")
	}
}

func (c *ChartVersionSync) StartVersionSync(release *v1alpha1.Release) {

	// fetch git details from CR annotations
	provider, ownerRepo, branch, releaseFilePath, err := c.ensureReqAnnotationsExists(release)
	if err != nil {
		c.Logger.Error(err, "Required annotations do not exist on release CR")
		wg.Done()
		return
	}

	repoChart := release.Spec.Chart
	repoAlias, chartName := strings.Split(repoChart, "/")[0], strings.Split(repoChart, "/")[1]
	versionInCluster := release.Spec.Version

	latestVersion, err := c.HelmV3.GetLatestChartVersionAvailable(repoAlias, chartName)
	if err != nil {
		c.Logger.Error(err, "Failed to find out if there is a new chart version available")
		wg.Done()
		return
	}

	if latestVersion != "" && latestVersion != versionInCluster {

		gitFactory := git.NewGitFactoryBasedOnProvider(git.GitProvider(provider))
		url, _ := gitFactory.GetSelfHostedUrl()
		accessToken, err := gitFactory.GetAccessToken()
		if err != nil {
			c.Logger.Error(err, "Failed to get access token for git provider")
			wg.Done()
			return
		}
		scmClient, err := scmFactory.NewClient(provider, url, accessToken)
		if err != nil {
			c.Logger.Error(err, "Failed to setup scmClient")
			wg.Done()
			return
		}
		// meaning git-branch-to-follow annotation does not exist, assume default branch for that repo
		if branch == "" {
			branch, err = utils.GetDefaultBranch(scmClient, ownerRepo)
			if err != nil {
				c.Logger.Error(err, "Failed to get default branch name")
				wg.Done()
				return
			}
		}

		hrFromGit, gitContent, err := utils.GetReleaseFileFromGit(scmClient, ownerRepo, releaseFilePath, branch)
		if err != nil {
			c.Logger.Error(err, "Failed to get release file from git")
			wg.Done()
			return
		}

		// make sure whats in git also does not match
		if hrFromGit.Spec.Version != latestVersion {
			newBranchName := fmt.Sprintf("update/%v-%v", release.GetNamespace(), release.GetName())
			if errCreatingBranch := utils.CreateBranch(scmClient, ownerRepo, branch, newBranchName); errCreatingBranch != nil {
				c.Logger.Error(errCreatingBranch, "Failed to create branch")
				wg.Done()
				return
			}

			hrFromGit.Spec.Version = latestVersion
			prData, err := yaml.Marshal(release)
			if err != nil {
				c.Logger.Error(err, "Failed to marshal release into []byte")
				wg.Done()
				return
			}

			if _, errUpdating := scmClient.Contents.Update(context.TODO(), ownerRepo, releaseFilePath, &scm.ContentParams{
				Branch:  newBranchName,
				Message: "test",
				Data:    prData,
				Sha:     gitContent.Sha,
			}); errUpdating != nil {
				c.Logger.Error(errUpdating, "Failed to update file contents in git")
				wg.Done()
				return
			}

			pr, errCreatingPR := utils.CreatePR(scmClient, ownerRepo, "test pr title", newBranchName, branch)
			if errCreatingPR != nil {
				c.Logger.Error(errCreatingPR, "Failed to create PR")
				wg.Done()
				return
			}
			c.Logger.Info(pr.Link)
		}
	}
	wg.Done()
}

func (c *ChartVersionSync) ensureReqAnnotationsExists(release *v1alpha1.Release) (string, string, string, string, error) {
	namespacedName := fmt.Sprintf("%v/%v", release.GetNamespace(), release.GetName())
	annotations := release.GetAnnotations()
	// git url not found? log it and return
	gitProvider, foundGitProvider := annotations[pkg.GitProvider]
	if !foundGitProvider {
		c.Logger.Info(fmt.Sprintf("%v : does not have a git provider annotation set... skipping", namespacedName))
		return "", "", "", "", errors.New("ErrorGitProviderAnnotationDoesNotExist")
	}

	// release file path not found? log it and return
	releaseFile, foundReleaseFile := annotations[pkg.ReleaseFilePath]
	if !foundReleaseFile {
		c.Logger.Info(fmt.Sprintf("%v : does not have release file path annotation set... skipping", namespacedName))
		return "", "", "", "", errors.New("ErrorReleaseFilePathAnnotationDoesNotExist")
	}

	//ownerRepo now found? log it and return
	ownerRepo, foundOwnerRepo := annotations[pkg.GitOwnerRepo]
	if !foundOwnerRepo {
		c.Logger.Info(fmt.Sprintf("%v : does not have git-owner-repo annotation set... skipping", namespacedName))
		return "", "", "", "", errors.New("ErrorGitOwnerRepoAnnotationDoesNotExist")
	}

	return gitProvider, ownerRepo, annotations[pkg.GitBranchToFollowAnnotation], releaseFile, nil
}
