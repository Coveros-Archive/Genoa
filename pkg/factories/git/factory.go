package git

import (
	"net/http"
	"os"
)

const (
	GithubEventHeaderKey            = "X-Github-Event"
	GitlabEventHeaderKey            = "X-Gitlab-Event"
	GitlabWebhookSecretHeaderKey    = "X-Gitlab-Token"
	EnvVarGithubReleaseFilesDir     = "GITHUB_DEPLOY_DIRECTORY"
	EnvVarGitlabReleaseFilesDir     = "GITLAB_DEPLOY_DIRECTORY"
	EnvVarGithubWebhookSecret       = "GITHUB_WEBHOOK_SECRET"
	EnvVarGitlabWebhookSecret       = "GITLAB_WEBHOOK_SECRET"
	EnvVarGithubPersonalAccessToken = "GITHUB_PERSONAL_ACCESS_TOKEN"
	EnvVarGitlabPersonalAccessToken = "GITLAB_PERSONAL_ACCESS_TOKEN"
)

type Git interface {
	GetFileContents(owner, repo, branch, file string) (string, error)
	ParseWebhook(req *http.Request) (interface{}, error)
	PushEventToPushEventMeta(pushEvent interface{}) *PushEventMeta
	GetDeployDir() string
}

func Factory(req *http.Request) Git {
	isGithubReq := req.Header.Get(GithubEventHeaderKey)
	if isGithubReq != "" {
		return NewGithub(os.Getenv(EnvVarGithubPersonalAccessToken))
	}

	isGitlab := req.Header.Get(GitlabEventHeaderKey)
	if isGitlab != "" {
		return NewGitlab(os.Getenv(EnvVarGitlabPersonalAccessToken))
	}
	return nil
}

// custom push event struct to deal with factory pattern for diff git providers
type PushEventMeta struct {
	Ref     string
	Commits []Commit
	SHA     string
	Before  string
	After   string
	Repo    string
	Owner   string
}
type Commit struct {
	Added    []string
	Removed  []string
	Modified []string
	SHA      string
}
