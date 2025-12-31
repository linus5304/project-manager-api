package store

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/linus5304/project-manager-api/internal/domain"
	"github.com/linus5304/project-manager-api/internal/store/sqlc"
)

type PostgresStore struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

var _ ProjectStore = (*PostgresStore)(nil)

func NewPostgresStore(ctx context.Context, dsn string) (*PostgresStore, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	return &PostgresStore{
		pool:    pool,
		queries: sqlc.New(pool),
	}, nil
}

func (s *PostgresStore) Close() {
	s.pool.Close()
}

func (s *PostgresStore) InsertProject(ctx context.Context, name string) (domain.Project, error) {
	p := domain.Project{
		ID:        uuid.New(),
		Name:      name,
		CreatedAt: time.Now().UTC(),
	}

	row, err := s.queries.InsertProject(ctx, sqlc.InsertProjectParams{
		ID:        p.ID,
		Name:      p.Name,
		CreatedAt: p.CreatedAt,
	})
	if err != nil {
		return domain.Project{}, err
	}
	return domain.Project{
		ID:        row.ID,
		Name:      row.Name,
		CreatedAt: row.CreatedAt,
	}, nil
}

func (s *PostgresStore) GetProject(ctx context.Context, id uuid.UUID) (domain.Project, error) {
	row, err := s.queries.GetProject(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Project{}, ErrNotFound
		}
		return domain.Project{}, err
	}
	return domain.Project{
		ID:        row.ID,
		Name:      row.Name,
		CreatedAt: row.CreatedAt,
	}, nil
}

func (s *PostgresStore) ListProjects(ctx context.Context) ([]domain.Project, error) {
	rows, err := s.queries.ListProjects(ctx)
	if err != nil {
		return nil, err
	}

	projects := make([]domain.Project, 0, len(rows))
	for _, row := range rows {
		projects = append(projects, domain.Project{
			ID:        row.ID,
			Name:      row.Name,
			CreatedAt: row.CreatedAt,
		})
	}
	return projects, nil
}

func (s *PostgresStore) InsertTask(ctx context.Context, projectID uuid.UUID, title, description string) (domain.Task, error) {
	t := domain.Task{
		ID:          uuid.New(),
		ProjectID:   projectID,
		Title:       title,
		Description: description,
		Status:      "todo",
		CreatedAt:   time.Now().UTC(),
	}

	row, err := s.queries.InsertTask(ctx, sqlc.InsertTaskParams{
		ID:          t.ID,
		ProjectID:   t.ProjectID,
		Title:       t.Title,
		Description: t.Description,
		Status:      t.Status,
		CreatedAt:   t.CreatedAt,
	})

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return domain.Task{}, ErrProjectNotFound
		}
		return domain.Task{}, err
	}

	return domain.Task{
		ID:          row.ID,
		ProjectID:   row.ProjectID,
		Title:       row.Title,
		Description: row.Description,
		Status:      row.Status,
		CreatedAt:   row.CreatedAt,
	}, nil
}

func (s *PostgresStore) ListTasks(ctx context.Context, projectID uuid.UUID) ([]domain.Task, error) {
	rows, err := s.queries.ListTasks(ctx, projectID)
	if err != nil {
		return nil, err
	}

	if len(rows) == 0 {
		_, err := s.GetProject(ctx, projectID)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				return nil, ErrProjectNotFound
			}
			return nil, err
		}
		return []domain.Task{}, nil
	}

	tasks := make([]domain.Task, 0, len(rows))
	for _, r := range rows {
		tasks = append(tasks, domain.Task{
			ID:          r.ID,
			ProjectID:   r.ProjectID,
			Title:       r.Title,
			Description: r.Description,
			Status:      r.Status,
			CreatedAt:   r.CreatedAt,
		})
	}
	return tasks, nil
}

func optText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func (s *PostgresStore) UpdateTask(ctx context.Context, projectID, taskID uuid.UUID, update TaskUpdate) (domain.Task, error) {
	row, err := s.queries.UpdateTask(ctx, sqlc.UpdateTaskParams{
		ProjectID:   projectID,
		ID:          taskID,
		Title:       optText(update.Title),
		Description: optText(update.Description),
		Status:      optText(update.Status),
	})

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// could be project not found or task not found
			_, perr := s.GetProject(ctx, projectID)
			if perr != nil {
				if errors.Is(perr, ErrNotFound) {
					return domain.Task{}, ErrProjectNotFound
				}
				return domain.Task{}, perr
			}
			return domain.Task{}, ErrTaskNotFound
		}
		return domain.Task{}, err
	}

	return domain.Task{
		ID:          row.ID,
		ProjectID:   row.ProjectID,
		Title:       row.Title,
		Description: row.Description,
		Status:      row.Status,
		CreatedAt:   row.CreatedAt,
	}, nil
}
