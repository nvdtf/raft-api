package gitkit

import (
	"context"
	"encoding/base64"

	"github.com/google/go-github/v45/github"
	"golang.org/x/oauth2"
)

type GitKitInterface interface {
	Process(owner string, repo string, network string) (*ProcessResult, error)
	Read(owner string, repo string, path string) ([]byte, error)
}

type GitKit struct {
	client *github.Client
}

func NewGitKit(githubToken string) GitKitInterface {
	ctx := context.Background()

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	return &GitKit{
		client: client,
	}
}

func (gk *GitKit) Read(owner string, repo string, path string) ([]byte, error) {
	ctx := context.Background()

	// fmt.Printf("Reading %s ...\n", path)

	result, _, _, err := gk.client.Repositories.GetContents(ctx, owner, repo, path, nil)
	if err != nil {
		return nil, err
	}

	contents, err := base64.StdEncoding.DecodeString(*result.Content)
	if err != nil {
		return nil, err
	}

	return contents, nil
}
