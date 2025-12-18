package httpadapter

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"articles/internal/domain"
	"articles/internal/usecase"
)

type stubRepo struct {
	saveFn    func(ctx context.Context, article domain.Article) (domain.Article, error)
	getByIDFn func(ctx context.Context, id int64) (domain.Article, error)
}

func (s *stubRepo) Save(ctx context.Context, article domain.Article) (domain.Article, error) {
	return s.saveFn(ctx, article)
}

func (s *stubRepo) GetByID(ctx context.Context, id int64) (domain.Article, error) {
	return s.getByIDFn(ctx, id)
}

func setupRouter(t *testing.T, repo domain.ArticleRepository) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	service := usecase.NewArticleService(repo)
	handler := NewArticleHandler(service)

	router := gin.New()
	router.POST("/article", handler.CreateArticle)
	router.GET("/article/:id", handler.GetArticle)

	return router
}

func performRequest(r http.Handler, method, path string, body []byte) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestCreateArticle_Success(t *testing.T) {
	expected := domain.Article{ID: 1, Title: "Hello", CreatedAt: time.Unix(0, 0)}
	router := setupRouter(t, &stubRepo{
		saveFn: func(_ context.Context, article domain.Article) (domain.Article, error) {
			if article.Title != "Hello" {
				t.Fatalf("expected title %q, got %q", "Hello", article.Title)
			}
			return expected, nil
		},
		getByIDFn: nil,
	})

	rec := performRequest(router, http.MethodPost, "/article", []byte(`{"title":"Hello"}`))

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	var resp articleResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.ID != expected.ID || resp.Title != expected.Title || !resp.CreatedAt.Equal(expected.CreatedAt) {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestCreateArticle_BadJSON(t *testing.T) {
	router := setupRouter(t, &stubRepo{})

	rec := performRequest(router, http.MethodPost, "/article", []byte(`{"title":`))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("invalid request body")) {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func TestCreateArticle_InvalidTitle(t *testing.T) {
	router := setupRouter(t, &stubRepo{})

	rec := performRequest(router, http.MethodPost, "/article", []byte(`{"title":"   "}`))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(domain.ErrInvalidTitle.Error())) {
		t.Fatalf("expected error message, got %s", rec.Body.String())
	}
}

func TestCreateArticle_TitleTooLong(t *testing.T) {
	router := setupRouter(t, &stubRepo{})
	longTitle := strings.Repeat("a", domain.MaxTitleLength+1)

	rec := performRequest(router, http.MethodPost, "/article", []byte(`{"title":"`+longTitle+`"}`))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(domain.ErrTitleTooLong.Error())) {
		t.Fatalf("expected error message, got %s", rec.Body.String())
	}
}

func TestCreateArticle_RepoError(t *testing.T) {
	repoErr := errors.New("db error")
	router := setupRouter(t, &stubRepo{
		saveFn: func(_ context.Context, _ domain.Article) (domain.Article, error) {
			return domain.Article{}, repoErr
		},
		getByIDFn: nil,
	})

	rec := performRequest(router, http.MethodPost, "/article", []byte(`{"title":"Hello"}`))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("internal server error")) {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func TestGetArticle_Success(t *testing.T) {
	expected := domain.Article{ID: 2, Title: "Hello", CreatedAt: time.Unix(0, 0)}
	router := setupRouter(t, &stubRepo{
		saveFn: nil,
		getByIDFn: func(_ context.Context, id int64) (domain.Article, error) {
			if id != expected.ID {
				t.Fatalf("expected id %d, got %d", expected.ID, id)
			}
			return expected, nil
		},
	})

	rec := performRequest(router, http.MethodGet, "/article/2", nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp articleResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.ID != expected.ID || resp.Title != expected.Title || !resp.CreatedAt.Equal(expected.CreatedAt) {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestGetArticle_InvalidID(t *testing.T) {
	router := setupRouter(t, &stubRepo{})

	rec := performRequest(router, http.MethodGet, "/article/not-a-number", nil)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(domain.ErrInvalidID.Error())) {
		t.Fatalf("expected invalid id message, got %s", rec.Body.String())
	}
}

func TestGetArticle_NotFound(t *testing.T) {
	router := setupRouter(t, &stubRepo{
		saveFn: nil,
		getByIDFn: func(_ context.Context, _ int64) (domain.Article, error) {
			return domain.Article{}, domain.ErrArticleNotFound
		},
	})

	rec := performRequest(router, http.MethodGet, "/article/123", nil)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(domain.ErrArticleNotFound.Error())) {
		t.Fatalf("expected not found message, got %s", rec.Body.String())
	}
}

func TestGetArticle_RepoError(t *testing.T) {
	repoErr := errors.New("db error")
	router := setupRouter(t, &stubRepo{
		saveFn: nil,
		getByIDFn: func(_ context.Context, _ int64) (domain.Article, error) {
			return domain.Article{}, repoErr
		},
	})

	rec := performRequest(router, http.MethodGet, "/article/1", nil)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("internal server error")) {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}
