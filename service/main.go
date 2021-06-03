package main

import (
	"fmt"
	"net/http"
	"os"

	acoEnv "github.com/argoproj-labs/argo-cloudops/internal/env"
	"github.com/argoproj-labs/argo-cloudops/service/internal/workflow"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	vault "github.com/hashicorp/vault/api"
)

func main() {
	var (
		logger      = log.With(log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout)), "ts", log.DefaultTimestampUTC)
		vaultRole   = os.Getenv("VAULT_ROLE")
		vaultSecret = os.Getenv("VAULT_SECRET")
		vaultAddr   = os.Getenv("VAULT_ADDR")
		argoAddr    = os.Getenv("ARGO_ADDR")
		sshPemFile  = os.Getenv("SSH_PEM_FILE")
		logLevel    = os.Getenv("ARGO_CLOUD_OPS_LOG_LEVEL")
		port        = acoEnv.Getenv("ARGO_CLOUD_OPS_PORT", "8443")
	)

	setLogLevel(&logger, logLevel)

	if len(acoEnv.AdminSecret()) < 16 {
		panic("ARGO_CLOUDOPS_ADMIN_SECRET must be 16 characers long.")
	}

	if vaultRole == "" {
		panic("VAULT_ROLE is undefined")
	}

	if vaultSecret == "" {
		panic("VAULT_SECRET is undefined")
	}

	if vaultAddr == "" {
		panic("VAULT_ADDR is undefined")
	}

	if argoAddr == "" {
		panic("ARGO_ADDR is undefined")
	}

	if sshPemFile == "" {
		panic("SSH_PEM_FILE is undefined")
	}

	level.Info(logger).Log("message", fmt.Sprintf("loading config '%s'", acoEnv.ConfigFilePath()))
	config, err := loadConfig()
	if err != nil {
		panic(fmt.Sprintf("Unable to load config %s", err))
	}
	level.Info(logger).Log("message", fmt.Sprintf("loading config '%s' completed", acoEnv.ConfigFilePath()))

	vaultSvc, err := newVaultSvc(vaultAddr, vaultRole, vaultSecret)
	if err != nil {
		level.Error(logger).Log("message", "error creating vault service client", "error", err)
		panic("error creating vault service client")
	}

	gitClient, err := newBasicGitClient(sshPemFile)
	if err != nil {
		level.Error(logger).Log("message", "error creating git client", "error", err)
		panic("error creating git client")
	}

	_, argoClient, err := workflow.NewArgoAPIClient(argoAddr, logger)
	if err != nil {
		level.Error(logger).Log("message", "error creating argo client", "error", err)
		panic("error creating argo client")
	}

	namespace := acoEnv.ArgoNamespace()

	h := handler{
		logger:                 logger,
		newCredentialsProvider: newVaultProvider(vaultSvc),
		argo:                   workflow.NewArgoWorkflow(argoClient.NewWorkflowServiceClient(), namespace),
		config:                 config,
		gitClient:              gitClient,
	}

	level.Info(logger).Log("message", "starting web service", "vault addr", vaultAddr, "argoAddr", argoAddr)

	r := setupRouter(h)
	err = http.ListenAndServeTLS(fmt.Sprintf(":%s", port), "ssl/certificate.crt", "ssl/certificate.key", r)
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
