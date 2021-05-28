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
	os.Setenv("ARGO_CLOUDOPS_NAMESPACE", "argo-ns")
	os.Setenv("CONFIG", "/app/test/config/path")
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
	if env.Namespace != "argo-ns" {
		t.Errorf("expected %v, got %v", "argo-ns", env.Namespace)
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

// type EnvVars struct {
// 	AdminSecret    string `split_words:"true"`
// 	VaultRole      string `envconfig:"VAULT_ROLE" required:"true"`
// 	VaultSecret    string `envconfig:"VAULT_SECRET" required:"true"`
// 	VaultAddress   string `envconfig:"VAULT_ADDR" required:"true"`
// 	ArgoAddress    string `envconfig:"ARGO_ADDR" required:"true"`
// 	Namespace      string `default:"argo"`
// 	ConfigFilePath string `envconfig:"CONFIG" default:"argo-cloudops.yaml"`
// 	SshPemFile     string `envconfig:"SSH_PEM_FILE" required:"true"`
// 	LogLevel       string `split_words:"true"`
// 	Port           int32
// }

func TestDefaults(t *testing.T) {
	// TODO
}

func TestRequiredVars(t *testing.T) {
	// TODO
}

func assertPanic(t *testing.T, f func()) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	f()
}
