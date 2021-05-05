package main

import (
	"os"
)

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

// TODO exit if this is not set to somehting long
func adminSecret() string {
	return os.Getenv("ARGO_CLOUDOPS_ADMIN_SECRET")
}

func argoNamespace() string {
	return getenv("ARGO_CLOUDOPS_NAMESPACE", "argo")
}

func configFilePath() string {
	return getenv("ARGO_CLOUDOPS_CONFIG", "argo-cloudops.yaml")
}
