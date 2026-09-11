package sqlite

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/siddhantmadhur/pequod/internal/domain"
)

func setupTestDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:?_foreign_keys=on")
	if err != nil {
		t.Fatalf("failed to open test sqlite db: %v", err)
	}

	if _, err := db.Exec(ddl); err != nil {
		_ = db.Close()
		t.Fatalf("failed to apply schema: %v", err)
	}

	// Insert test profiles for foreign key constraints
	_, err = db.Exec(`
		INSERT INTO profiles (id, username, password, type) 
		VALUES (1, 'user1', X'68617368', 0), (2, 'user2', X'68617368', 1)
	`)
	if err != nil {
		_ = db.Close()
		t.Fatalf("failed to insert test profiles: %v", err)
	}

	cleanup := func() {
		_ = db.Close()
	}

	return db, cleanup
}

func TestSessionRepository_Lifecycle(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewSessionRepository(db)
	ctx := context.Background()

	now := time.Now().Truncate(time.Second)

	s1 := &domain.Session{
		ID:               "session-1",
		UserID:           1,
		AccessToken:      "access-1",
		RefreshToken:     "refresh-1",
		CreatedAt:        now,
		AccessExpiresAt:  now.Add(20 * time.Minute),
		RefreshExpiresAt: now.Add(300 * time.Hour),
		Device:           "mac",
		DeviceName:       "MacBook Pro",
		ClientName:       "Pequod Web",
		ClientVersion:    "1.0",
	}

	s2 := &domain.Session{
		ID:               "session-2",
		UserID:           1,
		AccessToken:      "access-2",
		RefreshToken:     "refresh-2",
		CreatedAt:        now,
		AccessExpiresAt:  now.Add(20 * time.Minute),
		RefreshExpiresAt: now.Add(-1 * time.Hour), // Expired
		Device:           "phone",
		DeviceName:       "iPhone",
		ClientName:       "Pequod Mobile",
		ClientVersion:    "1.0",
	}

	t.Run("create and retrieve session", func(t *testing.T) {
		err := repo.Create(ctx, s1)
		if err != nil {
			t.Fatalf("unexpected error creating session: %v", err)
		}

		retrieved, err := repo.GetByID(ctx, s1.ID)
		if err != nil {
			t.Fatalf("unexpected error fetching session: %v", err)
		}
		if retrieved.ID != s1.ID || retrieved.UserID != s1.UserID {
			t.Fatalf("retrieved session mismatch: got %+v, expected %+v", retrieved, s1)
		}
		if retrieved.ClientName != "Pequod Web" {
			t.Fatalf("expected ClientName 'Pequod Web', got %s", retrieved.ClientName)
		}
	})

	t.Run("delete session by ID", func(t *testing.T) {
		err := repo.Delete(ctx, s1.ID)
		if err != nil {
			t.Fatalf("unexpected error deleting session: %v", err)
		}

		_, err = repo.GetByID(ctx, s1.ID)
		if err != domain.ErrNotFound {
			t.Fatalf("expected ErrNotFound after deleting session, got %v", err)
		}
	})

	t.Run("delete sessions by user ID", func(t *testing.T) {
		// Re-create s1 and another session for user 1
		_ = repo.Create(ctx, s1)
		_ = repo.Create(ctx, s2)

		err := repo.DeleteByUserID(ctx, 1)
		if err != nil {
			t.Fatalf("unexpected error deleting by user ID: %v", err)
		}

		_, err = repo.GetByID(ctx, s1.ID)
		if err != domain.ErrNotFound {
			t.Fatalf("expected ErrNotFound for s1 after user deletion, got %v", err)
		}
		_, err = repo.GetByID(ctx, s2.ID)
		if err != domain.ErrNotFound {
			t.Fatalf("expected ErrNotFound for s2 after user deletion, got %v", err)
		}
	})

	t.Run("delete expired sessions", func(t *testing.T) {
		_ = repo.Create(ctx, s1) // not expired
		_ = repo.Create(ctx, s2) // expired 1 hour ago

		err := repo.DeleteExpired(ctx, now)
		if err != nil {
			t.Fatalf("unexpected error deleting expired sessions: %v", err)
		}

		// s1 (active) must still exist
		active, err := repo.GetByID(ctx, s1.ID)
		if err != nil || active == nil {
			t.Fatalf("expected active session s1 to remain, got err: %v", err)
		}

		// s2 (expired) must be deleted
		_, err = repo.GetByID(ctx, s2.ID)
		if err != domain.ErrNotFound {
			t.Fatalf("expected expired session s2 to be deleted, got %v", err)
		}
	})
}
