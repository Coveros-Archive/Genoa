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
	EnvVarGithubReleaseFilesDir     = "GITHUB_DEPLOY_DIRECTORY"
	EnvVarGitlabReleaseFilesDir     = "GITLAB_DEPLOY_DIRECTORY"
	EnvVarGithubWebhookSecret       = "GITHUB_WEBHOOK_SECRET"
	EnvVarGitlabWebhookSecret       = "GITLAB_WEBHOOK_SECRET"
	EnvVarGithubPersonalAccessToken = "GITHUB_PERSONAL_ACCESS_TOKEN"
	EnvVarGitlabPersonalAccessToken = "GITLAB_PERSONAL_ACCESS_TOKEN"
)

var (
	GithubReleaseFilesDir string = os.Getenv(EnvVarGithubReleaseFilesDir)
	GithubWebhookSecret   string = os.Getenv(EnvVarGithubWebhookSecret)
	GithubAccessToken     string = os.Getenv(EnvVarGithubPersonalAccessToken)
	GitlabReleaseFilesDir string = os.Getenv(EnvVarGitlabReleaseFilesDir)
	GitlabWebhookSecret   string = os.Getenv(EnvVarGitlabWebhookSecret)
	GitlabAccessToken     string = os.Getenv(EnvVarGitlabPersonalAccessToken)
)

var log = logf.Log.WithName("gitSync.webhookHandler")

type WebhookHandler struct {
	Client client.Client
}

func webhookSecretAccessTokenReleaseDir(r *http.Request) (string, string, string) {
	var (
		secret      = GithubWebhookSecret
		accessToken = GithubAccessToken
		releaseDir  = GithubReleaseFilesDir
	)
	if r.Header.Get(git.GitlabEventHeaderKey) != "" {
		secret = GitlabWebhookSecret
		accessToken = GitlabAccessToken
		releaseDir = GitlabReleaseFilesDir
	}
	return secret, accessToken, releaseDir
}

func (wH WebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	webhookSecret, accessToken, releaseDir := webhookSecretAccessTokenReleaseDir(r)
	git := git.GitFactory(r, accessToken)
	log.Info(fmt.Sprintf("Git provider: %T", git))
	eventType, errParsingWebhookReq := git.ParseWebhook(r, webhookSecret)
	if errParsingWebhookReq != nil {
		log.Error(errParsingWebhookReq, "Failed to parse git webhook")
		return
	}
	switch e := eventType.(type) {
	case *github.PushEvent, *lab.PushEvent:
		wH.handleGitPushEvents(git.PushEventToPushEventMeta(e), releaseDir, git)
	default:
		log.Info("Git webhook event type not supported: %T ... skipping...", github.WebHookType(r))
		return
	}
}

func (wH WebhookHandler) handleGitPushEvents(e *git.PushEventMeta, releaseDir string, git git.Git) {
	for _, commit := range e.Commits {

		if len(commit.Added) > 0 {
			for _, eAdded := range commit.Added {
				if strings.HasPrefix(eAdded, releaseDir) {
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
				if strings.HasPrefix(eModified, releaseDir) {
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
				if strings.HasPrefix(eRemoved, releaseDir) {
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
