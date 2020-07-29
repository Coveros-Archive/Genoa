package github

import (
	"context"
	"encoding/base64"
	googleGithub "github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"net/http"
)

func NewGithub(token string) *github {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(context.TODO(), ts)
	return &github{token: token, client: googleGithub.NewClient(tc)}
}

func (g *github) GetFileContents(owner, repo, branch, file string) (string, error) {
	getContentsOptions := &googleGithub.RepositoryContentGetOptions{Ref: branch}
	fileContent, _, _, err := g.client.Repositories.GetContents(
		context.TODO(), owner, repo, file, getContentsOptions)
	if err != nil {
		return "", err
	}
	decodedContentInBytes, errDecoding := base64.StdEncoding.DecodeString(*fileContent.Content)
	return string(decodedContentInBytes), errDecoding
}

func (g *github) ParseWebhook(req *http.Request, webhookSecret string) (interface{}, error) {
	payload, err := googleGithub.ValidatePayload(req, []byte(webhookSecret))
	if err != nil {
		return nil, err
	}
	defer req.Body.Close()
	return googleGithub.ParseWebHook(googleGithub.WebHookType(req), payload)
}
