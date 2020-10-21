package git

import (
	"coveros.com/pkg"
	"net/http"
)

type GitProvider string

const (
	Github GitProvider = "github"
	Gitlab GitProvider = "gitlab"
	Noop   GitProvider = "noop"
	// TODO: Add more as needed
)

// helper factory to lookup the correct env vars based on git provider webhook payload
type Git interface {
	GetAccessToken() (string, error)
	GetDeployDir() (string, error)
	GetWebhookSecret() (string, error)
	GetSelfHostedUrl() (string, error)
}

func NewGitFactory(webhookReq *http.Request) (Git, GitProvider) {
	isGithubReq := webhookReq.Header.Get(pkg.GithubEventHeaderKey)
	if isGithubReq != "" {
		return NewGithub(), Github
	}
	isGitlab := webhookReq.Header.Get(pkg.GitlabEventHeaderKey)
	if isGitlab != "" {
		return NewGitlab(), Gitlab
	}
	return nil, Noop
}

func NewGitFactoryBasedOnProvider(provider GitProvider) Git {
	switch provider {
	case Github:
		return NewGithub()
	case Gitlab:
		return NewGitlab()
	}
	return nil
}
