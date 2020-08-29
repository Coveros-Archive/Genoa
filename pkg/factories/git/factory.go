package git

import (
	"coveros.com/pkg/utils"
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
	isGithubReq := webhookReq.Header.Get(utils.GithubEventHeaderKey)
	if isGithubReq != "" {
		return NewGithub(), Github
	}
	isGitlab := webhookReq.Header.Get(utils.GitlabEventHeaderKey)
	if isGitlab != "" {
		return NewGitlab(), Gitlab
	}
	return nil, Noop
}
