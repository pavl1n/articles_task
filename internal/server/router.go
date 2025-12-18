package server

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	httpadapter "articles/internal/adapter/http"
)

const healthCheckTimeout = time.Second

func NewRouter(articleHandler *httpadapter.ArticleHandler, healthCheck func(context.Context) error) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	if healthCheck == nil {
		healthCheck = func(context.Context) error { return nil }
	}

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery(), limitRequestBody(1<<20))

	router.POST("/article", articleHandler.CreateArticle)
	router.GET("/article/:id", articleHandler.GetArticle)
	router.GET("/healthz", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), healthCheckTimeout)
		defer cancel()

		if err := healthCheck(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "degraded", "db_error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	return router
}

func limitRequestBody(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		c.Next()
	}
}
