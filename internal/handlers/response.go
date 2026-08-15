package handlers

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/siddhantmadhur/pequod/internal/domain"
)

type ErrorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

func HandleError(c echo.Context, err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, domain.ErrInvalidCredentials):
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: "Password does not match or there was an error",
			Error:   err.Error(),
		})
	case errors.Is(err, domain.ErrInvalidInput):
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: "Invalid input provided",
			Error:   err.Error(),
		})
	case errors.Is(err, domain.ErrUnauthorized):
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Message: "You are not authorized",
			Error:   err.Error(),
		})
	case errors.Is(err, domain.ErrForbidden):
		return c.JSON(http.StatusForbidden, ErrorResponse{
			Message: "You do not have permission to perform this action",
			Error:   err.Error(),
		})
	case errors.Is(err, domain.ErrNotFound), errors.Is(err, domain.ErrSessionNotFound):
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Message: "Requested resource not found",
			Error:   err.Error(),
		})
	case errors.Is(err, domain.ErrWizardAlreadyDone):
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"msg": "Server is already setup and you no longer have access to this feature",
		})
	default:
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: "Internal server error",
			Error:   err.Error(),
		})
	}
}
