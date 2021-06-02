package main

import (
	"fmt"
	"net/http"
	"os"

	int_env "github.com/argoproj-labs/argo-cloudops/internal/env"
	"github.com/argoproj-labs/argo-cloudops/service/internal/workflow"
	"github.com/argoproj/argo-workflows/v3/cmd/argo/commands/client"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	vault "github.com/hashicorp/vault/api"
)

var env = int_env.GetEnv()

func main() {
	var (
		logger = log.With(log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout)), "ts", log.DefaultTimestampUTC)
	)

	setLogLevel(&logger, env.LogLevel)

	level.Info(logger).Log("message", fmt.Sprintf("loading config '%s'", env.ConfigFilePath))
	config, err := loadConfig()
	if err != nil {
		panic(fmt.Sprintf("Unable to load config %s", err))
	}
	level.Info(logger).Log("message", fmt.Sprintf("loading config '%s' completed", env.ConfigFilePath))

	vaultSvc, err := newVaultSvc(env.VaultAddress, env.VaultRole, env.VaultSecret)
	if err != nil {
		level.Error(logger).Log("message", "error creating vault service client", "error", err)
		panic("error creating vault service client")
	}

	gitClient, err := newBasicGitClient(env.SshPemFile)
	if err != nil {
		level.Error(logger).Log("message", "error creating git client", "error", err)
		panic("error creating git client")
	}

	_, argoClient := client.NewAPIClient()

	h := handler{
		logger:                 logger,
		newCredentialsProvider: newVaultProvider(vaultSvc),
		argo:                   workflow.NewArgoWorkflow(argoClient.NewWorkflowServiceClient(), env.ArgoNamespace),
		config:                 config,
		gitClient:              gitClient,
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

// TODO before open sourcing we should provide the token instead of generating it
func newVaultSvc(vaultAddr, role, secret string) (*vault.Client, error) {
	config := &vault.Config{
		Address: vaultAddr,
	}

	vaultSvc, err := vault.NewClient(config)
	if err != nil {
		return nil, err

	}

	options := map[string]interface{}{
		"role_id":   role,
		"secret_id": secret,
	}

	sec, err := vaultSvc.Logical().Write("auth/approle/login", options)
	if err != nil {
		return nil, err
	}

	vaultSvc.SetToken(sec.Auth.ClientToken)
	return vaultSvc, nil
}
