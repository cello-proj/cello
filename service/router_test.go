package main

import (
	"github.com/cello-proj/cello/service/test/testhelpers"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func dummyHandler(w http.ResponseWriter, r *http.Request) {}

func newRequest(method, url string) *http.Request {
	req, _ := http.NewRequest(method, url, nil)
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
		res := testhelpers.NewMockResponseWriter()
		router.ServeHTTP(res, req)

		// Then
		for _, header := range tc.want {
			assert.NotNil(t, res.Header().Get(header))
		}
	}
}
