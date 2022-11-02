package api

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/go-redis/cache/v8"
	"github.com/nvdtf/raft-api/pkg/gitkit"
	"go.uber.org/zap"
)

type ProcessRepoParams struct {
	Owner   string `json:"owner"`
	Repo    string `json:"repo"`
	Network string `json:"network"`
}

func (api *Api) ProcessRepoHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	paramsBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		serverError(w, api.logger, err, "error reading request")
		return
	}

	params := &ProcessRepoParams{}
	json.Unmarshal(paramsBody, params)

	log := api.logger.With(
		"repo", params.Owner+"/"+params.Repo,
		"network", params.Network,
	)

	RegisterRepoProcessRequestMetrics(params.Owner+"/"+params.Repo, params.Network)

	latestCommit, err := api.kit.GetLatestCommitHash(ctx, params.Owner, params.Repo)
	if err != nil {
		serverError(w, api.logger, err, "error accessing repository")
		return
	}
	log.Info(latestCommit)

	var result *gitkit.ProcessResult
	err = api.cache.Once(&cache.Item{
		Key:   params.Owner + "/" + params.Repo + "-" + params.Network + "-" + latestCommit,
		Value: &result,
		Do: func(*cache.Item) (interface{}, error) {
			log.Info("Processing repo")
			RegisterRepoProcessCacheMissMetrics(params.Owner+"/"+params.Repo, params.Network)
			res, err := api.kit.Process(ctx, params.Owner, params.Repo, params.Network)
			return res, err
		},
	})
	if err != nil {
		serverError(w, api.logger, err, "error processing repo")
		return
	}

	log.Info("Repo processed")

	json.NewEncoder(w).Encode(result)
}

func serverError(
	w http.ResponseWriter,
	log *zap.SugaredLogger,
	err error,
	desc string,
) {
	log.With("error", err).Error(desc)
	w.WriteHeader(500)
	json.NewEncoder(w).Encode(&JSONError{Error: err.Error()})
}
