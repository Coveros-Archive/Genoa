package utils

import (
	"context"
	"encoding/base64"
	"github.com/google/go-github/github"
)

func NewGitClient() *github.Client {
	return github.NewClient(nil)
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
