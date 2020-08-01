package gitSync

import (
	"coveros.com/pkg/factories/git"
	"fmt"
	"github.com/google/go-github/github"
	lab "github.com/xanzy/go-gitlab"
	"net/http"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"strings"
)

const (
	EnvVarReleaseFilesDir           = "DEPLOY_DIRECTORY"
	EnvVarWebhookSecret             = "WEBHOOK_SECRET"
	EnvVarGithubPersonalAccessToken = "GITHUB_PERSONAL_ACCESS_TOKEN"
)

var ReleaseFilesDir string
var WebhookSecret string
var GithubAccessToken string

var log = logf.Log.WithName("gitSync.webhookHandler")

type WebhookHandler struct {
	Client client.Client
}

func init() {
	if val, ok := os.LookupEnv(EnvVarReleaseFilesDir); ok {
		ReleaseFilesDir = val
	}

	if val, ok := os.LookupEnv(EnvVarWebhookSecret); ok {
		WebhookSecret = val
	}

	if val, ok := os.LookupEnv(EnvVarGithubPersonalAccessToken); ok {
		GithubAccessToken = val
	}
}

func (wH WebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	// based on the payload, we switch between github, gitlab, etc
	git := git.GitFactory(r, GithubAccessToken)
	log.Info(fmt.Sprintf("Webhook provider: %T", git))
	eventType, errParsingWebhookReq := git.ParseWebhook(r, WebhookSecret)
	if errParsingWebhookReq != nil {
		log.Error(errParsingWebhookReq, "Failed to parse git webhook")
		return
	}
	switch e := eventType.(type) {
	case *github.PushEvent, *lab.PushEvent:
		wH.handleGitPushEvents(git.PushEventToPushEventMeta(e), git)
	default:
		log.Info("Git webhook event type not supported: %T ... skipping...", github.WebHookType(r))
		return
	}
}

func (wH WebhookHandler) handleGitPushEvents(e *git.PushEventMeta, git git.Git) {
	for _, commit := range e.Commits {

		if len(commit.Added) > 0 {
			for _, eAdded := range commit.Added {
				if strings.HasPrefix(eAdded, ReleaseFilesDir) {
					wH.syncHelmReleaseWithGithub(
						e.Owner, e.Repo,
						strings.Replace(e.Ref, "refs/heads/", "", -1),
						commit.SHA,
						eAdded, git, false)
				}
			}
		}

		if len(commit.Modified) > 0 {
			for _, eModified := range commit.Modified {
				if strings.HasPrefix(eModified, ReleaseFilesDir) {
					wH.syncHelmReleaseWithGithub(
						e.Owner, e.Repo,
						strings.Replace(e.Ref, "refs/heads/", "", -1),
						commit.SHA,
						eModified, git, false)
				}
			}
		}

		if len(commit.Removed) > 0 {
			for _, eRemoved := range commit.Removed {
				if strings.HasPrefix(eRemoved, ReleaseFilesDir) {
					wH.syncHelmReleaseWithGithub(
						e.Owner, e.Repo,
						strings.Replace(e.Ref, "refs/heads/", "", -1),
						e.Before,
						eRemoved, git, true)
				}
			}
		}
	}
}
