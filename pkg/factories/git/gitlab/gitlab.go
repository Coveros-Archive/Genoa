package gitlab

import "net/http"

func NewGitlab(token string) *gitlab {
	return &gitlab{token: token}
}

//TODO: add implementation
func (g *gitlab) GetFileContents(owner, repo, branch, file string) (string, error) {
	//NO-OP
	return "", nil
}

//TODO: add implementation
func (g *gitlab) ParseWebhook(req *http.Request, webhookSecret string) (interface{}, error) {
	//NO-OP
	return nil, nil
}
