package sqlite

import (
	"context"
	"database/sql"
	"errors"

	"github.com/siddhantmadhur/pequod/internal/domain"
	"github.com/siddhantmadhur/pequod/internal/repository"
)

type contentRepository struct {
	queries *Queries
}

func NewContentRepository(db DBTX) repository.ContentRepository {
	return &contentRepository{
		queries: New(db),
	}
}

func (r *contentRepository) Create(ctx context.Context, params repository.CreateContentParams) (*domain.ContentFile, error) {
	c, err := r.queries.AddNewContentFile(ctx, AddNewContentFileParams{
		MediaLibraryID:     params.MediaLibraryID,
		CreatedAt:          params.CreatedAt,
		FilePath:           params.FilePath,
		Name:               params.Name,
		MediaTitle:         params.MediaTitle,
		Description:        params.Description,
		CoverUrl:           params.CoverUrl,
		ParentID:           params.ParentID,
		ExternalProvider:   params.ExternalProvider,
		ExternalProviderID: params.ExternalProviderID,
		MediaType:          params.MediaType,
		Classifier:         params.Classifier,
	})
	if err != nil {
		return nil, err
	}
	return mapContentLibrary(&c), nil
}

func (r *contentRepository) GetByID(ctx context.Context, id int64) (*domain.ContentFile, error) {
	c, err := r.queries.GetContentByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return mapContentLibrary(&c), nil
}

func (r *contentRepository) GetByPath(ctx context.Context, path string) (*domain.ContentFile, error) {
	c, err := r.queries.GetContentFromPath(ctx, path)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return mapContentLibrary(&c), nil
}

func (r *contentRepository) GetByExternalID(ctx context.Context, externalID int64) (*domain.ContentFile, error) {
	c, err := r.queries.GetContentFromExternalId(ctx, sql.NullInt64{Int64: externalID, Valid: true})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return mapContentLibrary(&c), nil
}

func (r *contentRepository) GetByParentID(ctx context.Context, parentID int64) ([]domain.ContentFile, error) {
	items, err := r.queries.GetContentFromParentId(ctx, sql.NullInt64{Int64: parentID, Valid: true})
	if err != nil {
		return nil, err
	}
	var res []domain.ContentFile
	for _, c := range items {
		res = append(res, *mapContentLibrary(&c))
	}
	return res, nil
}

func (r *contentRepository) GetAllByLibraryID(ctx context.Context, libraryID int64) ([]domain.ContentFile, error) {
	items, err := r.queries.GetAllContentFiles(ctx, libraryID)
	if err != nil {
		return nil, err
	}
	var res []domain.ContentFile
	for _, c := range items {
		res = append(res, *mapContentLibrary(&c))
	}
	return res, nil
}

func (r *contentRepository) GetAllShows(ctx context.Context, libraryID int64, mediaType string) ([]domain.ContentFile, error) {
	items, err := r.queries.GetAllShows(ctx, GetAllShowsParams{
		MediaLibraryID: libraryID,
		MediaType:      mediaType,
	})
	if err != nil {
		return nil, err
	}
	var res []domain.ContentFile
	for _, c := range items {
		res = append(res, *mapContentLibrary(&c))
	}
	return res, nil
}

func mapContentLibrary(c *ContentLibrary) *domain.ContentFile {
	return &domain.ContentFile{
		ID:                 c.ID,
		MediaLibraryID:     c.MediaLibraryID,
		CreatedAt:          c.CreatedAt,
		FilePath:           c.FilePath,
		Name:               c.Name,
		MediaTitle:         c.MediaTitle,
		Description:        c.Description,
		CoverUrl:           c.CoverUrl,
		ParentID:           c.ParentID,
		Classifier:         c.Classifier,
		MediaType:          c.MediaType,
		ExternalProvider:   c.ExternalProvider,
		ExternalProviderID: c.ExternalProviderID,
	}
}
