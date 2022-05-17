package env

import (
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// #nosec
const testSecret = "tha5hei2Hee5le8n"

var prefixedEnvVars = map[string]string{
	"_ADMIN_SECRET":                 testSecret,
	"_WORKFLOW_EXECUTION_NAMESPACE": "argo-ns",
	"_CONFIG":                       "/app/test/config/path",
	"_GIT_AUTH_METHOD":              "https",
	"_GIT_HTTPS_USER":               "testuser",
	"_GIT_HTTPS_PASS":               "testpass",
	"_LOG_LEVEL":                    "DEBUG",
	"_PORT":                         "1234",
	"_DB_HOST":                      "localhost",
	"_DB_NAME":                      "argocloudops",
	"_DB_USER":                      "argoco",
	"_DB_PASSWORD":                  "1234",
}

var nonPrefixedEnvVars = map[string]string{
	"VAULT_ROLE":   "vaultRole",
	"VAULT_SECRET": testSecret,
	"VAULT_ADDR":   "1.2.3.4",
	"ARGO_ADDR":    "2.3.4.5",
	"SSH_PEM_FILE": "/app/test/ssh.pem",
}

func reset() {
	for k := range prefixedEnvVars {
		os.Unsetenv(appPrefix + k)
	}
	for k := range nonPrefixedEnvVars {
		os.Unsetenv(k)
	}

	instance = Vars{}
	once = sync.Once{}
}

func setEnvVars(vars map[string]string, prefix string) {
	for k, v := range vars {
		os.Setenv(prefix+k, v)
	}
}

func TestGetEnv(t *testing.T) {
	// Given
	reset()
	setEnvVars(prefixedEnvVars, appPrefix)
	setEnvVars(nonPrefixedEnvVars, "")

	// When
	vars, err := GetEnv()

	// Then
	assert.NoError(t, err)
	assert.Equal(t, testSecret, vars.AdminSecret)
	assert.Equal(t, "vaultRole", vars.VaultRole)
	assert.Equal(t, testSecret, vars.VaultSecret)
	assert.Equal(t, "1.2.3.4", vars.VaultAddress)
	assert.Equal(t, "argo-ns", vars.ArgoNamespace)
	assert.Equal(t, "/app/test/config/path", vars.ConfigFilePath)
	assert.Equal(t, "/app/test/ssh.pem", vars.SSHPEMFile)
	assert.Equal(t, "https", vars.GitAuthMethod)
	assert.Equal(t, "testuser", vars.GitHTTPSUser)
	assert.Equal(t, "testpass", vars.GitHTTPSPass)
	assert.Equal(t, "DEBUG", vars.LogLevel)
	assert.Equal(t, 1234, vars.Port)
	assert.Equal(t, "localhost", vars.DBHost)
	assert.Equal(t, "argocloudops", vars.DBName)
	assert.Equal(t, "argoco", vars.DBUser)
	assert.Equal(t, "1234", vars.DBPassword)
}

func TestDefaults(t *testing.T) {
	// Given
	reset()
	os.Setenv(appPrefix+"_ADMIN_SECRET", testSecret)
	os.Setenv("VAULT_ROLE", "vaultRole")
	os.Setenv("VAULT_SECRET", testSecret)
	os.Setenv("VAULT_ADDR", "1.2.3.4")
	os.Setenv("ARGO_ADDR", "2.3.4.5")
	os.Setenv(appPrefix+"_GIT_AUTH_METHOD", "https")

	// When
	vars, _ := GetEnv()

	// Then
	assert.Equal(t, "argo", vars.ArgoNamespace)
	assert.Equal(t, "cello.yaml", vars.ConfigFilePath)
	assert.Equal(t, 8443, vars.Port)
}

func TestValidations(t *testing.T) {
	// Given
	reset()
	os.Setenv(appPrefix+"_ADMIN_SECRET", "PW1234")
	os.Setenv("VAULT_ROLE", "vaultRole")
	os.Setenv("VAULT_SECRET", testSecret)
	os.Setenv("VAULT_ADDR", "1.2.3.4")
	os.Setenv("ARGO_ADDR", "2.3.4.5")
	os.Setenv(appPrefix+"_GIT_AUTH_METHOD", "https")

	// When
	_, err := GetEnv()

	// Then
	assert.Error(t, err)
}

func TestRequiredVars(t *testing.T) {
	// Given
	reset()
	os.Setenv("VAULT_ROLE", "vaultRole")
	os.Setenv("VAULT_SECRET", testSecret)
	os.Setenv("VAULT_ADDR", "1.2.3.4")
	os.Setenv("ARGO_ADDR", "2.3.4.5")
	os.Setenv(appPrefix+"_NAMESPACE", "argo-ns")
	os.Setenv(appPrefix+"_CONFIG", "/app/test/config/path")
	os.Setenv(appPrefix+"_GIT_AUTH_METHOD", "https")
	os.Setenv(appPrefix+"_LOG_LEVEL", "DEBUG")
	os.Setenv(appPrefix+"_PORT", "1234")

	// When
	_, err := GetEnv()

	// Then
	assert.Error(t, err)
}

func TestMigrateLegacyPrefix(t *testing.T) {
	reset()
	// Required vars with legacy prefixes
	setEnvVars(prefixedEnvVars, legacyAppPrefix)
	setEnvVars(nonPrefixedEnvVars, "")

	vars, err := GetEnv()
	assert.Equal(t, 0, vars.Port) // unsuccessfully set
	assert.Error(t, err)

	reset() // because this is forced to a singleton
	setEnvVars(nonPrefixedEnvVars, "")
	migrateLegacyPrefix()

	vars, err = GetEnv()
	assert.NoError(t, err)
	assert.Equal(t, 1234, vars.Port)

}
