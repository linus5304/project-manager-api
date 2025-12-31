package pgtest

import (
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestPostgresContainer_Smoke(t *testing.T) {
	pg := StartPostgres(t)
	t.Logf("started postgres container with connection string: %s", pg.ConnString)

	ctx := t.Context()
	pool, err := pgxpool.New(ctx, pg.ConnString)
	if err != nil {
		t.Fatalf("connect pgxpool: %v", err)
	}
	t.Cleanup(pool.Close)

	var n int
	if err := pool.QueryRow(ctx, "select 1").Scan(&n); err != nil {
		t.Fatalf("select 1: %v", err)
	}
	if n != 1 {
		t.Fatalf("expected 1, got %d", n)
	}
}
