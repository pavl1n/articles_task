package domain

import "context"

type ArticleRepository interface {
	Save(ctx context.Context, article Article) (Article, error)
	GetByID(ctx context.Context, id int64) (Article, error)
}
