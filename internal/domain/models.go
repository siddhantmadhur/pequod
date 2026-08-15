package domain

import (
	"database/sql"
	"time"
)

// User represents an authenticated user in the system.
type User struct {
	UID             int64  `json:"uid"`
	Username        string `json:"username"`
	AccessToken     string `json:"-"`
	RefreshToken    string `json:"-"`
	ProfilePicture  string `json:"profile_picture,omitempty"`
	PermissionLevel int    `json:"-"`
}

// Profile represents the user entity stored in the database.
type Profile struct {
	ID       int64
	Username string
	Password []byte
	Type     int64
}

// Session represents an active authentication session.
type Session struct {
	ID               string
	UserID           int64
	AccessToken      string
	RefreshToken     string
	CreatedAt        time.Time
	AccessExpiresAt  time.Time
	RefreshExpiresAt time.Time
	Device           string
	DeviceName       string
	ClientName       string
	ClientVersion    string
}

// MediaLibrary represents a root media collection (movies or series).
type MediaLibrary struct {
	ID          int64     `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	DevicePath  string    `json:"path"`
	MediaType   string    `json:"media_type"`
	OwnerID     int64     `json:"owner_uid"`
}

// ContentFile represents an indexed media item (movie, series, season, episode, stream).
type ContentFile struct {
	ID                 int64          `json:"id"`
	MediaLibraryID     int64          `json:"media_library_id"`
	CreatedAt          time.Time      `json:"created_at"`
	FilePath           string         `json:"file_path"`
	Name               string         `json:"name"`
	MediaTitle         string         `json:"title"`
	Description        sql.NullString `json:"description"`
	CoverUrl           sql.NullString `json:"cover_url"`
	ParentID           sql.NullInt64  `json:"parent_id"`
	Classifier         string         `json:"classifier"`
	MediaType          string         `json:"media_type"`
	ExternalProvider   sql.NullString `json:"external_provider"`
	ExternalProviderID sql.NullInt64  `json:"external_provider_id"`
}

// ContentDetail is an enriched view for content items returned to the client.
type ContentDetail struct {
	ID                 int64  `json:"id,omitempty"`
	Title              string `json:"title"`
	Description        string `json:"description"`
	CoverUrl           string `json:"cover_url,omitempty"`
	MediaType          string `json:"media_type"`
	ExternalProviderID int64  `json:"external_provider_id,omitempty"`
	ExternalProvider   string `json:"external_provider,omitempty"`
	Season             string `json:"season,omitempty"`
	Episode            string `json:"episode,omitempty"`
}

// ServerInformation holds runtime details about the host server.
type ServerInformation struct {
	Hostname        string `json:"hostname"`
	ServerVersion   string `json:"server_version"`
	OperatingSystem string `json:"operating_system"`
	FinishedWizard  bool   `json:"finished_wizard"`
}

// FolderItem represents a directory entry on the host filesystem.
type FolderItem struct {
	Name string `json:"name"`
}

// TokenPair represents access and refresh tokens returned on login/refresh.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// PlaybackInfoResponse represents playback configuration for a video stream.
type PlaybackInfoResponse struct {
	SessionID string `json:"session_id,omitempty"`
	MediaID   string `json:"media_id"`
	Preset    string `json:"preset,omitempty"`
	StreamURL string `json:"stream_url"`
	UserID    string `json:"user_id"`
}

// StreamingSessionInfo represents an active HLS transcoding session.
type StreamingSessionInfo struct {
	SessionID string `json:"session_id"`
	MediaID   string `json:"media_id"`
}
