package pgtest

import (
	"context"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

type Postgres struct {
	ConnString string
}

func StartPostgres(t *testing.T) Postgres {
	t.Helper()

	if testing.Short() {
		t.Skip("skipping postgres integration test in -short mode")
	}

	testcontainers.SkipIfProviderIsNotHealthy(t)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	t.Cleanup(cancel)

	ctr, err := postgres.Run(
		ctx,
		"postgres:16-alpine",
		postgres.WithDatabase(("pm_test")),
		postgres.WithUsername("pm"),
		postgres.WithPassword("pm"),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}
	testcontainers.CleanupContainer(t, ctr)
	connStr, err := ctr.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	return Postgres{ConnString: connStr}
}
