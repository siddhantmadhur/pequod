package handlers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/siddhantmadhur/pequod/internal/domain"
	"github.com/siddhantmadhur/pequod/internal/handlers"
)

// MockServerService implements services.ServerService for handler tests.
type MockServerService struct {
	FinishedWizard bool
}

func (m *MockServerService) GetServerInfo() (*domain.ServerInformation, error) {
	return &domain.ServerInformation{
		Hostname:        "test-host",
		ServerVersion:   "1.0.0",
		OperatingSystem: "linux",
		FinishedWizard:  m.FinishedWizard,
	}, nil
}

func (m *MockServerService) FinishWizard() error {
	m.FinishedWizard = true
	return nil
}

func (m *MockServerService) IsWizardCompleted() bool {
	return m.FinishedWizard
}

// MockLibraryService implements services.LibraryService for handler tests.
type MockLibraryService struct{}

func (m *MockLibraryService) GetLibraries(ctx context.Context) ([]domain.MediaLibrary, error) {
	return []domain.MediaLibrary{}, nil
}

func (m *MockLibraryService) AddLibrary(ctx context.Context, currentUser *domain.User, name, path, mediaType, description string) (*domain.MediaLibrary, error) {
	return &domain.MediaLibrary{ID: 1, Name: name, DevicePath: path}, nil
}

func (m *MockLibraryService) GetContentByMediaType(ctx context.Context, libraryID int64, mediaType string) ([]domain.ContentDetail, error) {
	return []domain.ContentDetail{}, nil
}

func (m *MockLibraryService) GetMediaChildren(ctx context.Context, mediaID int64) ([]domain.ContentDetail, error) {
	return []domain.ContentDetail{}, nil
}

func (m *MockLibraryService) ListFolders(directory string) ([]domain.FolderItem, error) {
	return []domain.FolderItem{{Name: "Movies"}, {Name: "Shows"}}, nil
}

func TestServerInformation(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/server/information", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockServerService := &MockServerService{}
	mockLibraryService := &MockLibraryService{}
	handler := handlers.NewServerHandler(mockServerService, mockLibraryService)

	if err := handler.GetServerInformation(c); err != nil {
		t.Fatalf("unexpected handler error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if !strings.Contains(rec.Body.String(), "test-host") {
		t.Errorf("expected body to contain test-host, got %s", rec.Body.String())
	}
}

func TestGetPathFolders(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/server/information/folders?directory=/media", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockServerService := &MockServerService{}
	mockLibraryService := &MockLibraryService{}
	handler := handlers.NewServerHandler(mockServerService, mockLibraryService)

	if err := handler.GetPathFolders(c); err != nil {
		t.Fatalf("unexpected handler error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if !strings.Contains(rec.Body.String(), "Movies") {
		t.Errorf("expected body to contain Movies, got %s", rec.Body.String())
	}
}
