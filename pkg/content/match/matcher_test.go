package matcher

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"
)

func TestGrabbingShowData(t *testing.T) {
	season, err := regexp.Compile(`(S[0-9]{2,})|(Season\s[0-9]{1,})`)
	num, err := regexp.Compile(`[0-9]+`)
	episode, err := regexp.Compile(`(E[0-9]{2,})|(Episode\s[0-9]{1,})`)
	if err != nil {
		fmt.Printf("Regex did not compile\n")
		t.FailNow()
	}

	type Value struct {
		Name            string
		Got             int
		ExpectedSeason  int
		ExpectedEpisode int
	}

	var seasons = []Value{
		Value{
			Name:            "/Teenage Mutant NINJA TURTLES [2012-2017]/Season 1 (2012-13)/TMNT - S01 E01 - Rise of the Turtles, Part 1 (720p Web-DL).mp4",
			ExpectedSeason:  1,
			ExpectedEpisode: 1,
		},
		Value{
			Name:            "/Teenage Mutant NINJA TURTLES [2012-2017]/Season 2 (2012-13)/TMNT - S02 E01 - Rise of the Turtles, Part 1 (720p Web-DL).mp4",
			ExpectedSeason:  2,
			ExpectedEpisode: 1,
		},
		Value{
			Name:            "/Teenage Mutant NINJA TURTLES [2012-2017]/Season 1 (2012-13)/TMNT - S01 E03 - Rise of the Turtles, Part 1 (720p Web-DL).mp4",
			ExpectedSeason:  1,
			ExpectedEpisode: 3,
		},
	}
	for _, val := range seasons {
		seasonString := season.FindString(val.Name)
		val.Got, err = strconv.Atoi(num.FindString(seasonString))
		if err != nil {
			t.Fail()
			fmt.Printf("Did not get number\n")
		}
		if val.ExpectedSeason != val.Got {
			t.Fail()
			fmt.Printf("Got: %d, Expected: %d\n", val.Got, val.ExpectedSeason)
		}

		episodeString := episode.FindString(val.Name)
		ep, err := strconv.Atoi(num.FindString(episodeString))
		if err != nil {
			t.Fail()
			fmt.Printf("Did not get number\n")
		}

		if ep != val.ExpectedEpisode {
			t.Fail()
			fmt.Printf("Got: %d, Expected: %d\n", val.Got, val.ExpectedSeason)
		}
	}
}
