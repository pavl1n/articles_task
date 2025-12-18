package postgres

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"articles/internal/domain"
)

func setupTestDB(t *testing.T) (*gorm.DB, func()) {
	t.Helper()

	ensureDocker(t)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	t.Cleanup(cancel)

	const (
		user     = "testuser"
		password = "testpass"
		dbname   = "testdb"
	)

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "postgres:16-alpine",
			ExposedPorts: []string{"5432/tcp"},
			Env: map[string]string{
				"POSTGRES_USER":     user,
				"POSTGRES_PASSWORD": password,
				"POSTGRES_DB":       dbname,
			},
			WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(90 * time.Second),
		},
		Started: true,
	})
	if err != nil {
		t.Fatalf("start postgres container: %v", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		container.Terminate(ctx)
		t.Fatalf("get container host: %v", err)
	}
	mappedPort, err := container.MappedPort(ctx, "5432")
	if err != nil {
		container.Terminate(ctx)
		t.Fatalf("get mapped port: %v", err)
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, mappedPort.Port(), dbname)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		container.Terminate(ctx)
		t.Fatalf("open gorm: %v", err)
	}

	if err := db.AutoMigrate(&articleModel{}); err != nil {
		container.Terminate(ctx)
		t.Fatalf("auto-migrate: %v", err)
	}

	cleanup := func() {
		container.Terminate(context.Background())
	}

	return db, cleanup
}

func ensureDocker(t *testing.T) {
	t.Helper()
	cmd := exec.Command("docker", "info")
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		t.Skipf("docker not available: %v", err)
	}
	if os.Getenv("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE") == "" {
		_ = os.Setenv("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE", "unix:///var/run/docker.sock")
	}
	testcontainers.Logger = log.New(io.Discard, "", 0)
}

func TestArticleRepository_Save(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	repo := NewArticleRepository(db)

	created, err := repo.Save(context.Background(), domain.Article{Title: "Hello"})
	if err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	if created.ID == 0 {
		t.Fatalf("expected ID to be set")
	}
	if created.Title != "Hello" {
		t.Fatalf("expected title to be %q, got %q", "Hello", created.Title)
	}
	if created.CreatedAt.IsZero() {
		t.Fatalf("expected CreatedAt to be set")
	}
}

func TestArticleRepository_GetByID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	repo := NewArticleRepository(db)

	saved, err := repo.Save(context.Background(), domain.Article{Title: "Hello"})
	if err != nil {
		t.Fatalf("seed Save returned error: %v", err)
	}

	got, err := repo.GetByID(context.Background(), saved.ID)
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}

	if got.ID != saved.ID {
		t.Fatalf("expected ID %d, got %d", saved.ID, got.ID)
	}
	if got.Title != saved.Title {
		t.Fatalf("expected title %q, got %q", saved.Title, got.Title)
	}
	if got.CreatedAt.IsZero() {
		t.Fatalf("expected CreatedAt to be set")
	}
}

func TestArticleRepository_GetByID_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	repo := NewArticleRepository(db)

	_, err := repo.GetByID(context.Background(), 404)
	if !errors.Is(err, domain.ErrArticleNotFound) {
		t.Fatalf("expected ErrArticleNotFound, got %v", err)
	}
}
