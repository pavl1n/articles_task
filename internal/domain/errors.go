package domain

import "errors"

var (
	ErrArticleNotFound = errors.New("article not found")
	ErrInvalidID       = errors.New("id must be a positive integer")
	ErrInvalidTitle    = errors.New("title is required")
	ErrTitleTooLong    = errors.New("title must be at most 140 characters")
)
