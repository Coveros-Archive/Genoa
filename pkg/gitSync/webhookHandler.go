package gitSync

import (
	"coveros.com/pkg/factories/git"
	"fmt"
	cNotifyLib "github.com/coveros/notification-library"
	"github.com/go-logr/logr"
	"github.com/google/go-github/github"
	lab "github.com/xanzy/go-gitlab"
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
	git := git.Factory(r)
	wH.Logger.Info(fmt.Sprintf("Git provider: %T", git))
	eventType, errParsingWebhookReq := git.ParseWebhook(r)
	if errParsingWebhookReq != nil {
		wH.Logger.Error(errParsingWebhookReq, "Failed to parse git webhook")
		return
	}
	switch e := eventType.(type) {
	case *github.PushEvent, *lab.PushEvent:
		wH.handleGitPushEvents(git.PushEventToPushEventMeta(e), git)
	default:
		wH.Logger.Info("Git webhook event type not supported: %T ... skipping...", github.WebHookType(r))
		return
	}
}

func (wH WebhookHandler) handleGitPushEvents(e *git.PushEventMeta, git git.Git) {
	for _, commit := range e.Commits {

		if len(commit.Added) > 0 {
			for _, eAdded := range commit.Added {
				if strings.HasPrefix(eAdded, git.GetDeployDir()) {
					wH.syncReleaseWithGithub(
						e.Owner, e.Repo,
						strings.Replace(e.Ref, "refs/heads/", "", -1),
						commit.SHA,
						eAdded, git, false)
				}
			}
		}

		if len(commit.Modified) > 0 {
			for _, eModified := range commit.Modified {
				if strings.HasPrefix(eModified, git.GetDeployDir()) {
					wH.syncReleaseWithGithub(
						e.Owner, e.Repo,
						strings.Replace(e.Ref, "refs/heads/", "", -1),
						commit.SHA,
						eModified, git, false)
				}
			}
		}

		if len(commit.Removed) > 0 {
			for _, eRemoved := range commit.Removed {
				if strings.HasPrefix(eRemoved, git.GetDeployDir()) {
					wH.syncReleaseWithGithub(
						e.Owner, e.Repo,
						strings.Replace(e.Ref, "refs/heads/", "", -1),
						e.Before,
						eRemoved, git, true)
				}
			}
		}
	}
}
