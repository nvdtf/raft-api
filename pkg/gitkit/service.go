package gitkit

import (
	"context"
	"encoding/base64"

	"github.com/google/go-github/v45/github"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

type GitKitInterface interface {
	Process(owner string, repo string, network string) (*ProcessResult, error)
	Read(owner string, repo string, path string) ([]byte, error)
}

type GitKit struct {
	client *github.Client
	logger *zap.SugaredLogger
}

func NewGitKit(
	githubToken string,
) (
	GitKitInterface,
	error,
) {
	ctx := context.Background()

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}

	return &GitKit{
		client,
		logger.Sugar(),
	}, nil
}

func (gk *GitKit) Read(owner string, repo string, path string) ([]byte, error) {
	ctx := context.Background()

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
