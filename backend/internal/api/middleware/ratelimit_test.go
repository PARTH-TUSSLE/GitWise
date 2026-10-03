package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gitwise/backend/internal/api/middleware"
)

func TestRateLimiter_AllowsWithinLimit(t *testing.T) {
	rl := middleware.NewRateLimiter(5, 5, time.Minute)
	defer rl.Close()

	handler := rl.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("request %d expected status 200, got %d", i+1, rec.Code)
		}
		if rec.Header().Get("X-RateLimit-Limit") != "5" {
			t.Errorf("expected limit header 5, got %s", rec.Header().Get("X-RateLimit-Limit"))
		}
	}
}

func TestRateLimiter_RejectsWhenLimitExceeded(t *testing.T) {
	rl := middleware.NewRateLimiter(2, 2, time.Minute)
	defer rl.Close()

	handler := rl.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req1 := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	req1.RemoteAddr = "10.0.0.1:54321"
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusOK {
		t.Fatalf("req 1: expected 200, got %d", rec1.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	req2.RemoteAddr = "10.0.0.1:54321"
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("req 2: expected 200, got %d", rec2.Code)
	}

	// 3rd request should exceed limit
	req3 := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	req3.RemoteAddr = "10.0.0.1:54321"
	rec3 := httptest.NewRecorder()
	handler.ServeHTTP(rec3, req3)

	if rec3.Code != http.StatusTooManyRequests {
		t.Fatalf("req 3: expected 429 Too Many Requests, got %d", rec3.Code)
	}
	if rec3.Header().Get("Retry-After") == "" {
		t.Errorf("expected Retry-After header on 429 response")
	}

	// Different IP should still be allowed
	reqOther := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	reqOther.RemoteAddr = "10.0.0.2:54321"
	recOther := httptest.NewRecorder()
	handler.ServeHTTP(recOther, reqOther)
	if recOther.Code != http.StatusOK {
		t.Errorf("expected different IP to succeed with 200, got %d", recOther.Code)
	}
}

func TestRateLimiter_BypassesHealthz(t *testing.T) {
	rl := middleware.NewRateLimiter(1, 1, time.Minute)
	defer rl.Close()

	handler := rl.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Exhaust the 1 token
	req1 := httptest.NewRequest(http.MethodGet, "/api/v1/something", nil)
	req1.RemoteAddr = "127.0.0.1:9999"
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec1.Code)
	}

	// /healthz should bypass and return 200
	reqHealth := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	reqHealth.RemoteAddr = "127.0.0.1:9999"
	recHealth := httptest.NewRecorder()
	handler.ServeHTTP(recHealth, reqHealth)

	if recHealth.Code != http.StatusOK {
		t.Fatalf("expected 200 on /healthz even when limit exhausted, got %d", recHealth.Code)
	}
}
