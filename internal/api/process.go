package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type ProcessRepoParams struct {
	Owner   string `json:"owner"`
	Repo    string `json:"repo"`
	Network string `json:"network"`
}

func (api *Api) ProcessRepoHandler(w http.ResponseWriter, r *http.Request) {
	paramsBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	params := &ProcessRepoParams{}
	json.Unmarshal(paramsBody, params)

	log := api.logger.With(
		"repo", params.Owner+"/"+params.Repo,
		"network", params.Network,
	)
	log.Info("Processing repo")

	result, err := api.kit.Process(params.Owner, params.Repo, params.Network)
	if err != nil {
		log.With("error", err).Error("Error processing repo")

		w.WriteHeader(500)
		json.NewEncoder(w).Encode(&JSONError{Error: err.Error()})
		return
	}

	log.Info("Repo processed")

	json.NewEncoder(w).Encode(result)
}
