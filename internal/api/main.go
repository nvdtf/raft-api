package api

import "github.com/nvdtf/raft-api/pkg/gitkit"

type Api struct {
	kit gitkit.GitKitInterface
}

func NewApi(githubToken string) Api {
	gk := gitkit.NewGitKit(githubToken)
	return Api{
		kit: gk,
	}
}
