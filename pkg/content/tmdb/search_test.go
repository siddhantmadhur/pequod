package tmdb

import (
	"os"
	"testing"

	"github.com/siddhantmadhur/pequod/pkg/content"
)

func TestSearchMovie(t *testing.T) {
	readToken := os.Getenv("TMDB_READ_TOKEN")
	if readToken == "" {
		t.Skip("TMDB_READ_TOKEN env variable not provided, skipping external test.")
	}

	var tmdb Client
	tmdb.ApiKey = readToken

	res, err := tmdb.SearchMovies(content.SearchParam{
		Query: "star wars empire strikes back",
	})
	if err != nil {
		t.Fatalf("unexpected search error: %v", err)
	}

	if len(res.Results) == 0 {
		t.Fatal("did not get any results")
	}

	if res.Results[0].Title != "The Empire Strikes Back" {
		t.Fatalf("expected The Empire Strikes Back, got %s", res.Results[0].Title)
	}
}

func TestSearchShows(t *testing.T) {
	readToken := os.Getenv("TMDB_READ_TOKEN")
	if readToken == "" {
		t.Skip("TMDB_READ_TOKEN env variable not provided, skipping external test.")
	}

	var tmdb Client
	tmdb.ApiKey = readToken

	res, err := tmdb.SearchShows(content.SearchParam{
		Query: "modern family",
	})
	if err != nil {
		t.Fatalf("unexpected search error: %v", err)
	}

	if len(res.Results) == 0 {
		t.Fatal("did not get any results")
	}
}

func TestGetSeasonInfo(t *testing.T) {
	readToken := os.Getenv("TMDB_READ_TOKEN")
	if readToken == "" {
		t.Skip("TMDB_READ_TOKEN env variable not provided, skipping external test.")
	}

	var tmdb Client
	tmdb.ApiKey = readToken

	res, err := tmdb.GetSeasonInformation(76479, 2)
	if err != nil {
		t.Fatalf("unexpected season fetch error: %v", err)
	}

	if res.Name != "Season 2" {
		t.Fatalf("expected Season 2, got %s", res.Name)
	}
}
