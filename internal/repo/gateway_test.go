package repo_test

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/mondial7/smart-360/internal/db"
	"github.com/mondial7/smart-360/internal/models"
	"github.com/mondial7/smart-360/internal/repo"
)

// testPool is a package-shared pool against a throwaway Postgres container. It
// is nil when tests run under -short (or Docker is unavailable), in which case
// every gateway test skips.
var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	flag.Parse()
	if testing.Short() {
		os.Exit(m.Run())
	}

	ctx := context.Background()
	container, err := tcpostgres.Run(ctx, "postgres:16-alpine",
		tcpostgres.WithDatabase("smart360"),
		tcpostgres.WithUsername("smart360"),
		tcpostgres.WithPassword("smart360"),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("5432/tcp").WithStartupTimeout(90*time.Second)),
	)
	if err != nil {
		log.Printf("skipping gateway tests: could not start postgres container: %v", err)
		os.Exit(m.Run())
	}
	defer func() { _ = container.Terminate(ctx) }()

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatalf("connection string: %v", err)
	}
	pool, err := db.Connect(ctx, connStr)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer pool.Close()
	if err := db.Migrate(ctx, pool); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	testPool = pool

	os.Exit(m.Run())
}

// gateway returns fresh repositories and truncates all data first, or skips the
// test when no container is available.
func gateway(t *testing.T) repo.Repositories {
	t.Helper()
	if testPool == nil {
		t.Skip("no postgres container (short mode or Docker unavailable)")
	}
	_, err := testPool.Exec(context.Background(), `
		TRUNCATE sessions, moderation_logs, audit_logs, consolidations, submissions,
		         round_reviewers, feedback_rounds, team_members, teams, templates
		RESTART IDENTITY CASCADE;
		DELETE FROM users;`)
	if err != nil {
		t.Fatalf("reset db: %v", err)
	}
	return repo.NewPostgres(testPool)
}

// makeUser inserts a user with a unique email and returns it.
func makeUser(t *testing.T, r repo.Repositories, name string) string {
	t.Helper()
	u := &models.User{Email: fmt.Sprintf("%s-%d@example.com", name, time.Now().UnixNano()), Name: name}
	if err := r.Users.Create(context.Background(), u); err != nil {
		t.Fatalf("create user: %v", err)
	}
	return u.ID
}
