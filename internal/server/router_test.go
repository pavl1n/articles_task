package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	httpadapter "articles/internal/adapter/http"
	"articles/internal/usecase"
)

func TestHealthz_OK(t *testing.T) {
	router := NewRouter(httpadapter.NewArticleHandler(usecase.NewArticleService(nil)), func(context.Context) error { return nil })

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["status"] != "ok" {
		t.Fatalf("expected status \"ok\", got %q", body["status"])
	}
	if _, exists := body["db_error"]; exists {
		t.Fatalf("did not expect db_error key, got %v", body)
	}
}

func TestHealthz_DBError(t *testing.T) {
	dbErr := errors.New("db unavailable")
	router := NewRouter(httpadapter.NewArticleHandler(usecase.NewArticleService(nil)), func(context.Context) error {
		return dbErr
	})

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, rec.Code)
	}

	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["status"] != "degraded" {
		t.Fatalf("expected status \"degraded\", got %q", body["status"])
	}
	if body["db_error"] == "" {
		t.Fatalf("expected db_error in response, got %v", body)
	}
}
