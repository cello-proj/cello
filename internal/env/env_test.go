package env

import (
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

const testSecret = "tha5hei2Hee5le8neezu"

func setup() {
	os.Clearenv()
	instance = EnvVars{}
	once = sync.Once{}
}

func TestGetEnv(t *testing.T) {
	// Given
	setup()
	os.Setenv("ARGO_CLOUDOPS_ADMIN_SECRET", testSecret)
	os.Setenv("VAULT_ROLE", "vaultRole")
	os.Setenv("VAULT_SECRET", testSecret)
	os.Setenv("VAULT_ADDR", "1.2.3.4")
	os.Setenv("ARGO_ADDR", "2.3.4.5")
	os.Setenv("ARGO_CLOUDOPS_WORKFLOW_EXECUTION_NAMESPACE", "argo-ns")
	os.Setenv("ARGO_CLOUDOPS_CONFIG", "/app/test/config/path")
	os.Setenv("SSH_PEM_FILE", "/app/test/ssh.pem")
	os.Setenv("ARGO_CLOUDOPS_LOG_LEVEL", "DEBUG")
	os.Setenv("ARGO_CLOUDOPS_PORT", "1234")

	// When
	var env EnvVars = GetEnv()

	// Then
	assert.Equal(t, env.AdminSecret, testSecret)
	assert.Equal(t, env.VaultRole, "vaultRole")
	assert.Equal(t, env.VaultSecret, testSecret)
	assert.Equal(t, env.VaultAddress, "1.2.3.4")
	assert.Equal(t, env.ArgoNamespace, "argo-ns")
	assert.Equal(t, env.ConfigFilePath, "/app/test/config/path")
	assert.Equal(t, env.SshPemFile, "/app/test/ssh.pem")
	assert.Equal(t, env.LogLevel, "DEBUG")
	assert.Equal(t, env.Port, int32(1234))
}

func TestDefaults(t *testing.T) {
	// Given
	setup()
	os.Setenv("ARGO_CLOUDOPS_ADMIN_SECRET", testSecret)
	os.Setenv("VAULT_ROLE", "vaultRole")
	os.Setenv("VAULT_SECRET", testSecret)
	os.Setenv("VAULT_ADDR", "1.2.3.4")
	os.Setenv("ARGO_ADDR", "2.3.4.5")
	os.Setenv("SSH_PEM_FILE", "/app/test/ssh.pem")

	// When
	var env EnvVars = GetEnv()

	// Then
	assert.Equal(t, env.ArgoNamespace, "argo")
	assert.Equal(t, env.ConfigFilePath, "argo-cloudops.yaml")
	assert.Equal(t, env.Port, int32(8443))
}

func TestValidations(t *testing.T) {
	// Given
	setup()
	os.Setenv("ARGO_CLOUDOPS_ADMIN_SECRET", "PW1234")
	os.Setenv("VAULT_ROLE", "vaultRole")
	os.Setenv("VAULT_SECRET", testSecret)
	os.Setenv("VAULT_ADDR", "1.2.3.4")
	os.Setenv("ARGO_ADDR", "2.3.4.5")
	os.Setenv("SSH_PEM_FILE", "/app/test/ssh.pem")

	// When
	subject := func() { GetEnv() }

	// Then
	assert.Panics(t, subject, "The code did not panic")
}

func TestRequiredVars(t *testing.T) {
	// Given
	setup()
	os.Setenv("VAULT_ROLE", "vaultRole")
	os.Setenv("VAULT_SECRET", testSecret)
	os.Setenv("VAULT_ADDR", "1.2.3.4")
	os.Setenv("ARGO_ADDR", "2.3.4.5")
	os.Setenv("ARGO_CLOUDOPS_NAMESPACE", "argo-ns")
	os.Setenv("ARGO_CLOUDOPS_CONFIG", "/app/test/config/path")
	os.Setenv("SSH_PEM_FILE", "/app/test/ssh.pem")
	os.Setenv("ARGO_CLOUDOPS_LOG_LEVEL", "DEBUG")
	os.Setenv("ARGO_CLOUDOPS_PORT", "1234")
	// When
	subject := func() { GetEnv() }

	// Then
	assert.Panics(t, subject, "The code did not panic")
}
