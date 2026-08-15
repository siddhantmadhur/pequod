package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/siddhantmadhur/pequod/internal/domain"
)

// UserRepository defines data access operations for user profiles.
type UserRepository interface {
	Create(ctx context.Context, username string, password []byte, permLevel int64) error
	Update(ctx context.Context, id int64, username string, password []byte) error
	GetByID(ctx context.Context, id int64) (*domain.Profile, error)
	GetByUsername(ctx context.Context, username string) (*domain.Profile, error)
	GetAdmin(ctx context.Context) (*domain.Profile, error)
	GetAll(ctx context.Context) ([]domain.Profile, error)
	Count(ctx context.Context) (int64, error)
}

// SessionRepository defines data access operations for user sessions.
type SessionRepository interface {
	Create(ctx context.Context, session *domain.Session) error
	GetByID(ctx context.Context, id string) (*domain.Session, error)
}

// CreateLibraryParams contains fields to create a media library record.
type CreateLibraryParams struct {
	Name        string
	Description string
	DevicePath  string
	MediaType   string
	OwnerID     int64
	CreatedAt   time.Time
}

// LibraryRepository defines data access operations for media libraries.
type LibraryRepository interface {
	Create(ctx context.Context, params CreateLibraryParams) (*domain.MediaLibrary, error)
	GetAll(ctx context.Context) ([]domain.MediaLibrary, error)
	GetByID(ctx context.Context, id int64) (*domain.MediaLibrary, error)
}

// CreateContentParams contains fields to create a content library record.
type CreateContentParams struct {
	MediaLibraryID     int64
	CreatedAt          time.Time
	FilePath           string
	Name               string
	MediaTitle         string
	Description        sql.NullString
	CoverUrl           sql.NullString
	ParentID           sql.NullInt64
	Classifier         string
	MediaType          string
	ExternalProvider   sql.NullString
	ExternalProviderID sql.NullInt64
}

// ContentRepository defines data access operations for content items.
type ContentRepository interface {
	Create(ctx context.Context, params CreateContentParams) (*domain.ContentFile, error)
	GetByID(ctx context.Context, id int64) (*domain.ContentFile, error)
	GetByPath(ctx context.Context, path string) (*domain.ContentFile, error)
	GetByExternalID(ctx context.Context, externalID int64) (*domain.ContentFile, error)
	GetByParentID(ctx context.Context, parentID int64) ([]domain.ContentFile, error)
	GetAllByLibraryID(ctx context.Context, libraryID int64) ([]domain.ContentFile, error)
	GetAllShows(ctx context.Context, libraryID int64, mediaType string) ([]domain.ContentFile, error)
}
