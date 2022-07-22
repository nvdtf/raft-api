package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/nvdtf/raft-api/pkg/gitkit"
)

type ProcessRepoParams struct {
	Owner   string `json:"owner"`
	Repo    string `json:"repo"`
	Network string `json:"network"`
}

func ProcessRepoHandler(w http.ResponseWriter, r *http.Request) {
	paramsBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	params := &ProcessRepoParams{}
	json.Unmarshal(paramsBody, params)

	gk := gitkit.NewGitKit(params.Owner, params.Repo)

	result, err := gk.Process(params.Network)
	if err != nil {
		panic(err)
	}

	json.NewEncoder(w).Encode(result)
}

func main() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/processRepo", ProcessRepoHandler)
	log.Fatal(http.ListenAndServe(":8080", router))
}
