package env

import (
	"os"
)

func Getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

// TODO exit if this is not set to something long
func AdminSecret() string {
	return os.Getenv("ARGO_CLOUDOPS_ADMIN_SECRET")
}

func ArgoNamespace() string {
	return Getenv("ARGO_CLOUDOPS_WORKFLOW_EXECUTION_NAMESPACE", "argo")
}

func ConfigFilePath() string {
	return Getenv("ARGO_CLOUDOPS_CONFIG", "argo-cloudops.yaml")
}
