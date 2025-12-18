package httpadapter

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"articles/internal/domain"
	"articles/internal/usecase"
)

type ArticleHandler struct {
	service *usecase.ArticleService
}

func NewArticleHandler(service *usecase.ArticleService) *ArticleHandler {
	return &ArticleHandler{service: service}
}

type createArticleRequest struct {
	Title string `json:"title"`
}

type articleResponse struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
}

func (h *ArticleHandler) CreateArticle(c *gin.Context) {
	var req createArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	article, err := h.service.CreateArticle(c.Request.Context(), req.Title)
	if err != nil {
		log.Printf("create article failed: %v", err)
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, toResponse(article))
}

func (h *ArticleHandler) GetArticle(c *gin.Context) {
	rawID := c.Param("id")
	id, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil {
		h.handleError(c, domain.ErrInvalidID)
		return
	}

	article, err := h.service.GetArticle(c.Request.Context(), id)
	if err != nil {
		log.Printf("get article failed: %v", err)
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toResponse(article))
}

func (h *ArticleHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidTitle),
		errors.Is(err, domain.ErrInvalidID),
		errors.Is(err, domain.ErrTitleTooLong):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrArticleNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}

func toResponse(article domain.Article) articleResponse {
	return articleResponse{
		ID:        article.ID,
		Title:     article.Title,
		CreatedAt: article.CreatedAt,
	}
}
