package gitSync

import (
	"coveros.com/pkg/factories/git"
	"fmt"
	"github.com/agill17/go-scm/scm"
	scmFactory "github.com/agill17/go-scm/scm/factory"
	cNotifyLib "github.com/coveros/notification-library"
	"github.com/go-logr/logr"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

type WebhookHandler struct {
	Client client.Client
	Logger logr.Logger
	Notify cNotifyLib.Notify
}

func (wH WebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	gitFactory, gitProvider := git.NewGitFactory(r)

	// explicitly ignoring errors because not everyone will be using self-hosted provider
	url, _ := gitFactory.GetSelfHostedUrl()

	token, errGettingToken := gitFactory.GetAccessToken()
	if errGettingToken != nil {
		wH.Logger.Error(errGettingToken, "Failed to get SCM access token")
		return
	}

	deployDir, errGettingDeployDir := gitFactory.GetDeployDir()
	if errGettingDeployDir != nil {
		wH.Logger.Error(errGettingDeployDir, "Failed to get deployDir")
		return
	}

	scmClient, errGettingClient := scmFactory.NewClient(string(gitProvider), url, token)
	if errGettingClient != nil {
		wH.Logger.Error(errGettingClient, "Failed to set up a SCM client...")
		return
	}
	webhook, errParsingWebhook := scmClient.Webhooks.Parse(r, func(webhook scm.Webhook) (string, error) {
		return gitFactory.GetWebhookSecret()
	})
	if errParsingWebhook != nil {
		wH.Logger.Error(errParsingWebhook, "Failed to parse git webhook...")
		return
	}

	wH.Logger.Info(fmt.Sprintf("Webhook Kind: %v", webhook.Kind()))

	switch o := webhook.(type) {
	case *scm.PushHook:
		wH.handleGitPushEvents(o, scmClient, deployDir)
		return
	default:
		wH.Logger.Info("%T webhook event is not yet supported", o)
		return
	}

}

func (wH WebhookHandler) handleGitPushEvents(pushHook *scm.PushHook, scmClient *scm.Client, deployDir string) {
	repoFullName := pushHook.Repository().FullName
	branch := strings.Replace(pushHook.Ref, "refs/heads/", "", -1)
	for _, commit := range pushHook.Commits {
		if len(commit.Added) > 0 {
			for _, eAdded := range commit.Added {
				if strings.HasPrefix(eAdded, deployDir) {
					wH.syncReleaseWithGithub(
						repoFullName, branch,
						pushHook.Commit.Sha,
						eAdded, scmClient, false)
				}
			}
		}
		if len(commit.Modified) > 0 {
			for _, eModified := range commit.Modified {
				if strings.HasPrefix(eModified, deployDir) {
					wH.syncReleaseWithGithub(
						repoFullName, branch,
						pushHook.Commit.Sha,
						eModified, scmClient, false)
				}
			}
		}
		if len(commit.Removed) > 0 {
			for _, eRemoved := range commit.Removed {
				if strings.HasPrefix(eRemoved, deployDir) {
					wH.syncReleaseWithGithub(
						repoFullName,
						branch, pushHook.Before,
						eRemoved, scmClient, true)
				}
			}
		}
	}
}
