package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/siddhantmadhur/pequod/internal/domain"
	"github.com/siddhantmadhur/pequod/internal/middleware"
	"github.com/siddhantmadhur/pequod/internal/services"
)

type MediaHandler struct {
	mediaService services.MediaService
}

func NewMediaHandler(mediaService services.MediaService) *MediaHandler {
	return &MediaHandler{
		mediaService: mediaService,
	}
}

type PlaybackInfoRequest struct {
	Preset          string `json:"preset"`
	PlaybackSeconds int64  `json:"playback_seconds"`
	DirectPlay      bool   `json:"direct_play"`
}

func (h *MediaHandler) GetPlaybackInfo(c echo.Context) error {
	mediaID, err := strconv.ParseInt(c.Param("mediaId"), 10, 64)
	if err != nil {
		return HandleError(c, domain.ErrInvalidInput)
	}

	var req PlaybackInfoRequest
	if err := c.Bind(&req); err != nil {
		return HandleError(c, domain.ErrInvalidInput)
	}

	currentUser := middleware.CurrentUser(c)
	var userID int64
	if currentUser != nil {
		userID = currentUser.UID
	}

	resp, err := h.mediaService.GetPlaybackInfo(c.Request().Context(), mediaID, req.Preset, req.PlaybackSeconds, req.DirectPlay, userID)
	if err != nil {
		return HandleError(c, err)
	}

	status := http.StatusOK
	if !req.DirectPlay {
		status = http.StatusCreated
	}

	return c.JSON(status, map[string]string{
		"session_id": resp.SessionID,
		"media_id":   resp.MediaID,
		"preset":     resp.Preset,
		"stream_url": resp.StreamURL,
		"user_id":    resp.UserID,
	})
}

func (h *MediaHandler) GetMasterPlaylist(c echo.Context) error {
	sessionID := c.Param("sessionId")
	playlist, err := h.mediaService.GetMasterPlaylist(sessionID)
	if err != nil {
		return HandleError(c, err)
	}
	return c.String(http.StatusOK, playlist)
}

func (h *MediaHandler) GetStreamFile(c echo.Context) error {
	sessionID := c.Param("sessionId")
	segmentStr := c.Param("segment")
	segmentNo, err := strconv.ParseInt(segmentStr, 10, 64)
	if err != nil {
		return HandleError(c, domain.ErrInvalidInput)
	}

	filePath, err := h.mediaService.GetStreamFile(sessionID, segmentNo)
	if err != nil {
		return HandleError(c, err)
	}

	return c.File(filePath)
}

func (h *MediaHandler) GetDirectPlayVideo(c echo.Context) error {
	mediaID, err := strconv.ParseInt(c.Param("mediaId"), 10, 64)
	if err != nil {
		return HandleError(c, domain.ErrInvalidInput)
	}

	filePath, err := h.mediaService.GetDirectPlayFilePath(c.Request().Context(), mediaID)
	if err != nil {
		return HandleError(c, err)
	}

	return c.File(filePath)
}

func (h *MediaHandler) GetAllSessions(c echo.Context) error {
	sessions := h.mediaService.GetAllSessions()
	var result []map[string]string
	for _, s := range sessions {
		result = append(result, map[string]string{
			"session_id": s.SessionID,
			"media_id":   s.MediaID,
		})
	}
	return c.JSON(http.StatusOK, result)
}
