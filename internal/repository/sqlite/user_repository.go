package sqlite

import (
	"context"
	"database/sql"
	"errors"

	"github.com/siddhantmadhur/pequod/internal/domain"
	"github.com/siddhantmadhur/pequod/internal/repository"
)

type userRepository struct {
	queries *Queries
}

func NewUserRepository(db DBTX) repository.UserRepository {
	return &userRepository{
		queries: New(db),
	}
}

func (r *userRepository) Create(ctx context.Context, username string, password []byte, permLevel int64) error {
	return r.queries.CreateProfile(ctx, CreateProfileParams{
		Username: username,
		Password: password,
		Type:     permLevel,
	})
}

func (r *userRepository) Update(ctx context.Context, id int64, username string, password []byte) error {
	return r.queries.UpdateUser(ctx, UpdateUserParams{
		ID:       id,
		Username: username,
		Password: password,
	})
}

func (r *userRepository) GetByID(ctx context.Context, id int64) (*domain.Profile, error) {
	p, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &domain.Profile{
		ID:       p.ID,
		Username: p.Username,
		Password: p.Password,
		Type:     p.Type,
	}, nil
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*domain.Profile, error) {
	p, err := r.queries.GetUserFromUsername(ctx, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &domain.Profile{
		ID:       p.ID,
		Username: p.Username,
		Password: p.Password,
		Type:     p.Type,
	}, nil
}

func (r *userRepository) GetAdmin(ctx context.Context) (*domain.Profile, error) {
	p, err := r.queries.GetAdminUser(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &domain.Profile{
		ID:       p.ID,
		Username: p.Username,
		Password: p.Password,
		Type:     p.Type,
	}, nil
}

func (r *userRepository) GetAll(ctx context.Context) ([]domain.Profile, error) {
	profiles, err := r.queries.GetProfiles(ctx)
	if err != nil {
		return nil, err
	}
	var res []domain.Profile
	for _, p := range profiles {
		res = append(res, domain.Profile{
			ID:       p.ID,
			Username: p.Username,
			Password: p.Password,
			Type:     p.Type,
		})
	}
	return res, nil
}

func (r *userRepository) Count(ctx context.Context) (int64, error) {
	return r.queries.IsFinishedSetup(ctx)
}
