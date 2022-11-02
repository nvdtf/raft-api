package api

import (
	"github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
	"github.com/nvdtf/raft-api/pkg/gitkit"
	"go.uber.org/zap"
)

type Api struct {
	kit    gitkit.GitKitInterface
	logger *zap.SugaredLogger
	cache  *cache.Cache
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

	rdb := redis.NewClient(&redis.Options{
		Addr: "redis:6379",
	})

	mycache := cache.New(&cache.Options{
		Redis: rdb,
		// LocalCache: cache.NewTinyLFU(1000, time.Hour),
	})

	return &Api{
		kit:    gk,
		logger: logger.Sugar(),
		cache:  mycache,
	}, nil
}

type JSONError struct {
	Error string `json:"error"`
}
