package matcher_test

import (
	"testing"

	"github.com/siddhantmadhur/pequod/pkg/content/match"
)

func TestGrabbingShowData(t *testing.T) {
	tests := []struct {
		path            string
		expectedSeason  int
		expectedEpisode int
	}{
		{
			path:            "/Teenage Mutant NINJA TURTLES [2012-2017]/Season 1 (2012-13)/TMNT - S01 E01 - Rise of the Turtles, Part 1 (720p Web-DL).mp4",
			expectedSeason:  1,
			expectedEpisode: 1,
		},
		{
			path:            "/Teenage Mutant NINJA TURTLES [2012-2017]/Season 2 (2012-13)/TMNT - S02 E01 - Rise of the Turtles, Part 1 (720p Web-DL).mp4",
			expectedSeason:  2,
			expectedEpisode: 1,
		},
		{
			path:            "/Teenage Mutant NINJA TURTLES [2012-2017]/Season 1 (2012-13)/TMNT - S01 E03 - Rise of the Turtles, Part 1 (720p Web-DL).mp4",
			expectedSeason:  1,
			expectedEpisode: 3,
		},
	}

	for _, tt := range tests {
		show, err := matcher.SeriesData(tt.path)
		if err != nil {
			t.Fatalf("unexpected error parsing %s: %v", tt.path, err)
		}
		if show.SeasonNumber != tt.expectedSeason {
			t.Errorf("expected season %d, got %d for %s", tt.expectedSeason, show.SeasonNumber, tt.path)
		}
		if show.EpisodeNumber != tt.expectedEpisode {
			t.Errorf("expected episode %d, got %d for %s", tt.expectedEpisode, show.EpisodeNumber, tt.path)
		}
	}
}
