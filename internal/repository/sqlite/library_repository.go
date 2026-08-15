package sqlite

import (
	"context"
	"database/sql"
	"errors"

	"github.com/siddhantmadhur/pequod/internal/domain"
	"github.com/siddhantmadhur/pequod/internal/repository"
)

type libraryRepository struct {
	queries *Queries
}

func NewLibraryRepository(db DBTX) repository.LibraryRepository {
	return &libraryRepository{
		queries: New(db),
	}
}

func (r *libraryRepository) Create(ctx context.Context, params repository.CreateLibraryParams) (*domain.MediaLibrary, error) {
	lib, err := r.queries.CreateMediaLibrary(ctx, CreateMediaLibraryParams{
		CreatedAt:   params.CreatedAt,
		Name:        params.Name,
		Description: params.Description,
		DevicePath:  params.DevicePath,
		MediaType:   params.MediaType,
		OwnerID:     params.OwnerID,
	})
	if err != nil {
		return nil, err
	}
	return &domain.MediaLibrary{
		ID:          lib.ID,
		CreatedAt:   lib.CreatedAt,
		Name:        lib.Name,
		Description: lib.Description,
		DevicePath:  lib.DevicePath,
		MediaType:   lib.MediaType,
		OwnerID:     lib.OwnerID,
	}, nil
}

func (r *libraryRepository) GetAll(ctx context.Context) ([]domain.MediaLibrary, error) {
	libs, err := r.queries.GetAllMediaLibraries(ctx)
	if err != nil {
		return nil, err
	}
	var res []domain.MediaLibrary
	for _, lib := range libs {
		res = append(res, domain.MediaLibrary{
			ID:          lib.ID,
			CreatedAt:   lib.CreatedAt,
			Name:        lib.Name,
			Description: lib.Description,
			DevicePath:  lib.DevicePath,
			MediaType:   lib.MediaType,
			OwnerID:     lib.OwnerID,
		})
	}
	return res, nil
}

func (r *libraryRepository) GetByID(ctx context.Context, id int64) (*domain.MediaLibrary, error) {
	lib, err := r.queries.GetMediaLibrary(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &domain.MediaLibrary{
		ID:          lib.ID,
		CreatedAt:   lib.CreatedAt,
		Name:        lib.Name,
		Description: lib.Description,
		DevicePath:  lib.DevicePath,
		MediaType:   lib.MediaType,
		OwnerID:     lib.OwnerID,
	}, nil
}
