package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWithAuth(t *testing.T) {
	cases := []struct {
		name           string
		secret         string
		queryKey       string
		expectedStatus int
	}{
		{
			name:           "No secret set",
			secret:         "",
			queryKey:       "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "No secret set, key provided",
			secret:         "",
			queryKey:       "anykey",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Secret set, correct key",
			secret:         "mysecret",
			queryKey:       "mysecret",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Secret set, incorrect key",
			secret:         "mysecret",
			queryKey:       "wrong",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Secret set, no key",
			secret:         "mysecret",
			queryKey:       "",
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			// Create middleware with the secret
			middleware := withAuth(handler, tc.secret)

			req := httptest.NewRequest("GET", "/?key="+tc.queryKey, nil)
			rr := httptest.NewRecorder()

			middleware.ServeHTTP(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, rr.Code)
			}
		})
	}
}
