package migrations

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/linus5304/project-manager-api/internal/store/pgtest"
)

func TestApply_CreateSchema(t *testing.T) {
	pg := pgtest.StartPostgres(t)
	t.Logf("started postgres container with connection string: %s", pg.ConnString)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)

	pool, err := pgxpool.New(ctx, pg.ConnString)
	if err != nil {
		t.Fatalf("connect pgxpool: %v", err)
	}
	t.Cleanup(pool.Close)

	if err := Apply(ctx, pool); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}

	// verify tables exists
	var projectReg, taskReg *string
	if err := pool.QueryRow(ctx, "select to_regclass('public.projects')").Scan(&projectReg); err != nil {
		t.Fatalf("query projects table: %v", err)
	}
	if projectReg == nil {
		t.Fatalf("projects table does not exist")
	}
	if err := pool.QueryRow(ctx, "select to_regclass('public.tasks')").Scan(&taskReg); err != nil {
		t.Fatalf("query tasks table: %v", err)
	}
	if taskReg == nil {
		t.Fatalf("tasks table does not exist")
	}
}
