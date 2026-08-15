package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/siddhantmadhur/pequod/internal/domain"
	"github.com/siddhantmadhur/pequod/internal/repository"
	"github.com/siddhantmadhur/pequod/pkg/content"
	"github.com/siddhantmadhur/pequod/pkg/content/tmdb"
	"github.com/siddhantmadhur/pequod/pkg/transcoder"
)

type ScannerService interface {
	ScanLibrary(ctx context.Context, lib *domain.MediaLibrary) error
}

type scannerService struct {
	contentRepo repository.ContentRepository
}

func NewScannerService(contentRepo repository.ContentRepository) ScannerService {
	return &scannerService{
		contentRepo: contentRepo,
	}
}

func (s *scannerService) ScanLibrary(ctx context.Context, lib *domain.MediaLibrary) error {
	client, err := content.NewClient(tmdb.Client{
		ApiKey: os.Getenv("TMDB_READ_TOKEN"),
	})
	if err != nil {
		return err
	}

	queryFilter, err := regexp.Compile(`^[a-zA-Z0-9_ ]+`)
	if err != nil {
		return err
	}

	if lib.MediaType == "series" {
		showFiles, err := os.ReadDir(lib.DevicePath)
		if err != nil {
			return err
		}

		for _, showFile := range showFiles {
			if !showFile.IsDir() {
				continue
			}

			fullPath := filepath.Join(lib.DevicePath, showFile.Name())
			contentFile, err := s.contentRepo.GetByPath(ctx, fullPath)
			if err != nil {
				cleanedName := queryFilter.FindString(strings.ReplaceAll(showFile.Name(), ".", " "))
				results, searchErr := client.SearchShows(content.SearchParam{
					Query: cleanedName,
				})

				if searchErr != nil || len(results.Results) == 0 {
					contentFile, err = s.contentRepo.Create(ctx, repository.CreateContentParams{
						MediaLibraryID: lib.ID,
						CreatedAt:      time.Now(),
						FilePath:       fullPath,
						Name:           showFile.Name(),
						MediaTitle:     showFile.Name(),
						MediaType:      "series",
						Classifier:     "show",
					})
					if err != nil {
						log.Printf("[ERROR]: failed to add content file %s: %v\n", fullPath, err)
					}
				} else {
					bestResult := results.Results[0]
					contentFile, err = s.contentRepo.Create(ctx, repository.CreateContentParams{
						MediaLibraryID:     lib.ID,
						CreatedAt:          time.Now(),
						FilePath:           fullPath,
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
						log.Printf("[ERROR]: failed to add content file %s: %v\n", fullPath, err)
					}
				}
			}

			if contentFile != nil {
				s.scanShowForVideos(ctx, fullPath, client, contentFile, lib)
			}
		}
	} else if lib.MediaType == "movies" {
		movies, err := os.ReadDir(lib.DevicePath)
		if err != nil {
			return err
		}

		for _, movie := range movies {
			if !movie.IsDir() {
				continue
			}

			childPath := filepath.Join(lib.DevicePath, movie.Name())
			movieContent, err := s.contentRepo.GetByPath(ctx, childPath)
			if err != nil {
				results, searchErr := client.SearchMovies(content.SearchParam{
					Query: queryFilter.FindString(movie.Name()),
				})

				if searchErr != nil || len(results.Results) == 0 {
					movieContent, err = s.contentRepo.Create(ctx, repository.CreateContentParams{
						MediaLibraryID: lib.ID,
						CreatedAt:      time.Now(),
						FilePath:       childPath,
						Name:           movie.Name(),
						MediaTitle:     movie.Name(),
						Classifier:     "movie",
						MediaType:      "movies",
					})
				} else {
					bestResult := results.Results[0]
					movieContent, err = s.contentRepo.Create(ctx, repository.CreateContentParams{
						MediaLibraryID:     lib.ID,
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

			if movieContent != nil {
				_ = filepath.WalkDir(childPath, func(path string, d fs.DirEntry, walkErr error) error {
					if walkErr != nil || d.IsDir() {
						return nil
					}
					_, err = s.contentRepo.GetByPath(ctx, path)
					if err != nil {
						_, _ = s.contentRepo.Create(ctx, repository.CreateContentParams{
							MediaLibraryID: lib.ID,
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

	return nil
}

func (s *scannerService) scanShowForVideos(ctx context.Context, currentPath string, client content.Client, parentContent *domain.ContentFile, lib *domain.MediaLibrary) {
	files, err := os.ReadDir(currentPath)
	if err != nil {
		return
	}

	for _, file := range files {
		childPath := filepath.Join(currentPath, file.Name())
		if file.IsDir() {
			s.scanShowForVideos(ctx, childPath, client, parentContent, lib)
		} else {
			if _, err := transcoder.FFprobe(childPath); err != nil {
				continue
			}

			if _, err := s.contentRepo.GetByPath(ctx, childPath); err != nil {
				_, seasonNo, episodeNo, parseErr := parseShowInfo(currentPath, file.Name())
				if parseErr == nil && parentContent.ExternalProviderID.Valid {
					epInfo, fetchErr := client.GetEpisodeInformation(int(parentContent.ExternalProviderID.Int64), seasonNo, episodeNo)
					if fetchErr == nil {
						_, _ = s.contentRepo.Create(ctx, repository.CreateContentParams{
							MediaLibraryID:     lib.ID,
							CreatedAt:          time.Now(),
							FilePath:           childPath,
							Name:               file.Name(),
							MediaTitle:         epInfo.Name,
							Description:        sql.NullString{String: epInfo.Overview, Valid: true},
							ParentID:           sql.NullInt64{Int64: parentContent.ID, Valid: true},
							ExternalProvider:   sql.NullString{String: "tmdb", Valid: true},
							ExternalProviderID: sql.NullInt64{Int64: int64(epInfo.Id), Valid: true},
							MediaType:          "episode",
							Classifier:         fmt.Sprintf("S%dE%d", seasonNo, episodeNo),
						})
						continue
					}
				}

				_, _ = s.contentRepo.Create(ctx, repository.CreateContentParams{
					MediaLibraryID: lib.ID,
					CreatedAt:      time.Now(),
					FilePath:       childPath,
					Name:           file.Name(),
					MediaTitle:     file.Name(),
					ParentID:       sql.NullInt64{Int64: parentContent.ID, Valid: true},
					MediaType:      "episode",
					Classifier:     fmt.Sprintf("S%dE%d", seasonNo, episodeNo),
				})
			}
		}
	}
}

func parseShowInfo(fullPath, name string) (string, int, int, error) {
	tokens := strings.Split(fullPath, "/")
	if len(tokens) < 3 {
		return "", 0, 0, errors.New("not enough path information")
	}

	getNumber := regexp.MustCompile(`[0-9]+`)
	getSeasonString := regexp.MustCompile(`s[0-9]+|S[0-9]+|Season [0-9]+`)
	getEpisodeString := regexp.MustCompile(`e[0-9]+|E[0-9]+|Episode [0-9]+`)

	seasonString := getSeasonString.FindString(fullPath)
	episodeString := getEpisodeString.FindString(name)

	season, err := strconv.Atoi(getNumber.FindString(seasonString))
	if err != nil {
		return "", 0, 0, err
	}
	episode, err := strconv.Atoi(getNumber.FindString(episodeString))
	if err != nil {
		return "", 0, 0, err
	}

	return strings.ReplaceAll(tokens[1], ".", " "), season, episode, nil
}
