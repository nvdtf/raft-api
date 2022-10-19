include .env

.SILENT: dev
dev:
	GITHUB_TOKEN=${GITHUB_TOKEN} go run cmd/raft-api/*.go

.SILENT: run
run:
	docker-compose up --build -d

.SILENT: stop
stop:
	docker-compose down