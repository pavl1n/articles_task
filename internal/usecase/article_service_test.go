package usecase

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"articles/internal/domain"
)

type stubArticleRepo struct {
	saveFn    func(ctx context.Context, article domain.Article) (domain.Article, error)
	getByIDFn func(ctx context.Context, id int64) (domain.Article, error)
}

func (s *stubArticleRepo) Save(ctx context.Context, article domain.Article) (domain.Article, error) {
	return s.saveFn(ctx, article)
}

func (s *stubArticleRepo) GetByID(ctx context.Context, id int64) (domain.Article, error) {
	return s.getByIDFn(ctx, id)
}

func TestArticleService_CreateArticle_Success(t *testing.T) {
	want := domain.Article{ID: 1, Title: "Hello", CreatedAt: time.Unix(0, 0)}
	repo := &stubArticleRepo{
		saveFn: func(_ context.Context, article domain.Article) (domain.Article, error) {
			if article.Title != "Hello" {
				t.Fatalf("expected title to be trimmed to %q, got %q", "Hello", article.Title)
			}
			return want, nil
		},
		getByIDFn: nil,
	}
	svc := NewArticleService(repo)

	got, err := svc.CreateArticle(context.Background(), "  Hello  ")
	if err != nil {
		t.Fatalf("CreateArticle returned error: %v", err)
	}
	if got != want {
		t.Fatalf("expected %+v, got %+v", want, got)
	}
}

func TestArticleService_CreateArticle_InvalidTitle(t *testing.T) {
	svc := NewArticleService(&stubArticleRepo{})

	_, err := svc.CreateArticle(context.Background(), "   ")
	if !errors.Is(err, domain.ErrInvalidTitle) {
		t.Fatalf("expected ErrInvalidTitle, got %v", err)
	}
}

func TestArticleService_CreateArticle_TitleTooLong(t *testing.T) {
	svc := NewArticleService(&stubArticleRepo{})

	_, err := svc.CreateArticle(context.Background(), strings.Repeat("a", domain.MaxTitleLength+1))
	if !errors.Is(err, domain.ErrTitleTooLong) {
		t.Fatalf("expected ErrTitleTooLong, got %v", err)
	}
}

func TestArticleService_CreateArticle_RepoError(t *testing.T) {
	repoErr := errors.New("db error")
	repo := &stubArticleRepo{
		saveFn: func(_ context.Context, _ domain.Article) (domain.Article, error) {
			return domain.Article{}, repoErr
		},
		getByIDFn: nil,
	}
	svc := NewArticleService(repo)

	_, err := svc.CreateArticle(context.Background(), "Hello")
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error to propagate, got %v", err)
	}
}

func TestArticleService_GetArticle_Success(t *testing.T) {
	want := domain.Article{ID: 42, Title: "Hello", CreatedAt: time.Unix(0, 0)}
	repo := &stubArticleRepo{
		saveFn: nil,
		getByIDFn: func(_ context.Context, id int64) (domain.Article, error) {
			if id != want.ID {
				t.Fatalf("expected id %d, got %d", want.ID, id)
			}
			return want, nil
		},
	}
	svc := NewArticleService(repo)

	got, err := svc.GetArticle(context.Background(), want.ID)
	if err != nil {
		t.Fatalf("GetArticle returned error: %v", err)
	}
	if got != want {
		t.Fatalf("expected %+v, got %+v", want, got)
	}
}

func TestArticleService_GetArticle_InvalidID(t *testing.T) {
	svc := NewArticleService(&stubArticleRepo{})

	_, err := svc.GetArticle(context.Background(), 0)
	if !errors.Is(err, domain.ErrInvalidID) {
		t.Fatalf("expected ErrInvalidID, got %v", err)
	}
}

func TestArticleService_GetArticle_RepoError(t *testing.T) {
	repoErr := errors.New("db error")
	repo := &stubArticleRepo{
		saveFn: nil,
		getByIDFn: func(_ context.Context, _ int64) (domain.Article, error) {
			return domain.Article{}, repoErr
		},
	}
	svc := NewArticleService(repo)

	_, err := svc.GetArticle(context.Background(), 1)
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error to propagate, got %v", err)
	}
}
