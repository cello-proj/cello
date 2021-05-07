package main

import (
	"context"
	"net/http"

	acolog "github.com/argoproj-labs/argo-cloudops/log"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func setupRouter(h handler, logLevel string) *mux.Router {
	r := mux.NewRouter()
	r.Use(commonMiddleware)
	r.Use(contextMiddleware)
	r.Use(loggingMiddleware(logLevel))

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

func contextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		txnID := uuid.NewString()
		ctx := context.WithValue(r.Context(), "txn_id", txnID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func loggingMiddleware(lvl string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			acolog.SetLevel(lvl)
			ctx = acolog.AddFields(ctx, zap.String("txn_id", ctx.Value("txn_id").(string)))
			defer acolog.Sync(ctx)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
		return http.HandlerFunc(fn)
	}
}
