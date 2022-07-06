package main

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

const (
	txIDHeader = "X-B3-TraceId"
)

func setupRouter(h handler) *mux.Router {
	r := mux.NewRouter()
	r.Use(commonMiddleware)
	r.Use(txIDMiddleware)

	r.HandleFunc("/workflows", h.createWorkflow).Methods(http.MethodPost)
	r.HandleFunc("/workflows/{workflowName}", h.getWorkflow).Methods(http.MethodGet)
	r.HandleFunc("/workflows/{workflowName}/logs", h.getWorkflowLogs).Methods(http.MethodGet)
	r.HandleFunc("/workflows/{workflowName}/logstream", h.getWorkflowLogStream).Methods(http.MethodGet)
	r.HandleFunc("/projects", h.createProject).Methods(http.MethodPost)
	r.HandleFunc("/projects/{projectName}", h.getProject).Methods(http.MethodGet)
	r.HandleFunc("/projects/{projectName}", h.deleteProject).Methods(http.MethodDelete)
	r.HandleFunc("/projects/{projectName}/targets", h.listTargets).Methods(http.MethodGet)
	r.HandleFunc("/projects/{projectName}/targets", h.createTarget).Methods(http.MethodPost)
	r.HandleFunc("/projects/{projectName}/targets/{targetName}", h.getTarget).Methods(http.MethodGet)
	r.HandleFunc("/projects/{projectName}/targets/{targetName}", h.deleteTarget).Methods(http.MethodDelete)
	r.HandleFunc("/projects/{projectName}/targets/{targetName}", h.updateTarget).Methods(http.MethodPatch)
	r.HandleFunc("/projects/{projectName}/targets/{targetName}/operations", h.createWorkflowFromGit).Methods(http.MethodPost)
	r.HandleFunc("/projects/{projectName}/targets/{targetName}/workflows", h.listWorkflows).Methods(http.MethodGet)
	r.HandleFunc("/projects/{projectName}/tokens", h.createToken).Methods(http.MethodPost)
	r.HandleFunc("/projects/{projectName}/tokens", h.listTokens).Methods(http.MethodGet)
	r.HandleFunc("/projects/{projectName}/tokens/{tokenID}", h.deleteToken).Methods(http.MethodDelete)
	r.HandleFunc("/health/full", h.healthCheck).Methods(http.MethodGet)
	return r
}

func commonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func txIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(txIDHeader) == "" {
			r.Header.Set(txIDHeader, uuid.NewString())
		}
		next.ServeHTTP(w, r)
	})
}
