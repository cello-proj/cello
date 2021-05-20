package main

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
)

func setupRouter(ctx context.Context, h handler) *mux.Router {
	r := mux.NewRouter()
	r.Use(contextMiddleware(ctx))
	r.Use(commonMiddleware)

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
	r.HandleFunc("/projects/{projectName}/targets/{targetName}/operations", h.createWorkflowFromGit).Methods(http.MethodPost)
	r.HandleFunc("/projects/{projectName}/targets/{targetName}/workflows", h.listWorkflows).Methods(http.MethodGet)
	r.HandleFunc("/health/full", h.healthCheck).Methods(http.MethodGet)
	return r
}

func commonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func contextMiddleware(ctx context.Context) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(ctx))
		}
		return http.HandlerFunc(fn)
	}
}
