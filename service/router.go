package main

import (
	"github.com/google/uuid"
	"net/http"

	"github.com/gorilla/mux"
)

const (
	txIDHeader      = "X-Trace-Id"
	b3TraceIDHeader = "X-B3-TraceId"
)

func setupRouter(h handler) *mux.Router {
	r := mux.NewRouter()
	r.Use(commonMiddleware)
	r.Use(traceIDsMiddleware(getTraceIDHeaders(h.env.TraceIDHeaders)))

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

// traceIDsMiddleware returns a middleware which allows to append trace ID headers to the request. By default, it will
// add a txIDHeader header to the request if it is not already set. It will also add any trace IDs specified in the
// environment variables
func traceIDsMiddleware(traceIDHeaders []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// If the request already has a trace ID header, we just reuse it for the following requests.
			// Now we only search for the first trace ID header found in the request headers.
			// We can specify the priority of the trace ID by the order of the trace ID headers in the environment variables
			traceId := findExistingTraceId(r, traceIDHeaders)
			if traceId == "" {
				// Now we just simply use a random UUID for the trace ID. We may adopt a more sophisticated approach
				// later with respect to known trace IDs following the W3 trace context Http Headers format
				// https://www.w3.org/TR/trace-context/#trace-context-http-headers-format
				traceId = uuid.NewString()
			}

			for _, header := range traceIDHeaders {
				if r.Header.Get(header) == "" {
					r.Header.Set(header, traceId)
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func getTraceIDHeaders(customTraceIDHeaders []string) []string {
	var traceHeaders []string
	traceHeaders = append(traceHeaders, customTraceIDHeaders...)

	// for backward compatibility, we also add the b3TraceIDHeader
	traceHeaders = append(traceHeaders, []string{b3TraceIDHeader, txIDHeader}...)

	return traceHeaders
}

// findExistingTraceId returns the first trace ID found in the request headers
func findExistingTraceId(r *http.Request, traceIDHeaders []string) string {
	traceId := ""

	for _, header := range traceIDHeaders {
		if r.Header.Get(header) != "" {
			traceId = r.Header.Get(header)
			break
		}
	}

	return traceId
}
