GO_LDFLAGS ?= -ldflags="-s -w"

all: test build_service build_cli

build_service: clean_service
	CGO_ENABLED=0 GOARCH=amd64 go build -trimpath $(GO_LDFLAGS) $(BUILDARGS) -o build/service ./service/

build_cli: clean_cli
	CGO_ENABLED=0 GOARCH=amd64 go build -trimpath $(GO_LDFLAGS) $(BUILDARGS) -o build/argo-cloudops ./cli/...

lint:
	@#Install the linter from here:
	@#https://github.com/golangci/golangci-lint#install
	golangci-lint run

test:
	env ARGO_CLOUDOPS_ADMIN_SECRET="D34DB33FD34DB33FD34DB33FD34DB33F" VAULT_ADDR="1.2.3.4" ARGO_ADDR="2.3.4.5" VAULT_ROLE="vault-role" \
	VAULT_SECRET="pw123" SSH_PEM_FILE="~/.ssh/id_rsa.pub" ARGO_CLOUDOPS_CONFIG=../service/testdata/argo-cloudops.yaml \
	go test -race -timeout=180s -coverprofile=coverage.out ./service #github.com/argoproj-labs/argo-cloudops

tidy:
	go mod tidy

vet: ## Runs go vet
	go vet $(VETARGS) ./...

cover: ## Generates coverage report
	@$(MAKE) test TESTARGS="-tags test -coverprofile=coverage.out"
	@go tool cover -html=coverage.out
	@rm -f coverage.out

clean_service:
	@rm -f ./build/service

clean_cli:
	@rm -f ./build/argo-cloudops

up: ## Starts a local vault and api locally
	bash scripts/start_local.sh

.PHONY: build_service build_cli lint test tidy vet cover clean_cli clean_service up
