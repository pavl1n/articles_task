package domain

import (
	"strings"
	"time"
)

const MaxTitleLength = 140

type Article struct {
	ID        int64
	Title     string
	CreatedAt time.Time
}

func NewArticle(title string) (Article, error) {
	normalized := strings.TrimSpace(title)
	if normalized == "" {
		return Article{}, ErrInvalidTitle
	}
	if len([]rune(normalized)) > MaxTitleLength {
		return Article{}, ErrTitleTooLong
	}

	return Article{Title: normalized}, nil
}
