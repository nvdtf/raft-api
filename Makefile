include .env

.SILENT: run
run:
	GITHUB_TOKEN=${GITHUB_TOKEN} go run cmd/raft-api/*.go