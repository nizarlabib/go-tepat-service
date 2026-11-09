package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequestIDMiddleware(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	h := RequestID(next)

	// Test with no request ID header
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	requestID := w.Header().Get("X-Request-ID")
	if requestID == "" {
		t.Error("expected X-Request-ID header to be set")
	}

	// Test with existing request ID header
	req2 := httptest.NewRequest("GET", "/", nil)
	req2.Header.Set("X-Request-ID", "custom-request-id")
	w2 := httptest.NewRecorder()
	h.ServeHTTP(w2, req2)

	requestID2 := w2.Header().Get("X-Request-ID")
	if requestID2 != "custom-request-id" {
		t.Errorf("expected X-Request-ID 'custom-request-id', got '%s'", requestID2)
	}
}

func TestBodyLimitMiddleware(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	h := BodyLimit(10)(next)

	// Small body should pass
	req := httptest.NewRequest("POST", "/", strings.NewReader("small body"))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200 for small body, got %d", w.Code)
	}
}

func TestContentTypeMiddleware(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	h := ContentType("application/json")(next)

	// GET request without Content-Type should pass (not in allowed methods)
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200 for GET without Content-Type, got %d", w.Code)
	}

	// POST with correct Content-Type should pass
	req2 := httptest.NewRequest("POST", "/", nil)
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	h.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Errorf("expected status 200 for POST with application/json, got %d", w2.Code)
	}

	// POST without Content-Type should fail
	req3 := httptest.NewRequest("POST", "/", nil)
	w3 := httptest.NewRecorder()
	h.ServeHTTP(w3, req3)
	if w3.Code != http.StatusUnsupportedMediaType {
		t.Errorf("expected status 415 for POST without Content-Type, got %d", w3.Code)
	}

	// POST with wrong Content-Type should fail
	req4 := httptest.NewRequest("POST", "/", nil)
	req4.Header.Set("Content-Type", "text/plain")
	w4 := httptest.NewRecorder()
	h.ServeHTTP(w4, req4)
	if w4.Code != http.StatusUnsupportedMediaType {
		t.Errorf("expected status 415 for POST with text/plain, got %d", w4.Code)
	}
}

func TestCompressionMiddleware(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"hello":"world"}`))
	})

	h := Compression(next)

	// Request without Accept-Encoding: gzip should pass through
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if w.Header().Get("Content-Encoding") != "" {
		t.Error("expected no Content-Encoding when client doesn't accept gzip")
	}

	// Request with Accept-Encoding: gzip should get compressed response
	req2 := httptest.NewRequest("GET", "/", nil)
	req2.Header.Set("Accept-Encoding", "gzip")
	w2 := httptest.NewRecorder()
	h.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w2.Code)
	}
	if w2.Header().Get("Content-Encoding") != "gzip" {
		t.Error("expected Content-Encoding: gzip")
	}
}
