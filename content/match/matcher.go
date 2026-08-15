package matcher

import (
	"regexp"
	"strconv"

	"github.com/siddhantmadhur/pequod/content"
)

func SeriesData(fullPath string) (content.Show, error) {
	var show content.Show

	seasonString, err := regexp.Compile(`(S[0-9]{2,})|(Season\s[0-9]{1,})`)
	num, err := regexp.Compile(`[0-9]+`)
	if err != nil {
		return show, err
	}

	show.SeasonNumber, err = strconv.Atoi(num.FindString(seasonString.String()))
	if err != nil {
		return show, err
	}

	return show, nil
}
