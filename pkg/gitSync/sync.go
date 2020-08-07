package gitSync

import (
	"context"
	"coveros.com/api/v1alpha1"
	"coveros.com/pkg/factories/git"
	"coveros.com/pkg/utils"
	"fmt"
	"k8s.io/client-go/kubernetes/scheme"
	"reflect"
)

func (wH WebhookHandler) syncReleaseWithGithub(owner, repo, branch, SHA, releaseFile string, gitFactory git.Git, isRemovedFromGithub bool) {
	var readFileFrom = branch
	if isRemovedFromGithub {
		readFileFrom = SHA
	}

	log.Info(fmt.Sprintf("Attempting to sync %v from %v/%v into cluster", releaseFile, owner, repo))
	gitFileContents, errReadingFromGit := gitFactory.GetFileContents(owner, repo, readFileFrom, releaseFile)
	if errReadingFromGit != nil {
		log.Error(errReadingFromGit, "Failed to get fileContents from git")
		return
	}

	hrFromGit := &v1alpha1.Release{}
	_, _, err := scheme.Codecs.UniversalDeserializer().Decode([]byte(gitFileContents), nil, hrFromGit)
	if err != nil {
		log.Error(err, "Could not decode release file from git, perhaps its not a release file..")
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
			log.Error(err, "Failed to delete Release which was removed from github")
			return
		}
		log.Info(fmt.Sprintf("Delete %v release from cluster initiated...", hrFromGit.GetName()))
		return
	}

	log.Info(fmt.Sprintf("Creating %v namespace if needed..", hrFromGit.GetNamespace()))
	if errCreatingNamespace := utils.CreateNamespace(hrFromGit.GetNamespace(), wH.Client); errCreatingNamespace != nil {
		log.Error(errCreatingNamespace, "Failed to create namespace")
		return
	}

	log.Info(fmt.Sprintf("Creating %v/%v release", hrFromGit.GetNamespace(), hrFromGit.GetName()))
	hrFromCluster, errCreatingHR := utils.CreateRelease(hrFromGit, wH.Client)
	if errCreatingHR != nil {
		log.Info(fmt.Sprintf("%v/%v failed to create release : %v", hrFromGit.GetNamespace(), hrFromGit.GetName(), errCreatingHR))
	}
	log.Info(fmt.Sprintf("Successfully created %v/%v Release", hrFromGit.GetNamespace(), hrFromGit.GetName()))

	specInSync := reflect.DeepEqual(hrFromCluster.Spec, hrFromGit.Spec)
	labelsInSync := reflect.DeepEqual(hrFromCluster.GetLabels(), hrFromGit.GetLabels())
	annotationsInSync := reflect.DeepEqual(hrFromCluster.GetAnnotations(), hrFromGit.GetAnnotations())
	if !specInSync || !labelsInSync || !annotationsInSync {
		hrFromCluster.SetAnnotations(hrFromGit.GetAnnotations())
		hrFromCluster.SetLabels(hrFromGit.GetLabels())
		hrFromCluster.Spec = hrFromGit.Spec
		if errUpdating := wH.Client.Update(context.TODO(), hrFromCluster); errUpdating != nil {
			log.Error(errUpdating, fmt.Sprintf("Failed to apply release from %v/%v - %v", owner, repo, hrFromGit.GetName()))
			return
		}

		log.Info(fmt.Sprintf("Updated release from %v/%v - %v/%v", owner, repo, hrFromGit.GetNamespace(), hrFromGit.GetName()))
	}

}
