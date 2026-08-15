package tmdb

import (
	"os"
	"testing"

	"github.com/siddhantmadhur/pequod/pkg/content"
)

func TestFetch(t *testing.T) {
	readToken := os.Getenv("TMDB_READ_TOKEN")
	if readToken == "" {
		t.Skip("TMDB_READ_TOKEN env variable not provided, skipping external test.")
	}

	var tmdb Client
	tmdb.ApiKey = readToken

	var result struct {
		Id    int    `json:"id"`
		Title string `json:"title"`
	}

	err := tmdb.Fetch(content.FetchParams{
		Method:   "GET",
		Endpoint: "/movie/11",
	}, &result)

	if err != nil {
		t.Fatalf("unexpected fetch error: %v", err)
	}

	if result.Title != "Star Wars" {
		t.Fatalf("expected Star Wars, got %s", result.Title)
	}
}

func TestGetFromId(t *testing.T) {
	readToken := os.Getenv("TMDB_READ_TOKEN")
	if readToken == "" {
		t.Skip("TMDB_READ_TOKEN env variable not provided, skipping external test.")
	}

	var tmdb Client
	tmdb.ApiKey = readToken

	result, err := tmdb.GetFromId(11)
	if err != nil {
		t.Fatalf("unexpected GetFromId error: %v", err)
	}

	if result.Title != "Star Wars" {
		t.Fatalf("expected Star Wars, got %s", result.Title)
	}
}
