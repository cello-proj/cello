// +build !test

package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/argoproj-labs/argo-cloudops/internal/env"
	"github.com/argoproj-labs/argo-cloudops/service/internal/credentials"
	"github.com/argoproj-labs/argo-cloudops/service/internal/workflow"

	"github.com/argoproj/argo-workflows/v3/cmd/argo/commands/client"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	vault "github.com/hashicorp/vault/api"
)

func main() {
	var (
		logger = log.With(log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout)), "ts", log.DefaultTimestampUTC)
	)

	env, err := env.GetEnv()
	if err != nil {
		panic(fmt.Sprintf("Unable to initialize environment variables %s", err))
	}

	setLogLevel(&logger, env.LogLevel)

	level.Info(logger).Log("message", fmt.Sprintf("loading config '%s'", env.ConfigFilePath))
	config, err := loadConfig(env.ConfigFilePath)
	if err != nil {
		panic(fmt.Sprintf("Unable to load config %s", err))
	}
	level.Info(logger).Log("message", fmt.Sprintf("loading config '%s' completed", env.ConfigFilePath))

	gitClient, err := newBasicGitClient(env.SSHPEMFile)
	if err != nil {
		level.Error(logger).Log("message", "error creating git client", "error", err)
		panic("error creating git client")
	}

	// The Argo context is needed for any Argo client method calls or else, nil errors.
	argoCtx, argoClient := client.NewAPIClient()

	// Any Argo Workflow client method calls need the context returned from NewAPIClient, otherwise
	// nil errors will occur. Mux sets its params in context, so passing the Argo Workflow context to
	// setupRouter and applying it to the request will wipe out Mux vars (or any other data Mux sets in its context).
	h := handler{
		logger:                 logger,
		newCredentialsProvider: credentials.NewVaultProvider,
		argo:                   workflow.NewArgoWorkflow(argoClient.NewWorkflowServiceClient(), env.ArgoNamespace),
		argoCtx:                argoCtx,
		config:                 config,
		gitClient:              gitClient,
		env:                    env,
		// Function that the handler will use to create a Vault svc using the vaultConfig below.
		// Vault svc needs to be created within the handler methods because Vault uses headers
		// to add metadata (e.g. transaction ID) to its logs.
		newCredsProviderSvc: credentials.NewVaultSvc,
		// Needed to pass some Vault config to the handlers to be able to create
		// a Vault service within the handlers.
		vaultConfig: vaultConfig{
			config: &vault.Config{Address: env.VaultAddress},
			role:   env.VaultRole,
			secret: env.VaultSecret,
		},
	}

	level.Info(logger).Log("message", "starting web service", "vault addr", env.VaultAddress, "argoAddr", env.ArgoAddress)

	r := setupRouter(h)
	err = http.ListenAndServeTLS(fmt.Sprintf(":%d", env.Port), "ssl/certificate.crt", "ssl/certificate.key", r)
	if err != nil {
		level.Error(logger).Log("message", "error starting service", "error", err)
		panic("error starting service")
	}
}

func setLogLevel(logger *log.Logger, logLevel string) {
	switch logLevel {
	case "DEBUG":
		*logger = level.NewFilter(*logger, level.AllowDebug())
	default:
		*logger = level.NewFilter(*logger, level.AllowInfo())
	}
}
