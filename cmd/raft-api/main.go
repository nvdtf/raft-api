package main

import (
	"log"
	"net/http"
	"os"

	"github.com/nvdtf/raft-api/internal/api"

	"github.com/gorilla/mux"
)

func main() {
	api := api.NewApi(os.Getenv("GITHUB_TOKEN"))

	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/processRepo", api.ProcessRepoHandler)

	log.Fatal(http.ListenAndServe(":8080", router))

	// temp()
}
