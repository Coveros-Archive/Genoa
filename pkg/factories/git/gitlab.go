package git

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	lab "github.com/xanzy/go-gitlab"
)

type gitlab struct {
	client *lab.Client
}

func NewGitlab(token string) *gitlab {
	client, err := lab.NewClient(token)
	if err != nil {
		return nil
	}
	return &gitlab{client: client}
}

func (g *gitlab) GetFileContents(owner, repo, branch, file string) (string, error) {
	ownerRepo := fmt.Sprintf("%s/%s", owner, repo)
	option := lab.GetFileOptions{Ref: &branch}
	f, resp, err := g.client.RepositoryFiles.GetFile(ownerRepo, file, &option)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	decodedContentInBytes, errDecoding := base64.StdEncoding.DecodeString(f.Content)
	return string(decodedContentInBytes), errDecoding
}

func (g *gitlab) ParseWebhook(req *http.Request) (interface{}, error) {
	if req.Header.Get(GitlabWebhookSecretHeaderKey) != os.Getenv(EnvVarGitlabWebhookSecret) {
		return nil, errors.New("WebhookSecretDoesNotMatch")
	}
	eventType := lab.HookEventType(req)
	reqBody, errReading := ioutil.ReadAll(req.Body)
	if errReading != nil {
		return nil, errReading
	}
	defer req.Body.Close()
	return lab.ParseWebhook(eventType, reqBody)
}

func (g *gitlab) PushEventToPushEventMeta(pushEvent interface{}) *PushEventMeta {
	pE, ok := pushEvent.(*lab.PushEvent)
	if !ok {
		return nil
	}
	ownerRepo := strings.Split(pE.Project.PathWithNamespace, "/")
	pEMeta := &PushEventMeta{
		Ref:     pE.Ref,
		Before:  pE.Before,
		After:   pE.After,
		Repo:    ownerRepo[1],
		Owner:   ownerRepo[0],
		Commits: make([]Commit, len(pE.Commits)),
	}

	for i := 0; i <= len(pE.Commits)-1; i++ {
		pEMeta.Commits[i].Added = pE.Commits[i].Added
		pEMeta.Commits[i].Removed = pE.Commits[i].Removed
		pEMeta.Commits[i].Modified = pE.Commits[i].Modified
		pEMeta.Commits[i].SHA = pE.Commits[i].ID
	}

	return pEMeta
}

func (g *gitlab) GetDeployDir() string {
	return os.Getenv(EnvVarGitlabReleaseFilesDir)
}
