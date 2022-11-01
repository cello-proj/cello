package env

import (
	"errors"
	"os"
	"strings"
	"sync"

	"github.com/kelseyhightower/envconfig"
)

func init() {
	migrateLegacyPrefix()
}

const legacyAppPrefix = "ARGO_CLOUDOPS"
const appPrefix = "CELLO"

type Vars struct {
	AdminSecret    string   `split_words:"true" required:"true"`
	VaultRole      string   `envconfig:"VAULT_ROLE" required:"true"`
	VaultSecret    string   `envconfig:"VAULT_SECRET" required:"true"`
	VaultAddress   string   `envconfig:"VAULT_ADDR" required:"true"`
	ArgoAddress    string   `envconfig:"ARGO_ADDR" required:"true"`
	ArgoNamespace  string   `envconfig:"WORKFLOW_EXECUTION_NAMESPACE" default:"argo"`
	ConfigFilePath string   `envconfig:"CONFIG" default:"cello.yaml"`
	SSHPEMFile     string   `envconfig:"SSH_PEM_FILE"`
	GitAuthMethod  string   `split_words:"true" required:"true"`
	GitHTTPSUser   string   `envconfig:"GIT_HTTPS_USER"`
	GitHTTPSPass   string   `envconfig:"GIT_HTTPS_PASS"`
	LogLevel       string   `split_words:"true"`
	Port           int      `default:"8443"`
	DBHost         string   `split_words:"true" required:"true"`
	DBUser         string   `split_words:"true" required:"true"`
	DBPassword     string   `split_words:"true" required:"true"`
	DBName         string   `split_words:"true" required:"true"`
	ImageURIs      []string `envconfig:"IMAGE_URIS"`
}

var (
	instance Vars
	once     sync.Once
	err      error
)

func GetEnv() (Vars, error) {
	once.Do(func() {
		err = envconfig.Process(appPrefix, &instance)
		if err != nil {
			return
		}
		err = instance.validate()
	})
	return instance, err
}

func (values Vars) validate() error {
	if len(values.AdminSecret) < 16 {
		return errors.New("admin secret must be at least 16 characers long")
	}
	return nil
}

func migrateLegacyPrefix() {
	for _, entry := range os.Environ() {
		if !strings.HasPrefix(entry, legacyAppPrefix) {
			continue
		}

		kv := strings.SplitN(entry, "=", 2)
		k, v := strings.TrimPrefix(kv[0], legacyAppPrefix), kv[1]
		os.Setenv(appPrefix+k, v)
	}
}
