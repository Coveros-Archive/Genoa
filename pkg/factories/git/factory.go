package git

import (
	"net/http"
)

const (
	GithubEventHeaderKey         = "X-Github-Event"
	GitlabEventHeaderKey         = "X-Gitlab-Event"
	GitlabWebhookSecretHeaderKey = "X-Gitlab-Token"
)

type Git interface {
	GetFileContents(owner, repo, branch, file string) (string, error)
	ParseWebhook(req *http.Request, webhookSecret string) (interface{}, error)
	PushEventToPushEventMeta(pushEvent interface{}) *PushEventMeta
}

func GitFactory(req *http.Request, token string) Git {
	isGithubReq := req.Header.Get(GithubEventHeaderKey)
	if isGithubReq != "" {
		return NewGithub(token)
	}

	isGitlab := req.Header.Get(GitlabEventHeaderKey)
	if isGitlab != "" {
		return NewGitlab(token)
	}
	return nil
}

// custom push event struct to deal with factory pattern for diff git providers
type PushEventMeta struct {
	Head    string
	Ref     string
	Size    int
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
