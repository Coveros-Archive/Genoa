package gitSync

import (
	"context"
	"coveros.com/pkg/factories/git"
	"coveros.com/pkg/utils"
	"fmt"
	"reflect"
)

func (wH WebhookHandler) syncHelmReleaseWithGithub(owner, repo, branch, SHA, releaseFile string, gitFactory git.Git, isRemovedFromGithub bool) {
	var readFileFrom = branch
	if isRemovedFromGithub {
		readFileFrom = SHA
	}

	log.Info(fmt.Sprintf("Attempting to sync %v from %v/%v into cluster", releaseFile, owner, repo))
	gitFileContents, errReadingFromGit := gitFactory.GetFileContents(owner, repo, readFileFrom, releaseFile)
	if errReadingFromGit != nil {
		log.Error(errReadingFromGit, "Failed to get fileContents from github")
		return
	}

	hrFromGit, unMarshalErr := utils.UnMarshalStringDataToHelmRelease(gitFileContents)
	if unMarshalErr != nil {
		log.Info("Failed to unmarshal")
		return
	}

	if hrFromGit == nil {
		log.Info(fmt.Sprintf("%v is not a valid HelmRelease, therefore skipping", releaseFile))
		return
	}

	if hrFromGit.Spec.ValuesOverride.V == nil {
		hrFromGit.Spec.ValuesOverride.V = map[string]interface{}{}
	}

	if hrFromGit.GetNamespace() == "" {
		hrFromGit.SetNamespace("default")
	}

	// if GitBranchToFollowAnnotation is specified, we ONLY create/update CR's if the current source branch is the same as GitBranchToFollow
	// this way users can have same CR's exist on many branches but only apply updates from the GitBranchToFollow
	if branchToFollow, ok := hrFromGit.Annotations[utils.GitBranchToFollowAnnotation]; ok && branchToFollow != "" {
		if branchToFollow != branch {
			log.Info(fmt.Sprintf("Skipping sync for %v from %v/%v, follow-git-branch '%v' does not match current branch '%v'", hrFromGit.GetName(), owner, repo, branchToFollow, branch))
			return
		}
	}

	if isRemovedFromGithub {
		if err := wH.Client.Delete(context.TODO(), hrFromGit); err != nil {
			log.Error(err, "Failed to delete HelmRelease which was removed from github")
			return
		}
		log.Info(fmt.Sprintf("Delete %v HelmRelease from cluster initiated...", hrFromGit.GetName()))
		return
	}

	log.Info(fmt.Sprintf("Creating %v namespace if needed..", hrFromGit.GetNamespace()))
	if errCreatingNamespace := utils.CreateNamespace(hrFromGit.GetNamespace(), wH.Client); errCreatingNamespace != nil {
		log.Error(errCreatingNamespace, "Failed to create namespace")
		return
	}

	log.Info(fmt.Sprintf("Creating %v/%v HelmRelease", hrFromGit.GetNamespace(), hrFromGit.GetName()))
	hrFromCluster, errCreatingHR := utils.CreateHelmRelease(hrFromGit, wH.Client)
	if errCreatingHR != nil {
		log.Info(fmt.Sprintf("%v/%v failed to create helmRelease : %v", hrFromGit.GetNamespace(), hrFromGit.GetName(), errCreatingHR))
	}
	log.Info(fmt.Sprintf("Successfully created %v/%v HelmRelease", hrFromGit.GetNamespace(), hrFromGit.GetName()))

	specInSync := reflect.DeepEqual(hrFromCluster.Spec, hrFromGit.Spec)
	labelsInSync := reflect.DeepEqual(hrFromCluster.GetLabels(), hrFromGit.GetLabels())
	annotationsInSync := reflect.DeepEqual(hrFromCluster.GetAnnotations(), hrFromGit.GetAnnotations())
	if !specInSync || !labelsInSync || !annotationsInSync {
		hrFromCluster.SetAnnotations(hrFromGit.GetAnnotations())
		hrFromCluster.SetLabels(hrFromGit.GetLabels())
		hrFromCluster.Spec = hrFromGit.Spec
		if errUpdating := wH.Client.Update(context.TODO(), hrFromCluster); errUpdating != nil {
			log.Error(errUpdating, fmt.Sprintf("Failed to apply HelmRelease from %v/%v - %v", owner, repo, hrFromGit.GetName()))
			return
		}

		log.Info(fmt.Sprintf("Updated HelmRelease from %v/%v - %v/%v", owner, repo, hrFromGit.GetNamespace(), hrFromGit.GetName()))
	}

}
