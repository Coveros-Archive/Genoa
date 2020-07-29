package gitSync

import (
	"coveros.com/pkg/factories/git"
	"github.com/google/go-github/github"
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
	//TODO: based on the payload, we should switch between github, gitlab, etc but I dont know what to switch it based off of yet
	git := git.GitFactory(git.GITHUB, GithubAccessToken)
	eventType, errParsingWebhookReq := git.ParseWebhook(r, WebhookSecret)
	if errParsingWebhookReq != nil {
		return
	}

	switch e := eventType.(type) {
	case *github.PushEvent:
		wH.handleGithubPushEvents(e, git)
	default:
		log.Info("Git webhook event type not supported: %T ... skipping...", github.WebHookType(r))
		return
	}
}

func (wH WebhookHandler) handleGithubPushEvents(e *github.PushEvent, git git.Git) {
	for _, commit := range e.Commits {

		if len(commit.Added) > 0 {
			for _, eAdded := range commit.Added {
				if strings.HasPrefix(eAdded, ReleaseFilesDir) {
					wH.syncHelmReleaseWithGithub(
						e.GetRepo().GetOwner().GetName(),
						e.GetRepo().GetName(),
						strings.Replace(*e.Ref, "refs/heads/", "", -1),
						commit.GetSHA(),
						eAdded, git, false)
				}
			}
		}

		if len(commit.Modified) > 0 {
			for _, eModified := range commit.Modified {
				if strings.HasPrefix(eModified, ReleaseFilesDir) {
					wH.syncHelmReleaseWithGithub(
						e.GetRepo().GetOwner().GetName(),
						e.GetRepo().GetName(),
						strings.Replace(*e.Ref, "refs/heads/", "", -1),
						commit.GetSHA(),
						eModified, git, false)
				}
			}
		}

		if len(commit.Removed) > 0 {
			for _, eRemoved := range commit.Removed {
				if strings.HasPrefix(eRemoved, ReleaseFilesDir) {
					wH.syncHelmReleaseWithGithub(
						e.GetRepo().GetOwner().GetName(),
						e.GetRepo().GetName(),
						strings.Replace(*e.Ref, "refs/heads/", "", -1),
						e.GetBefore(),
						eRemoved, git, true)
				}
			}
		}
	}
}
