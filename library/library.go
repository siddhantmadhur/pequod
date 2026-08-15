package library

import (
	"context"
	"fmt"
	"time"

	"github.com/siddhantmadhur/pequod/auth"
	"github.com/siddhantmadhur/pequod/config"
	"github.com/siddhantmadhur/pequod/storage"

	"github.com/labstack/echo/v4"
)

func GetLibraryFolders(c echo.Context, u *auth.User, cfg *config.Config) error {
	conn, query, err := storage.GetConn(cfg)
	defer conn.Close()

	if err != nil {
		return c.String(500, err.Error())
	}
	libraries, err := query.GetAllMediaLibraries(context.Background())
	var result []map[string]string
	for _, library := range libraries {
		result = append(result, map[string]string{
			"id":          fmt.Sprint(library.ID),
			"name":        library.Name,
			"path":        library.DevicePath,
			"description": library.Description,
			"owner_uid":   fmt.Sprint(library.OwnerID),
			"media_type":  library.MediaType,
		})
	}
	return c.JSON(200, result)
}

func AddLibraryFolder(c echo.Context, u *auth.User, cfg *config.Config) error {

	var request struct {
		Path        string `json:"path"`
		Name        string `json:"name"`
		Type        string `json:"type"`
		Description string `json:"description"`
	}
	err := c.Bind(&request)
	if err != nil {
		return c.String(500, err.Error())
	}
	if request.Path == "" {
		return c.String(500, "Request is invalid: Path not mentioned.")
	}
	conn, queries, err := storage.GetConn(cfg)
	defer conn.Close()
	if err != nil {
		return c.String(500, err.Error())
	}
	var userId int64
	if u == nil {

		user, err := queries.GetAdminUser(context.Background())
		if err != nil {
			return c.String(500, err.Error())
		}
		userId = user.ID
	} else {
		userId = u.UID
	}
	library, err := queries.CreateMediaLibrary(context.Background(), storage.CreateMediaLibraryParams{
		OwnerID:     userId,
		Name:        request.Name,
		DevicePath:  request.Path,
		MediaType:   request.Type,
		CreatedAt:   time.Now(),
		Description: request.Description,
	})
	if err != nil {
		return c.String(500, err.Error())
	}

	go ScanMediaFiles(library, cfg)

	return c.NoContent(201)
}
