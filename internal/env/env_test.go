package env

import (
	"os"
	"testing"
)

const testSecret = "tha5hei2Hee5le8neezu"

func TestGetEnv(t *testing.T) {
	os.Clearenv()
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

	var env EnvVars = GetEnv()
	if env.AdminSecret != testSecret {
		t.Errorf("expected %v, got %v", testSecret, env.AdminSecret)
	}
	if env.VaultRole != "vaultRole" {
		t.Errorf("expected %v, got %v", "vaultRole", env.VaultRole)
	}
	if env.VaultSecret != testSecret {
		t.Errorf("expected %v, got %v", testSecret, env.VaultSecret)
	}
	if env.VaultAddress != "1.2.3.4" {
		t.Errorf("expected %v, got %v", "1.2.3.4", env.VaultAddress)
	}
	if env.ArgoAddress != "2.3.4.5" {
		t.Errorf("expected %v, got %v", "2.3.4.5", env.ArgoAddress)
	}
	if env.ArgoNamespace != "argo-ns" {
		t.Errorf("expected %v, got %v", "argo-ns", env.ArgoNamespace)
	}
	if env.ConfigFilePath != "/app/test/config/path" {
		t.Errorf("expected %v, got %v", "/app/test/config/path", env.ConfigFilePath)
	}
	if env.SshPemFile != "/app/test/ssh.pem" {
		t.Errorf("expected %v, got %v", "/app/test/ssh.pem", env.SshPemFile)
	}
	if env.LogLevel != "DEBUG" {
		t.Errorf("expected %v, got %v", "DEBUG", env.LogLevel)
	}
	if env.Port != 1234 {
		t.Errorf("expected %v, got %v", 1234, env.Port)
	}
}

func TestDefaults(t *testing.T) {
	os.Clearenv()
	os.Setenv("ARGO_CLOUDOPS_ADMIN_SECRET", testSecret)
	os.Setenv("VAULT_ROLE", "vaultRole")
	os.Setenv("VAULT_SECRET", testSecret)
	os.Setenv("VAULT_ADDR", "1.2.3.4")
	os.Setenv("ARGO_ADDR", "2.3.4.5")
	os.Setenv("SSH_PEM_FILE", "/app/test/ssh.pem")
	var env EnvVars = GetEnv()
	if env.ArgoNamespace != "argo" {
		t.Errorf("expected %v, got %v", "argo-ns", env.ArgoNamespace)
	}
	if env.ConfigFilePath != "argo-cloudops.yaml" {
		t.Errorf("expected %v, got %v", "/app/test/config/path", env.ConfigFilePath)
	}
	if env.Port != 8443 {
		t.Errorf("expected %v, got %v", 8443, env.Port)
	}
}

// func TestRequiredVars(t *testing.T) {
// 	os.Clearenv()
// 	//os.Setenv("ARGO_CLOUDOPS_ADMIN_SECRET", testSecret)
// 	os.Setenv("VAULT_ROLE", "vaultRole")
// 	os.Setenv("VAULT_SECRET", testSecret)
// 	os.Setenv("VAULT_ADDR", "1.2.3.4")
// 	os.Setenv("ARGO_ADDR", "2.3.4.5")
// 	os.Setenv("ARGO_CLOUDOPS_NAMESPACE", "argo-ns")
// 	os.Setenv("ARGO_CLOUDOPS_CONFIG", "/app/test/config/path")
// 	os.Setenv("SSH_PEM_FILE", "/app/test/ssh.pem")
// 	os.Setenv("ARGO_CLOUDOPS_LOG_LEVEL", "DEBUG")
// 	os.Setenv("ARGO_CLOUDOPS_PORT", "1234")
// 	//assertPanic(t, env.GetEnv())
// 	GetEnv()
// }

// func assertPanic(t *testing.T, f func()) {
// 	defer func() {
// 		if r := recover(); r == nil {
// 			t.Errorf("The code did not panic")
// 		}
// 	}()
// 	f()
// }
