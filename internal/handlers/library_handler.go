package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/siddhantmadhur/pequod/internal/domain"
	"github.com/siddhantmadhur/pequod/internal/middleware"
	"github.com/siddhantmadhur/pequod/internal/services"
)

type LibraryHandler struct {
	libraryService services.LibraryService
}

func NewLibraryHandler(libraryService services.LibraryService) *LibraryHandler {
	return &LibraryHandler{
		libraryService: libraryService,
	}
}

func (h *LibraryHandler) GetLibraries(c echo.Context) error {
	libraries, err := h.libraryService.GetLibraries(c.Request().Context())
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	var result []map[string]string
	for _, lib := range libraries {
		result = append(result, map[string]string{
			"id":          fmt.Sprint(lib.ID),
			"name":        lib.Name,
			"path":        lib.DevicePath,
			"description": lib.Description,
			"owner_uid":   fmt.Sprint(lib.OwnerID),
			"media_type":  lib.MediaType,
		})
	}
	return c.JSON(http.StatusOK, result)
}

type AddLibraryRequest struct {
	Path        string `json:"path"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

func (h *LibraryHandler) AddLibrary(c echo.Context) error {
	var req AddLibraryRequest
	if err := c.Bind(&req); err != nil {
		return c.String(http.StatusBadRequest, "Invalid request payload")
	}

	if req.Path == "" {
		return c.String(http.StatusBadRequest, "Request is invalid: Path not mentioned.")
	}

	currentUser := middleware.CurrentUser(c)
	_, err := h.libraryService.AddLibrary(c.Request().Context(), currentUser, req.Name, req.Path, req.Type, req.Description)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusCreated)
}

func (h *LibraryHandler) GetContentFromLibrary(c echo.Context) error {
	libraryID, err := strconv.ParseInt(c.QueryParam("library"), 10, 64)
	if err != nil {
		return HandleError(c, domain.ErrInvalidInput)
	}

	mediaType := c.Param("mediaType")
	content, err := h.libraryService.GetContentByMediaType(c.Request().Context(), libraryID, mediaType)
	if err != nil {
		return HandleError(c, err)
	}

	return c.JSON(http.StatusOK, content)
}

func (h *LibraryHandler) GetVideoContentFromMedia(c echo.Context) error {
	mediaID, err := strconv.ParseInt(c.Param("mediaId"), 10, 64)
	if err != nil {
		return HandleError(c, domain.ErrInvalidInput)
	}

	children, err := h.libraryService.GetMediaChildren(c.Request().Context(), mediaID)
	if err != nil {
		return HandleError(c, err)
	}

	return c.JSON(http.StatusOK, children)
}
