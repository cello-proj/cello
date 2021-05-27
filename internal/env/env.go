package env

import (
	"log"
	"sync"

	"github.com/kelseyhightower/envconfig"
)

type EnvVars struct {
	AdminSecret    string `split_words:"true"`
	VaultRole      string `envconfig:"VAULT_ROLE" required:"true"`
	VaultSecret    string `envconfig:"VAULT_SECRET" required:"true"`
	VaultAddress   string `envconfig:"VAULT_ADDR" required:"true"`
	ArgoAddress    string `envconfig:"ARGO_ADDR" required:"true"`
	Namespace      string `default:"argo"`
	ConfigFilePath string `envconfig:"CONFIG" default:"argo-cloudops.yaml"`
	SshPemFile     string `envconfig:"SSH_PEM_FILE" required:"true"`
	LogLevel       string `split_words:"true"`
	Port           int32
}

var (
	instance *EnvVars
	once     sync.Once
)

func GetEnv() *EnvVars {
	once.Do(func() {
		err := envconfig.Process("ARGO_CLOUD_OPS", &instance)
		if err != nil {
			log.Fatal(err.Error())
		}
		instance.validate()
	})
	return instance
}

func (values *EnvVars) validate() {
	if len(values.AdminSecret) <= 16 {
		panic("Admin secret must be at least 16 characers long.")
	}
}
