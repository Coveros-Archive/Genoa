package github

import (
	googleGithub "github.com/google/go-github/github"
)

type github struct {
	token  string
	client *googleGithub.Client
}
