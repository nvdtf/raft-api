package gitkit

import (
	"context"
	"encoding/base64"
	"errors"

	"github.com/google/go-github/v56/github"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

type GitKitInterface interface {
	Process(ctx context.Context, owner string, repo string, network string) (*ProcessResult, error)
	Read(ctx context.Context, owner string, repo string, path string) ([]byte, error)
	GetLatestCommitHash(ctx context.Context, owner string, repo string) (string, error)
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

func (gk *GitKit) Read(ctx context.Context, owner string, repo string, path string) ([]byte, error) {
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

func (gk *GitKit) GetLatestCommitHash(ctx context.Context, owner string, repo string) (string, error) {
	res, _, err := gk.client.Repositories.ListCommits(ctx, owner, repo, &github.CommitsListOptions{ListOptions: github.ListOptions{PerPage: 1}})
	if err != nil {
		return "", err
	}

	if len(res) == 0 {
		return "", errors.New("no commits available")
	}

	return *res[0].SHA, nil
}
