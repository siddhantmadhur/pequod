package matcher

import (
	"errors"
	"regexp"
	"strconv"

	"github.com/siddhantmadhur/pequod/pkg/content"
)

var (
	seasonRegex  = regexp.MustCompile(`(?i)(?:s|season\s*)([0-9]+)`)
	episodeRegex = regexp.MustCompile(`(?i)(?:e|episode\s*)([0-9]+)`)
	digitsRegex  = regexp.MustCompile(`[0-9]+`)
)

func SeriesData(fullPath string) (content.Show, error) {
	var show content.Show

	seasonMatch := seasonRegex.FindString(fullPath)
	if seasonMatch != "" {
		if numStr := digitsRegex.FindString(seasonMatch); numStr != "" {
			if sNum, err := strconv.Atoi(numStr); err == nil {
				show.SeasonNumber = sNum
			}
		}
	}

	episodeMatch := episodeRegex.FindString(fullPath)
	if episodeMatch != "" {
		if numStr := digitsRegex.FindString(episodeMatch); numStr != "" {
			if epNum, err := strconv.Atoi(numStr); err == nil {
				show.EpisodeNumber = epNum
			}
		}
	}

	if show.SeasonNumber == 0 && show.EpisodeNumber == 0 {
		return show, errors.New("no season or episode information found in path")
	}

	return show, nil
}
