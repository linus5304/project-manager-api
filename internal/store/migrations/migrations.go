package migrations

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed *.sql
var migrationsFS embed.FS

// Apply runs all *.up.sql files in lexical order
// For our testcontainers DB (fresh DB each time), this is enough
// Later we can add a schema_migrations table to track versions.

func Apply(ctx context.Context, pool *pgxpool.Pool) error {
	files, err := fs.Glob(migrationsFS, "*.up.sql")
	if err != nil {
		return fmt.Errorf("glob migrations: %w", err)
	}
	sort.Strings(files)

	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire conn: %w", err)
	}
	defer conn.Release()

	for _, name := range files {
		b, err := migrationsFS.ReadFile(name)
		if err != nil {
			return fmt.Errorf("read migration %q: %w", name, err)
		}
		sql := strings.TrimSpace(string(b))
		if sql == "" {
			continue
		}

		// Use Pg.Conn.Exec so multi-statement SQL files work reliably
		_, err = conn.Conn().PgConn().Exec(ctx, sql).ReadAll()
		if err != nil {
			return fmt.Errorf("exec migration %q: %w", name, err)
		}
	}

	return nil
}
