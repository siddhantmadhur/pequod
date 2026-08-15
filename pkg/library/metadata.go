package library

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/siddhantmadhur/pequod/internal/config"
	"github.com/siddhantmadhur/pequod/internal/database"
	"github.com/siddhantmadhur/pequod/pkg/content"
	"github.com/siddhantmadhur/pequod/pkg/content/tmdb"
)

func ScanMediaFiles(library database.MediaLibrary, cfg *config.Config) error {
	client, err := content.NewClient(tmdb.Client{
		ApiKey: os.Getenv("TMDB_READ_TOKEN"),
	})
	if err != nil {
		return err
	}

	conn, queries, err := database.GetConn(cfg)
	if err != nil {
		return err
	}
	defer conn.Close()

	queryFilter, err := regexp.Compile("^[a-zA-Z0-9_ ]+")
	if err != nil {
		return err
	}

	if library.MediaType == "series" {
		showFiles, err := os.ReadDir(library.DevicePath)
		if err != nil {
			return err
		}

		for _, showFile := range showFiles {
			if showFile.IsDir() {
				contentFile, err := queries.GetContentFromPath(context.Background(), library.DevicePath+"/"+showFile.Name())
				if err != nil {
					results, err := client.SearchShows(content.SearchParam{
						Query: queryFilter.FindString(strings.ReplaceAll(showFile.Name(), ".", " ")),
					})
					// Could not find show
					if err != nil || len(results.Results) == 0 {
						contentFile, err = queries.AddNewContentFile(context.Background(), database.AddNewContentFileParams{
							MediaLibraryID: library.ID,
							CreatedAt:      time.Now(),
							FilePath:       library.DevicePath + "/" + showFile.Name(),
							Name:           showFile.Name(),
							MediaTitle:     showFile.Name(),
							MediaType:      "series",
							Classifier:     "show",
						})
						if err != nil {
							log.Printf("error: %s\n", err.Error())
						}
					} else {
						bestResult := results.Results[0]
						contentFile, err = queries.AddNewContentFile(context.Background(), database.AddNewContentFileParams{
							MediaLibraryID:     library.ID,
							CreatedAt:          time.Now(),
							FilePath:           library.DevicePath + "/" + showFile.Name(),
							Name:               showFile.Name(),
							MediaType:          "series",
							Classifier:         "show",
							MediaTitle:         bestResult.Name,
							CoverUrl:           sql.NullString{String: "https://image.tmdb.org/t/p/w500/" + bestResult.PosterPath, Valid: true},
							Description:        sql.NullString{String: bestResult.Overview, Valid: true},
							ExternalProvider:   sql.NullString{String: "tmdb", Valid: true},
							ExternalProviderID: sql.NullInt64{Int64: int64(bestResult.Id), Valid: true},
						})
						if err != nil {
							log.Printf("error: %s\n", err.Error())
						}
					}
				}
				ScanShowForVideos(library.DevicePath+"/"+showFile.Name(), client, contentFile, library, queries)
			}
		}
	} else if library.MediaType == "movies" {
		movies, err := os.ReadDir(library.DevicePath)

		if err != nil {
			return err
		}

		for _, movie := range movies {
			if movie.IsDir() {

				childPath := library.DevicePath + "/" + movie.Name()
				movieContent, err := queries.GetContentFromPath(context.Background(), childPath)
				if err != nil {
					results, err := client.SearchMovies(content.SearchParam{
						Query: queryFilter.FindString(movie.Name()),
					})
					if err != nil || len(results.Results) == 0 {
						//movieContent, err = queries.AddNewContentFile()
						movieContent, err = queries.AddNewContentFile(context.Background(), database.AddNewContentFileParams{
							MediaLibraryID: library.ID,
							CreatedAt:      time.Now(),
							FilePath:       childPath,
							Name:           movie.Name(),
							MediaTitle:     movie.Name(),
							Classifier:     "movie",
							MediaType:      "movies",
						})

					} else {
						bestResult := results.Results[0]
						movieContent, err = queries.AddNewContentFile(context.Background(), database.AddNewContentFileParams{
							MediaLibraryID:     library.ID,
							CreatedAt:          time.Now(),
							FilePath:           childPath,
							Name:               movie.Name(),
							MediaTitle:         bestResult.Title,
							Description:        sql.NullString{String: bestResult.Overview, Valid: true},
							CoverUrl:           sql.NullString{String: "https://image.tmdb.org/t/p/w500/" + bestResult.PosterPath, Valid: true},
							Classifier:         "movie",
							MediaType:          "movies",
							ExternalProvider:   sql.NullString{String: "tmdb", Valid: true},
							ExternalProviderID: sql.NullInt64{Int64: int64(bestResult.Id), Valid: true},
						})
					}
				}

				filepath.WalkDir(childPath, func(path string, d fs.DirEntry, err error) error {

					if d.IsDir() {
						return nil
					}
					_, err = queries.GetContentFromPath(context.Background(), path)
					if err != nil {
						_, err = queries.AddNewContentFile(context.Background(), database.AddNewContentFileParams{
							MediaLibraryID: library.ID,
							CreatedAt:      time.Now(),
							FilePath:       path,
							Name:           d.Name(),
							MediaTitle:     d.Name(),
							MediaType:      "video-stream",
							Classifier:     "stream",
							ParentID:       sql.NullInt64{Int64: movieContent.ID, Valid: true},
						})
					}
					return nil
				})

			}
		}

	}

	return err
}

func ScanShowForVideos(currentPath string, client content.Client, parentLibrary database.ContentLibrary, mediaLibrary database.MediaLibrary, queries *database.Queries) error {

	files, err := os.ReadDir(currentPath)
	if err != nil {
		return err
	}

	for _, file := range files {
		childPath := currentPath + "/" + file.Name()
		if file.IsDir() {
			ScanShowForVideos(childPath, client, parentLibrary, mediaLibrary, queries)
		} else {
			_, err := FFprobe(childPath)
			if err != nil {
				continue
			}
			_, err = queries.GetContentFromPath(context.Background(), childPath)
			// Does not exist in db
			if err != nil {
				_, seasonNo, episodeNo, err := getShowInformation(currentPath, file.Name())
				if err == nil && parentLibrary.ExternalProviderID.Valid {
					episodeInformation, err := client.GetEpisodeInformation(int(parentLibrary.ExternalProviderID.Int64), seasonNo, episodeNo)
					if err == nil {
						queries.AddNewContentFile(context.Background(), database.AddNewContentFileParams{
							MediaLibraryID:     mediaLibrary.ID,
							CreatedAt:          time.Now(),
							FilePath:           childPath,
							Name:               file.Name(),
							MediaTitle:         episodeInformation.Name,
							Description:        sql.NullString{String: episodeInformation.Overview, Valid: true},
							ParentID:           sql.NullInt64{Int64: parentLibrary.ID, Valid: true},
							ExternalProvider:   sql.NullString{String: "tmdb", Valid: true},
							ExternalProviderID: sql.NullInt64{Int64: int64(episodeInformation.Id), Valid: true},
							MediaType:          "episode",
							Classifier:         fmt.Sprintf("S%dE%d", seasonNo, episodeNo),
						})
						continue
					}
				}
				_, err = queries.AddNewContentFile(context.Background(), database.AddNewContentFileParams{
					MediaLibraryID: mediaLibrary.ID,
					CreatedAt:      time.Now(),
					FilePath:       childPath,
					Name:           file.Name(),
					MediaTitle:     file.Name(),
					ParentID:       sql.NullInt64{Int64: parentLibrary.ID, Valid: true},
					MediaType:      "episode",
					Classifier:     fmt.Sprintf("S%dE%d", seasonNo, episodeNo),
				})
			}
		}
	}

	return nil
}
