package main

import (
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func dummyHandler(w http.ResponseWriter, r *http.Request) {}

func newRequest(method, url string) *http.Request {
	req := httptest.NewRequest(method, url, nil)
	return req
}

func TestCommonMiddleware(t *testing.T) {
	tests := []struct {
		input []string
		want  []string
	}{
		{input: []string{"X-B3-TraceId", "traceparent"}, want: []string{"X-B3-TraceId", "traceparent", txIDHeader}},
		{input: []string(nil), want: []string{txIDHeader}},
	}

	for _, tc := range tests {
		// When
		router := mux.NewRouter()
		router.HandleFunc("/", dummyHandler).Methods("GET")

		middleware := traceIDsMiddleware(tc.input)
		router.Use(middleware)

		req := newRequest("GET", "/")
		res := httptest.NewRecorder()
		router.ServeHTTP(res, req)

		// Then
		for _, header := range tc.want {
			assert.NotNil(t, res.Header().Get(header))
		}
	}
}

func TestGetTraceIDHeaders(t *testing.T) {
	tests := []struct {
		input []string
		want  []string
	}{
		{input: []string{"traceparent"}, want: []string{"traceparent", "X-B3-TraceId", txIDHeader}},
		{input: []string(nil), want: []string{b3TraceIDHeader, txIDHeader}},
	}

	for _, tc := range tests {
		// When
		traceIDHeaders := getTraceIDHeaders(tc.input)

		// Then
		assert.Equal(t, tc.want, traceIDHeaders)
	}
}

func TestFindExistingTraceId(t *testing.T) {
	traceIDs := map[string]string{b3TraceIDHeader: "trace_id_1", txIDHeader: "trace_id_2"}

	tests := []struct {
		input []string
		want  []string
	}{
		{input: []string{b3TraceIDHeader, txIDHeader}, want: []string{"trace_id_1"}},
		{input: []string{"traceparent"}, want: []string{""}},
	}

	for _, tc := range tests {
		// When

		req := newRequest("GET", "/")
		for _, header := range tc.input {
			req.Header.Set(header, traceIDs[header])
		}

		traceID := findExistingTraceId(req, tc.input)

		// Then
		assert.Equal(t, tc.want[0], traceID)
	}
}
