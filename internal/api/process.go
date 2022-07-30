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

	result, err := api.kit.Process(params.Owner, params.Repo, params.Network)
	if err != nil {
		panic(err)
	}

	json.NewEncoder(w).Encode(result)
}