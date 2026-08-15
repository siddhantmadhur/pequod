package services

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/siddhantmadhur/pequod/internal/domain"
	"github.com/siddhantmadhur/pequod/internal/repository"
)

type LibraryService interface {
	GetLibraries(ctx context.Context) ([]domain.MediaLibrary, error)
	AddLibrary(ctx context.Context, currentUser *domain.User, name, path, mediaType, description string) (*domain.MediaLibrary, error)
	GetContentByMediaType(ctx context.Context, libraryID int64, mediaType string) ([]domain.ContentDetail, error)
	GetMediaChildren(ctx context.Context, mediaID int64) ([]domain.ContentDetail, error)
	ListFolders(directory string) ([]domain.FolderItem, error)
}

type libraryService struct {
	libraryRepo repository.LibraryRepository
	contentRepo repository.ContentRepository
	userRepo    repository.UserRepository
	scanner     ScannerService
}

func NewLibraryService(
	libraryRepo repository.LibraryRepository,
	contentRepo repository.ContentRepository,
	userRepo repository.UserRepository,
	scanner ScannerService,
) LibraryService {
	return &libraryService{
		libraryRepo: libraryRepo,
		contentRepo: contentRepo,
		userRepo:    userRepo,
		scanner:     scanner,
	}
}

func (s *libraryService) GetLibraries(ctx context.Context) ([]domain.MediaLibrary, error) {
	return s.libraryRepo.GetAll(ctx)
}

func (s *libraryService) AddLibrary(ctx context.Context, currentUser *domain.User, name, path, mediaType, description string) (*domain.MediaLibrary, error) {
	if path == "" {
		return nil, domain.ErrInvalidInput
	}

	var userID int64
	if currentUser == nil {
		admin, err := s.userRepo.GetAdmin(ctx)
		if err != nil {
			return nil, err
		}
		userID = admin.ID
	} else {
		userID = currentUser.UID
	}

	lib, err := s.libraryRepo.Create(ctx, repository.CreateLibraryParams{
		Name:        name,
		DevicePath:  path,
		MediaType:   mediaType,
		Description: description,
		OwnerID:     userID,
		CreatedAt:   time.Now(),
	})
	if err != nil {
		return nil, err
	}

	// Trigger library scan asynchronously
	go func(library domain.MediaLibrary) {
		_ = s.scanner.ScanLibrary(context.Background(), &library)
	}(*lib)

	return lib, nil
}

func (s *libraryService) GetContentByMediaType(ctx context.Context, libraryID int64, mediaType string) ([]domain.ContentDetail, error) {
	contents, err := s.contentRepo.GetAllShows(ctx, libraryID, mediaType)
	if err != nil {
		return nil, err
	}

	var results []domain.ContentDetail
	for _, c := range contents {
		results = append(results, domain.ContentDetail{
			ID:                 c.ID,
			Title:              c.MediaTitle,
			Description:        c.Description.String,
			CoverUrl:           c.CoverUrl.String,
			MediaType:          c.MediaType,
			ExternalProviderID: c.ExternalProviderID.Int64,
			ExternalProvider:   c.ExternalProvider.String,
		})
	}
	return results, nil
}

func (s *libraryService) GetMediaChildren(ctx context.Context, mediaID int64) ([]domain.ContentDetail, error) {
	contents, err := s.contentRepo.GetByParentID(ctx, mediaID)
	if err != nil {
		return nil, err
	}

	seasonRegex := regexp.MustCompile(`S[0-9]+`)
	episodeRegex := regexp.MustCompile(`E[0-9]+`)
	numRegex := regexp.MustCompile(`[0-9]+`)

	var results []domain.ContentDetail
	for _, c := range contents {
		detail := domain.ContentDetail{
			ID:                 c.ID,
			Title:              c.MediaTitle,
			Description:        c.Description.String,
			ExternalProvider:   c.ExternalProvider.String,
			ExternalProviderID: c.ExternalProviderID.Int64,
			MediaType:          c.MediaType,
		}

		if c.MediaType == "episode" {
			detail.Season = numRegex.FindString(seasonRegex.FindString(c.Classifier))
			detail.Episode = numRegex.FindString(episodeRegex.FindString(c.Classifier))
		}
		results = append(results, detail)
	}

	return results, nil
}

func (s *libraryService) ListFolders(directory string) ([]domain.FolderItem, error) {
	if directory == "" {
		directory = "/"
	}

	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, fmt.Errorf("could not read directory: %w", err)
	}

	var folders []domain.FolderItem
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			folders = append(folders, domain.FolderItem{
				Name: entry.Name(),
			})
		}
	}
	return folders, nil
}
