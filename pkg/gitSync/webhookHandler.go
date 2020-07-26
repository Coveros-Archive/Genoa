package gitSync

import (
	"coveros.com/pkg/utils"
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

	event, err := parseWebhookPayload(r)
	if err != nil {
		return
	}

	switch e := event.(type) {
	case *github.PushEvent:
		wH.handleGithubPushEvents(e)
	default:
		log.Info("Github webhook event type not supported: %T ... skipping...", github.WebHookType(r))
		return
	}
}

func parseWebhookPayload(req *http.Request) (interface{}, error) {
	payload, err := github.ValidatePayload(req, []byte(WebhookSecret))
	if err != nil {
		log.Error(err, "error reading github request body")
		return nil, err
	}
	defer req.Body.Close()

	event, err := github.ParseWebHook(github.WebHookType(req), payload)
	if err != nil {
		log.Error(err, "could not parse webhook payload")
		return nil, err
	}

	return event, nil

}

func (wH WebhookHandler) handleGithubPushEvents(e *github.PushEvent) {
	gitClient := utils.NewGitClient(GithubAccessToken)
	log.Info(GithubAccessToken)
	for _, commit := range e.Commits {

		if len(commit.Added) > 0 {
			for _, eAdded := range commit.Added {
				if strings.HasPrefix(eAdded, ReleaseFilesDir) {
					wH.syncHelmReleaseWithGithub(
						e.GetRepo().GetOwner().GetName(),
						e.GetRepo().GetName(),
						strings.Replace(*e.Ref, "refs/heads/", "", -1),
						eAdded, gitClient, false)
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
						eModified, gitClient, false)
				}
			}
		}

		if len(commit.Removed) > 0 {
			for _, eRemoved := range commit.Removed {
				if strings.HasPrefix(eRemoved, ReleaseFilesDir) {
					wH.syncHelmReleaseWithGithub(
						e.GetRepo().GetOwner().GetName(),
						e.GetRepo().GetName(),
						e.GetBefore(),
						eRemoved, gitClient, true)
				}
			}
		}
	}
}
