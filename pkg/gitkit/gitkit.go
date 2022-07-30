package gitkit

import (
	"context"
	"encoding/base64"
	"strings"

	"github.com/google/go-github/v45/github"
	"golang.org/x/oauth2"
)

type GitKitInterface interface {
	Process(owner string, repo string, network string) (*ProcessResult, error)
	Read(owner string, repo string, path string) ([]byte, error)
	GetContractDeployments(owner string, repo string, network string) (map[string]string, error)
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

func (gk *GitKit) GetContractDeployments(owner string, repo string, network string) (map[string]string, error) {
	flowJson, err := gk.Read(owner, repo, FLOW_JSON_PATH)
	if err != nil {
		return nil, err
	}

	contracts, err := parseJson(flowJson)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for name, j := range contracts.Contracts {
		for n, a := range j.Advanced.Aliases {
			if strings.Trim(n, " ") == network {
				result[name] = strings.TrimPrefix(a, "0x")
			}
		}
	}

	return result, nil
}
