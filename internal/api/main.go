package api

import (
	"github.com/nvdtf/raft-api/pkg/gitkit"
	"go.uber.org/zap"
)

type Api struct {
	kit    gitkit.GitKitInterface
	logger *zap.SugaredLogger
}

func NewApi(
	githubToken string,
) (
	*Api,
	error,
) {
	gk, err := gitkit.NewGitKit(githubToken)
	if err != nil {
		return nil, err
	}

	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}

	return &Api{
		kit:    gk,
		logger: logger.Sugar(),
	}, nil
}

type JSONError struct {
	Error string `json:"error"`
}
