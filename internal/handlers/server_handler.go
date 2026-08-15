package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/siddhantmadhur/pequod/internal/domain"
	"github.com/siddhantmadhur/pequod/internal/services"
)

type ServerHandler struct {
	serverService  services.ServerService
	libraryService services.LibraryService
}

func NewServerHandler(serverService services.ServerService, libraryService services.LibraryService) *ServerHandler {
	return &ServerHandler{
		serverService:  serverService,
		libraryService: libraryService,
	}
}

func (h *ServerHandler) GetServerInformation(c echo.Context) error {
	info, err := h.serverService.GetServerInfo()
	if err != nil {
		return HandleError(c, err)
	}
	return c.JSON(http.StatusOK, info)
}

func (h *ServerHandler) GetPathFolders(c echo.Context) error {
	directory := c.QueryParam("directory")
	folders, err := h.libraryService.ListFolders(directory)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, folders)
}

func (h *ServerHandler) FinishWizard(c echo.Context) error {
	if err := h.serverService.FinishWizard(); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusOK)
}

func (h *ServerHandler) GetWizardFirstUser(c echo.Context) error {
	if h.serverService.IsWizardCompleted() {
		return HandleError(c, domain.ErrWizardAlreadyDone)
	}
	return c.String(http.StatusOK, "User")
}
