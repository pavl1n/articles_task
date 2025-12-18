package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"articles/internal/domain"
)

type ArticleRepository struct {
	db *gorm.DB
}

const queryTimeout = 3 * time.Second

func NewArticleRepository(db *gorm.DB) *ArticleRepository {
	return &ArticleRepository{db: db}
}

func (r *ArticleRepository) Save(ctx context.Context, article domain.Article) (domain.Article, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	model := articleModel{Title: article.Title}

	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return domain.Article{}, fmt.Errorf("create article: %w", err)
	}

	return domain.Article{
		ID:        model.ID,
		Title:     model.Title,
		CreatedAt: model.CreatedAt,
	}, nil
}

func (r *ArticleRepository) GetByID(ctx context.Context, id int64) (domain.Article, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	var model articleModel
	err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return domain.Article{}, domain.ErrArticleNotFound
	case err != nil:
		return domain.Article{}, fmt.Errorf("get article by id %d: %w", id, err)
	}

	return domain.Article{
		ID:        model.ID,
		Title:     model.Title,
		CreatedAt: model.CreatedAt,
	}, nil
}

type articleModel struct {
	ID        int64     `gorm:"column:id;primaryKey"`
	Title     string    `gorm:"column:title"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (articleModel) TableName() string { return "articles" }
