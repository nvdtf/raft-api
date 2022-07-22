package gitkit

import (
	"context"
	"encoding/base64"
	"strings"

	"github.com/google/go-github/v45/github"
)

type GitKitInterface interface {
	Process(network string) (*ProcessResult, error)
	Read(path string) ([]byte, error)
	GetContractDeployments(network string) (map[string]string, error)
}

type GitKit struct {
	client *github.Client

	owner string
	repo  string
}

func NewGitKit(owner string, repo string) GitKitInterface {
	client := github.NewClient(nil)

	return &GitKit{
		client: client,

		owner: owner,
		repo:  repo,
	}
}

func (gk *GitKit) Read(path string) ([]byte, error) {
	ctx := context.Background()

	// fmt.Printf("Reading %s ...\n", path)

	result, _, _, err := gk.client.Repositories.GetContents(ctx, gk.owner, gk.repo, path, nil)
	if err != nil {
		return nil, err
	}

	contents, err := base64.StdEncoding.DecodeString(*result.Content)
	if err != nil {
		return nil, err
	}

	return contents, nil
}

func (gk *GitKit) GetContractDeployments(network string) (map[string]string, error) {
	flowJson, err := gk.Read(FLOW_JSON_PATH)
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
				result[name] = a
			}
		}
	}

	return result, nil
}
