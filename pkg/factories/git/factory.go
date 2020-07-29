package git

import (
	"coveros.com/pkg/factories/git/github"
	"coveros.com/pkg/factories/git/gitlab"
	"net/http"
)

type GitType string

const (
	GITHUB GitType = "github"
	GITLAB GitType = "gitlab"
)

type Git interface {
	GetFileContents(owner, repo, branch, file string) (string, error)
	ParseWebhook(req *http.Request, webhookSecret string) (interface{}, error)
}

func GitFactory(g GitType, token string) Git {
	switch g {
	case GITHUB:
		return github.NewGithub(token)
	case GITLAB:
		return gitlab.NewGitlab(token)
	}
	return nil
}
