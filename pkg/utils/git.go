package utils

import (
	"context"
	"encoding/base64"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

func NewGitClient(accessToken string) *github.Client {
	// https://github.com/google/go-github#authentication
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: accessToken},
	)
	tc := oauth2.NewClient(context.TODO(), ts)
	return github.NewClient(tc)
}

func GetFileContentsFromGitInString(owner, repo, branch, file string, gClient *github.Client) (string, error) {
	getContentsOptions := &github.RepositoryContentGetOptions{Ref: branch}
	fileContent, _, _, err := gClient.Repositories.GetContents(
		context.TODO(), owner, repo, file, getContentsOptions)
	if err != nil {
		return "", err
	}
	decodedContentInBytes, errDecoding := base64.StdEncoding.DecodeString(*fileContent.Content)
	return string(decodedContentInBytes), errDecoding
}
