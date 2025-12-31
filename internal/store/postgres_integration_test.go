package store

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/linus5304/project-manager-api/internal/store/migrations"
	"github.com/linus5304/project-manager-api/internal/store/pgtest"
)

func newPGStore(t *testing.T) (context.Context, *PostgresStore) {
	t.Helper()

	pg := pgtest.StartPostgres(t)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)

	s, err := NewPostgresStore(ctx, pg.ConnString)
	if err != nil {
		t.Fatalf("failed to create PostgresStore: %v", err)
	}
	t.Cleanup(s.Close)

	if err := migrations.Apply(ctx, s.pool); err != nil {
		t.Fatalf("failed to apply migrations: %v", err)
	}

	return ctx, s
}

func TestPostgresStore_InsertGetProject(t *testing.T) {
	ctx, s := newPGStore(t)

	created, err := s.InsertProject(ctx, "Alpha")
	if err != nil {
		t.Fatalf("InsertProject failed: %v", err)
	}

	got, err := s.GetProject(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetProject failed: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("expeected id %s: got %s", created.ID, got.ID)
	}
	if got.Name != "Alpha" {
		t.Errorf("expected name 'Alpha': got %s", got.Name)
	}

	_, err = s.GetProject(ctx, uuid.New())
	if err == nil {
		t.Fatalf("expected ErrNotFound, got nil")
	}
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestPostgresStore_InsertTask_ProjectNotFound(t *testing.T) {
	ctx, s := newPGStore(t)

	_, err := s.InsertTask(ctx, uuid.New(), "t1", "desc")
	if err == nil {
		t.Fatal("expected ErrProjectNotFound, got nil")
	}
	if err != ErrProjectNotFound {
		t.Fatalf("expected ErrProjectNotFound; got %v", err)
	}
}

func TestPostgresStore_ListTasks_EmptyVsMissingProject(t *testing.T) {
	ctx, s := newPGStore(t)

	// create project with no tasks
	p, err := s.InsertProject(ctx, "Alpha")
	if err != nil {
		t.Fatalf("InsertProject: %v", err)
	}

	// Existing project, no tasks -> empty list, nil error
	tasks, err := s.ListTasks(ctx, p.ID)
	if err != nil {
		t.Fatalf("ListTasks existing project: %v", err)
	}
	if len(tasks) != 0 {
		t.Fatalf("expected 0 tasks; got %d", len(tasks))
	}

	// Missing project -> ErrProjectNotFound
	_, err = s.ListTasks(ctx, uuid.New())
	if err == nil {
		t.Fatalf("expected ErrProjectNotFound, got nil")
	}
	if err != ErrProjectNotFound {
		t.Fatalf("expected ErrProjectNotFound; got %v", err)
	}
}

func TestPostgresStore_UpdateTask_Partial(t *testing.T) {
	ctx, s := newPGStore(t)

	p, err := s.InsertProject(ctx, "Alpha")
	if err != nil {
		t.Fatalf("InsertProject: %v", err)
	}
	task, err := s.InsertTask(ctx, p.ID, "T1", "desc")
	if err != nil {
		t.Fatalf("InsertTask: %v", err)
	}

	newStatus := "doing"
	update := TaskUpdate{Status: &newStatus}

	updated, err := s.UpdateTask(ctx, p.ID, task.ID, update)
	if err != nil {
		t.Fatalf("UpdateTask: %v", err)
	}

	if updated.Title != "T1" {
		t.Fatalf("title changed unexpectedly: got %q", updated.Title)
	}
	if updated.Description != "desc" {
		t.Fatalf("description changed unexpectedly: got %q", updated.Description)
	}
	if updated.Status != "doing" {
		t.Fatalf("expected status %q; got %q", "doing", updated.Status)
	}
}

func TestPostgresStore_UpdateTask_AllowsEmptyDescription(t *testing.T) {
	ctx, s := newPGStore(t)

	p, err := s.InsertProject(ctx, "Alpha")
	if err != nil {
		t.Fatalf("InsertProject: %v", err)
	}

	task, err := s.InsertTask(ctx, p.ID, "T1", "desc")
	if err != nil {
		t.Fatalf("InsertTask: %v", err)
	}

	empty := ""
	update := TaskUpdate{Description: &empty}

	updated, err := s.UpdateTask(ctx, p.ID, task.ID, update)
	if err != nil {
		t.Fatalf("UpdateTask: %v", err)
	}
	if updated.Description != "" {
		t.Fatalf("expected empty description; got %q", updated.Description)
	}
}

func TestPostgresStore_UpdateTask_NotFoundMapping(t *testing.T) {
	ctx, s := newPGStore(t)

	// project missing
	_, err := s.UpdateTask(ctx, uuid.New(), uuid.New(), TaskUpdate{})
	if err != ErrProjectNotFound {
		t.Fatalf("expected ErrProjectNotFound; got %v", err)
	}

	// project exists, task missing
	p, err := s.InsertProject(ctx, "Alpha")
	if err != nil {
		t.Fatalf("InsertProject: %v", err)
	}

	_, err = s.UpdateTask(ctx, p.ID, uuid.New(), TaskUpdate{})
	if err != ErrTaskNotFound {
		t.Fatalf("expected ErrTaskNotFound; got %v", err)
	}
}
