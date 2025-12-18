package usecase

import (
	"context"

	"articles/internal/domain"
)

type ArticleService struct {
	repo domain.ArticleRepository
}

func NewArticleService(repo domain.ArticleRepository) *ArticleService {
	return &ArticleService{repo: repo}
}

func (s *ArticleService) CreateArticle(ctx context.Context, title string) (domain.Article, error) {
	article, err := domain.NewArticle(title)
	if err != nil {
		return domain.Article{}, err
	}

	created, err := s.repo.Save(ctx, article)
	if err != nil {
		return domain.Article{}, err
	}

	return created, nil
}

func (s *ArticleService) GetArticle(ctx context.Context, id int64) (domain.Article, error) {
	if id <= 0 {
		return domain.Article{}, domain.ErrInvalidID
	}

	return s.repo.GetByID(ctx, id)
}
