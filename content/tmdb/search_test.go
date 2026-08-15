package tmdb

import (
	"fmt"
	"os"
	"testing"

	"github.com/siddhantmadhur/pequod/content"
)

func TestSearchMovie(t *testing.T) {
	readToken := os.Getenv("TMDB_READ_TOKEN")
	if readToken == "" {
		fmt.Printf("[ERROR]: TMDB_READ_TOKEN env variable not provided.\n")
		t.FailNow()
	}

	var tmdb Client
	tmdb.ApiKey = readToken

	res, err := tmdb.SearchMovies(content.SearchParam{
		Query: "star wars empire strikes back",
	})
	if err != nil {
		fmt.Printf("[ERROR]: %s\n", err.Error())
		t.FailNow()
	}

	if len(res.Results) == 0 {
		fmt.Printf("[ERROR]: Did not get any results\n")
		t.FailNow()
	}

	if res.Results[0].Title != "The Empire Strikes Back" {
		fmt.Printf("[ERROR]: Name does not match\n")
		t.FailNow()
	}
}

func TestSearchShows(t *testing.T) {
	readToken := os.Getenv("TMDB_READ_TOKEN")
	if readToken == "" {
		fmt.Printf("[ERROR]: TMDB_READ_TOKEN env variable not provided.\n")
		t.FailNow()
	}

	var tmdb Client
	tmdb.ApiKey = readToken

	res, err := tmdb.SearchShows(content.SearchParam{
		Query: "modern family",
	})
	if err != nil {
		fmt.Printf("[ERROR]: %s\n", err.Error())
		t.FailNow()
	}

	if len(res.Results) == 0 {
		fmt.Printf("[ERROR]: Did not get any results\n")
		t.FailNow()
	}

}

func TestGetSeasonInfo(t *testing.T) {
	readToken := os.Getenv("TMDB_READ_TOKEN")
	if readToken == "" {
		fmt.Printf("[ERROR]: TMDB_READ_TOKEN env variable not provided.\n")
		t.FailNow()
	}

	var tmdb Client
	tmdb.ApiKey = readToken

	res, err := tmdb.GetSeasonInformation(76479, 2)
	if err != nil {
		fmt.Printf("[ERROR]: %s\n", err.Error())
		t.FailNow()
	}

	if res.Name != "Season 2" {
		t.FailNow()
	}

}
