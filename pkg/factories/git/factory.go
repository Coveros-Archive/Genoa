package git

import (
	"coveros.com/pkg/factories/git/github"
	"coveros.com/pkg/factories/git/gitlab"
	"net/http"
)

type Git interface {
	GetFileContents(owner, repo, branch, file string) (string, error)
	ParseWebhook(req *http.Request, webhookSecret string) (interface{}, error)
}

// TODO: add support for other providers and add implementation to each inorder to satisfy Git
func GitFactory(req *http.Request, token string) Git {
	var githubHeaderEvent = "X-GitHub-Event"
	var gitlabHeaderEvent = "X-Gitlab-Event"

	isGithubReq := req.Header.Get(githubHeaderEvent)
	if isGithubReq != "" {
		return github.NewGithub(token)
	}

	isGitlab := req.Header.Get(gitlabHeaderEvent)
	if isGitlab != "" {
		return gitlab.NewGitlab(token)
	}

	return nil

}
