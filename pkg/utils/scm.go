package utils

import (
	"context"
	"coveros.com/api/v1alpha1"
	"coveros.com/pkg"
	"fmt"
	"github.com/agill17/go-scm/scm"
	"k8s.io/client-go/kubernetes/scheme"
)

// Utils related to go-scm client / go-scm package

func GetReleaseFileFromGit(scmClient *scm.Client, ownerRepo, filePath, ref string) (*v1alpha1.Release, *scm.Content, error) {
	gitFileContents, _, errGettingFileContents := scmClient.Contents.Find(context.TODO(), ownerRepo, filePath, ref)
	if errGettingFileContents != nil {
		return nil, nil, errGettingFileContents
	}

	hrFromGit := &v1alpha1.Release{}
	_, gvk, errDecoding := scheme.Codecs.UniversalDeserializer().Decode(gitFileContents.Data, nil, hrFromGit)
	if errDecoding != nil {
		return nil, gitFileContents, errDecoding
	}

	if gvk.Kind != "Release" && gvk.GroupVersion() != v1alpha1.GroupVersion {
		return nil, gitFileContents, pkg.ErrorFileContentIsNotReleaseFile{Message: fmt.Sprintf("%v is not a valid release file", filePath)}
	}

	return hrFromGit, gitFileContents, nil
}

func CreateBranch(scmClient *scm.Client, ownerRepo, fromBranch, branchName string) error {
	fromBranchHead, err := getLatestShaOfBranch(scmClient, ownerRepo, fromBranch)
	if err != nil {
		return err
	}

	if _, _, err := scmClient.Git.FindBranch(context.TODO(), ownerRepo, branchName); err != nil {

		if scmClient.Driver == scm.DriverGithub {
			branchName = fmt.Sprintf("refs/heads/%s", branchName)
		}

		_, _, errCreating := scmClient.Git.CreateRef(context.TODO(), ownerRepo, branchName, fromBranchHead)
		return errCreating
	}
	return nil
}

func getLatestShaOfBranch(scmClient *scm.Client, ownerRepo, branch string) (string, error) {
	sha, _, err := scmClient.Git.FindRef(context.TODO(), ownerRepo, fmt.Sprintf("heads/%s", branch))
	return sha, err
}

func GetDefaultBranch(scmClient *scm.Client, ownerRepo string) (string, error) {
	repoContents, _, errGettingRepo := scmClient.Repositories.Find(context.TODO(), ownerRepo)
	return repoContents.Branch, errGettingRepo
}

func CreatePR(scmClient *scm.Client, ownerRepo string, prTitle, fromBranch, targetBranch string) (*scm.PullRequest, error) {
	pr, _, err := scmClient.PullRequests.Create(context.TODO(), ownerRepo, &scm.PullRequestInput{
		Title: prTitle,
		Head:  fromBranch,
		Base:  targetBranch,
		Body:  "",
	})
	return pr, err
}
