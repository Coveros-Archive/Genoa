package gitSync

import (
	"context"
	"coveros.com/api/v1alpha1"
	"coveros.com/pkg/utils"
	"fmt"
	"github.com/agill17/go-scm/scm"
	cNotifyLib "github.com/coveros/notification-library"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes/scheme"
	"reflect"
)

func (wH WebhookHandler) syncReleaseWithGithub(ownerRepo, branch, SHA, releaseFile string, scmClient *scm.Client, isRemovedFromGithub bool) {
	var readFileFrom = branch
	if isRemovedFromGithub {
		readFileFrom = SHA
	}

	wH.Logger.Info(fmt.Sprintf("Attempting to sync %v from %v into cluster", releaseFile, ownerRepo))

	scmFileContents, _, errGettingFileContents := scmClient.Contents.Find(context.TODO(), ownerRepo, releaseFile, readFileFrom)
	if errGettingFileContents != nil {
		wH.Logger.Error(errGettingFileContents, "Failed to get file contents from git")
		return
	}
	gitFileContents := string(scmFileContents.Data)

	hrFromGit := &v1alpha1.Release{}
	_, gvk, err := scheme.Codecs.UniversalDeserializer().Decode([]byte(gitFileContents), nil, hrFromGit)
	if err != nil {
		wH.Logger.Error(err, "Could not decode release file from git, perhaps its not a release file..")
		return
	}

	if gvk.Kind != "Release" && gvk.GroupVersion() != v1alpha1.GroupVersion {
		wH.Logger.Info(fmt.Sprintf("Not a valid release %v from %v git repo", releaseFile, ownerRepo))
		return
	}

	if hrFromGit.Spec.ValuesOverride.V == nil {
		hrFromGit.Spec.ValuesOverride.V = map[string]interface{}{}
	}

	if hrFromGit.GetNamespace() == "" {
		hrFromGit.SetNamespace("default")
	}

	notificationChannel := utils.GetChannelIDForNotification(hrFromGit.ObjectMeta)

	namespacedName := fmt.Sprintf("%s/%s", hrFromGit.GetNamespace(), hrFromGit.GetName())
	ownerRepoBranch := fmt.Sprintf("%v@%v", ownerRepo, branch)

	// if GitBranchToFollowAnnotation is specified, we ONLY create/update CR's if the current source branch is the same as GitBranchToFollow
	// this way users can have same CR's exist on many branches but only apply updates from the GitBranchToFollow
	if branchToFollow, ok := hrFromGit.Annotations[utils.GitBranchToFollowAnnotation]; ok && branchToFollow != "" {
		if branchToFollow != branch {
			wH.Logger.Info(fmt.Sprintf("%v from %v, follow-git-branch '%v' does not match current branch '%v'",
				hrFromGit.GetName(), ownerRepo, branchToFollow, branch))
			return
		}
	}

	if isRemovedFromGithub {
		if err := wH.Client.Delete(context.TODO(), hrFromGit); err != nil {
			if errors.IsNotFound(err) {
				wH.Logger.Info(fmt.Sprintf("%v/%v release not found, skipping clean up..", hrFromGit.GetNamespace(), hrFromGit.GetName()))
				return
			}
			wH.Logger.Error(err, "Failed to delete Release which was removed from github")
			return
		}
		wH.Logger.Info(fmt.Sprintf("Delete %v release from cluster initiated...", hrFromGit.GetName()))
		wH.Notify.SendMsg(cNotifyLib.NotifyTemplate{
			Channel:   notificationChannel,
			Title:     ":interrobang:" + namespacedName,
			EventType: cNotifyLib.Warning,
			Fields: map[string]string{
				"Reason":       "Delete release from cluster initiated",
				"Git-Source":   ownerRepoBranch,
				"Release-File": releaseFile,
			},
		})
		return
	}

	wH.Logger.Info(fmt.Sprintf("Creating %v namespace if needed..", hrFromGit.GetNamespace()))
	if errCreatingNamespace := utils.CreateNamespace(hrFromGit.GetNamespace(), wH.Client); errCreatingNamespace != nil {
		wH.Logger.Error(errCreatingNamespace, "Failed to create namespace")
		return
	}

	wH.Logger.Info(fmt.Sprintf("Creating %v/%v release", hrFromGit.GetNamespace(), hrFromGit.GetName()))
	hrFromCluster, errCreatingHR := utils.CreateRelease(hrFromGit, wH.Client)
	if errCreatingHR != nil {
		wH.Logger.Info(fmt.Sprintf("%v failed to create release : %v", namespacedName, errCreatingHR))
		wH.Notify.SendMsg(cNotifyLib.NotifyTemplate{
			Channel:   notificationChannel,
			Title:     namespacedName,
			EventType: cNotifyLib.Failure,
			Fields: map[string]string{
				"Reason":       fmt.Sprintf("Failed to create release: %v", errCreatingHR),
				"Git-Source":   ownerRepoBranch,
				"Release-File": releaseFile,
			},
		})
	}
	wH.Logger.Info(fmt.Sprintf("Successfully created %v release in your cluster", namespacedName))
	wH.Notify.SendMsg(cNotifyLib.NotifyTemplate{
		Channel:   notificationChannel,
		Title:     ":rocket:" + namespacedName,
		EventType: cNotifyLib.Success,
		Fields: map[string]string{
			"Reason":       "Successfully created release in your cluster",
			"Git-Source":   ownerRepoBranch,
			"Release-File": releaseFile,
		},
	})

	specInSync := reflect.DeepEqual(hrFromCluster.Spec, hrFromGit.Spec)
	labelsInSync := reflect.DeepEqual(hrFromCluster.GetLabels(), hrFromGit.GetLabels())
	annotationsInSync := reflect.DeepEqual(hrFromCluster.GetAnnotations(), hrFromGit.GetAnnotations())
	if !specInSync || !labelsInSync || !annotationsInSync {
		hrFromCluster.SetAnnotations(hrFromGit.GetAnnotations())
		hrFromCluster.SetLabels(hrFromGit.GetLabels())
		hrFromCluster.Spec = hrFromGit.Spec
		if errUpdating := wH.Client.Update(context.TODO(), hrFromCluster); errUpdating != nil {
			wH.Logger.Error(errUpdating, fmt.Sprintf("Failed to apply release from %v - %v", ownerRepo, namespacedName))
			wH.Notify.SendMsg(cNotifyLib.NotifyTemplate{
				Channel:   notificationChannel,
				Title:     namespacedName,
				EventType: cNotifyLib.Failure,
				Fields: map[string]string{
					"Reason":       fmt.Sprintf("Failed to apply release from %v git repo: %v", ownerRepo, errUpdating),
					"Git-Source":   ownerRepoBranch,
					"Release-File": releaseFile,
				},
			})
			return
		}

		wH.Logger.Info(fmt.Sprintf("Updated release from %v - %v", ownerRepo, namespacedName))
		wH.Notify.SendMsg(cNotifyLib.NotifyTemplate{
			Channel:   notificationChannel,
			Title:     ":rocket:" + namespacedName,
			EventType: cNotifyLib.Success,
			Fields: map[string]string{
				"Reason":       "Successfully updated release in your cluster",
				"Git-Source":   ownerRepoBranch,
				"Release-File": releaseFile,
			},
		})
	}

}
