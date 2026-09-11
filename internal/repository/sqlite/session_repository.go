package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/siddhantmadhur/pequod/internal/domain"
	"github.com/siddhantmadhur/pequod/internal/repository"
)

type sessionRepository struct {
	queries *Queries
}

func NewSessionRepository(db DBTX) repository.SessionRepository {
	return &sessionRepository{
		queries: New(db),
	}
}

func (r *sessionRepository) Create(ctx context.Context, session *domain.Session) error {
	s, err := r.queries.CreateSession(ctx, CreateSessionParams{
		ID:               session.ID,
		UserID:           session.UserID,
		CreatedAt:        session.CreatedAt,
		AccessToken:      session.AccessToken,
		RefreshToken:     session.RefreshToken,
		RefreshExpiresAt: session.RefreshExpiresAt,
		AccessExpiresAt:  session.AccessExpiresAt,
		Device:           session.Device,
		DeviceName:       session.DeviceName,
		ClientName:       session.ClientName,
		ClientVersion:    session.ClientVersion,
	})
	if err != nil {
		return err
	}
	session.ID = s.ID
	return nil
}

func (r *sessionRepository) GetByID(ctx context.Context, id string) (*domain.Session, error) {
	s, err := r.queries.GetSessionByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &domain.Session{
		ID:               s.ID,
		UserID:           s.UserID,
		AccessToken:      s.AccessToken,
		RefreshToken:     s.RefreshToken,
		CreatedAt:        s.CreatedAt,
		AccessExpiresAt:  s.AccessExpiresAt,
		RefreshExpiresAt: s.RefreshExpiresAt,
		Device:           s.Device,
		DeviceName:       s.DeviceName,
		ClientName:       s.ClientName,
		ClientVersion:    s.ClientVersion,
	}, nil
}

func (r *sessionRepository) Delete(ctx context.Context, id string) error {
	return r.queries.DeleteSession(ctx, id)
}

func (r *sessionRepository) DeleteByUserID(ctx context.Context, userID int64) error {
	return r.queries.DeleteSessionsByUserID(ctx, userID)
}

func (r *sessionRepository) DeleteExpired(ctx context.Context, before time.Time) error {
	return r.queries.DeleteExpiredSessions(ctx, before)
}

