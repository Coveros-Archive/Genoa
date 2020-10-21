package git

import (
	"coveros.com/pkg"
	"os"
)

type github struct{}

func NewGithub() *github {
	return &github{}
}

func (g github) GetAccessToken() (string, error) {
	token, ok := os.LookupEnv(pkg.EnvVarGithubPersonalAccessToken)
	if !ok || token == "" {
		return "", ErrorEnvVarNotFound{Message: "Github personal access token not specified in env vars"}
	}
	return token, nil
}

func (g github) GetDeployDir() (string, error) {
	deployDir, ok := os.LookupEnv(pkg.EnvVarGithubReleaseFilesDir)
	if !ok || deployDir == "" {
		return "", ErrorEnvVarNotFound{Message: "Github deploy dir not specified in env vars"}
	}
	return deployDir, nil
}
func (g github) GetWebhookSecret() (string, error) {
	secret, ok := os.LookupEnv(pkg.EnvVarGithubWebhookSecret)
	if !ok || secret == "" {
		return "", ErrorEnvVarNotFound{Message: "Github webhook secret not specified in env vars"}
	}
	return secret, nil
}

func (g github) GetSelfHostedUrl() (string, error) {
	url, ok := os.LookupEnv(pkg.EnvVarGithubEnterpriseHostedUrl)
	if !ok || url == "" {
		return "", ErrorEnvVarNotFound{Message: "Github enterprise hosted url not specified in env vars"}
	}
	return url, nil
}
