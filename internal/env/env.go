package env

import (
	"sync"

	"github.com/kelseyhightower/envconfig"
)

const appPrefix = "ARGO_CLOUDOPS"

type EnvVars struct {
	AdminSecret    string `split_words:"true" required:"true"`
	VaultRole      string `envconfig:"VAULT_ROLE" required:"true"`
	VaultSecret    string `envconfig:"VAULT_SECRET" required:"true"`
	VaultAddress   string `envconfig:"VAULT_ADDR" required:"true"`
	ArgoAddress    string `envconfig:"ARGO_ADDR" required:"true"`
	ArgoNamespace  string `envconfig:"WORKFLOW_EXECUTION_NAMESPACE" default:"argo"`
	ConfigFilePath string `envconfig:"CONFIG" default:"argo-cloudops.yaml"`
	SSHPEMFile     string `envconfig:"SSH_PEM_FILE" required:"true"`
	LogLevel       string `split_words:"true"`
	Port           int  `default:"8443"`
}

var (
	instance EnvVars
	once     sync.Once
)

func GetEnv() EnvVars {
	once.Do(func() {
		err := envconfig.Process(appPrefix, &instance)
		if err != nil {
			panic(err.Error())
		}
		instance.validate()
	})
	return instance
}

func (values EnvVars) validate() {
	if len(values.AdminSecret) < 16 {
		panic("Admin secret must be at least 16 characers long.")
	}
}
