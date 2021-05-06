package main

import (
	"fmt"
	"net/http"
	"os"

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
		logLevel    = os.Getenv("ARGO_CLOUD_OPS_LOG_LEVEL")
		port        = os.Getenv("ARGO_CLOUD_OPS_PORT")
	)

	setLogLevel(&logger, logLevel)

	if len(adminSecret()) < 16 {
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

	if port == "" {
		port = "8443"
	}

	level.Info(logger).Log("message", fmt.Sprintf("loading config '%s'", configFilePath()))
	config, err := loadConfig()
	if err != nil {
		panic(fmt.Sprintf("Unable to load config %s", err))
	}
	level.Info(logger).Log("message", fmt.Sprintf("loading config '%s' completed", configFilePath()))

	vaultSvc, err := newVaultSvc(vaultAddr, vaultRole, vaultSecret)
	if err != nil {
		level.Error(logger).Log("message", "error creating vault service client", "error", err)
		panic("error creating vault service client")
	}

	h := handler{
		logger:                 logger,
		newCredentialsProvider: newVaultProvider(vaultSvc),
		argo:                   newArgoWorkflow(),
		config:                 config,
	}

	level.Info(logger).Log("message", "starting web service", "vault addr", vaultAddr, "argoAddr", argoAddr)

	r := setupRouter(h)
	err = http.ListenAndServeTLS(fmt.Sprintf(":%s", port), "ssl/certificate.crt", "ssl/certificate.key", r)
	if err != nil {
		level.Error(logger).Log("message", "error starting service", "error", err)
		panic("error starting service")
	}

	return
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
