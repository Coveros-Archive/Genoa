package git

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	lab "github.com/xanzy/go-gitlab"
)

type gitlab struct {
	token  string
	client *lab.Client
}

func NewGitlab(token string) *gitlab {
	client, err := lab.NewClient(token)
	if err != nil {
		return nil
	}
	return &gitlab{token: token, client: client}
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

func (g *gitlab) ParseWebhook(req *http.Request, webhookSecret string) (interface{}, error) {
	if req.Header.Get(GitlabWebhookSecretHeaderKey) != webhookSecret {
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

	pEMeta := &PushEventMeta{
		Head:   pE.CheckoutSHA,
		Ref:    pE.Ref,
		Size:   pE.TotalCommitsCount,
		Before: pE.Before,
		After:  pE.After,
		Repo:   pE.Project.Name,
		Owner:  pE.Project.Namespace,
	}

	for i := 0; i <= len(pE.Commits)-1; i++ {
		pEMeta.Commits[i].Added = pE.Commits[i].Added
		pEMeta.Commits[i].Removed = pE.Commits[i].Removed
		pEMeta.Commits[i].Modified = pE.Commits[i].Modified
		pEMeta.Commits[i].SHA = pE.Commits[i].ID
	}

	return pEMeta
}
