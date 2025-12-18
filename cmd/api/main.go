package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"

	httpadapter "articles/internal/adapter/http"
	"articles/internal/adapter/storage/postgres"
	"articles/internal/server"
	"articles/internal/usecase"
)

type Config struct {
	HTTPPort    string
	DatabaseURL string
}

const (
	readHeaderTimeout = 5 * time.Second
	shutdownTimeout   = 5 * time.Second
)

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("invalid configuration: %v", err)
	}

	db, cleanup, err := openDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to build server: %v", err)
	}
	defer func() {
		if err := cleanup(); err != nil {
			log.Printf("failed to close database: %v", err)
		}
	}()

	articleRepo := postgres.NewArticleRepository(db)
	articleService := usecase.NewArticleService(articleRepo)
	articleHandler := httpadapter.NewArticleHandler(articleService)
	router := server.NewRouter(articleHandler, func(ctx context.Context) error {
		return db.WithContext(ctx).Exec("SELECT 1").Error
	})
	httpServer := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           router,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("HTTP server listening on :%s", cfg.HTTPPort)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	log.Printf("shutting down...")
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
}

func loadConfig() (Config, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}

	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8080"
	}

	return Config{HTTPPort: port, DatabaseURL: dbURL}, nil
}

func openDB(databaseURL string) (*gorm.DB, func() error, error) {
	db, err := gorm.Open(gormpostgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		return nil, nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, err
	}

	if err := sqlDB.Ping(); err != nil {
		_ = sqlDB.Close()
		return nil, nil, err
	}

	cleanup := func() error {
		return sqlDB.Close()
	}

	return db, cleanup, nil
}
