# RFC3339 (to match GoReleaser)
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT := $(shell git rev-parse HEAD)
GO_LDFLAGS := -ldflags="-s -w -X 'main.version=$(GIT_COMMIT)' -X 'main.commit=$(GIT_COMMIT)' -X 'main.date=$(DATE)'"

all: test build_service build_cli

build_service: clean_service
	CGO_ENABLED=0 GOARCH=amd64 go build -trimpath $(GO_LDFLAGS) -o build/service ./service/

build_cli: clean_cli
	CGO_ENABLED=0 GOARCH=amd64 go build -trimpath $(GO_LDFLAGS) -o build/cello ./cli/

lint:
	@#Install the linter from here:
	@#https://github.com/golangci/golangci-lint#install
	golangci-lint run

test:
	go test -race -timeout=180s -coverprofile=coverage.out ./...

tidy:
	go mod tidy

cover: ## Generates coverage report
	@$(MAKE) test TESTARGS="-tags test -coverprofile=coverage.out"
	@go tool cover -html=coverage.out
	@rm -f coverage.out

clean_service:
	@rm -f ./build/service

clean_cli:
	@rm -f ./build/cello

up: ## Starts a local vault and api locally
	bash scripts/start_local.sh dev

.PHONY: build_service build_cli lint test tidy cover clean_cli clean_service up
