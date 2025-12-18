package domain

import (
	"strings"
	"testing"
)

func TestNewArticle_TrimsAndCreates(t *testing.T) {
	article, err := NewArticle("  Hello  ")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if article.Title != "Hello" {
		t.Fatalf("expected title to be trimmed to %q, got %q", "Hello", article.Title)
	}
}

func TestNewArticle_InvalidTitle(t *testing.T) {
	_, err := NewArticle("   ")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err != ErrInvalidTitle {
		t.Fatalf("expected ErrInvalidTitle, got %v", err)
	}
}

func TestNewArticle_TitleTooLong(t *testing.T) {
	_, err := NewArticle(strings.Repeat("a", MaxTitleLength+1))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err != ErrTitleTooLong {
		t.Fatalf("expected ErrTitleTooLong, got %v", err)
	}
}
