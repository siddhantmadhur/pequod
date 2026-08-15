package tmdb

import (
	"fmt"
	"os"
	"testing"

	"github.com/siddhantmadhur/pequod/pkg/content"
)

func TestFetch(t *testing.T) {
	readToken := os.Getenv("TMDB_READ_TOKEN")
	if readToken == "" {
		fmt.Printf("[ERROR]: TMDB_READ_TOKEN env variable not provided.\n")
		t.FailNow()
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
		fmt.Printf("[ERROR]: %s\n", err.Error())
		t.FailNow()
	}

	if result.Title != "Star Wars" {
		fmt.Printf("[ERROR]: \nExpected: Star Wars \nGot: %s \n", result.Title)

		t.FailNow()
	}
}

func TestGetFromId(t *testing.T) {
	readToken := os.Getenv("TMDB_READ_TOKEN")
	if readToken == "" {
		fmt.Printf("[ERROR]: TMDB_READ_TOKEN env variable not provided.\n")
		t.FailNow()
	}

	var tmdb Client
	tmdb.ApiKey = readToken

	result, err := tmdb.GetFromId(11)

	if err != nil {
		fmt.Printf("[ERROR]: %s\n", err.Error())
		t.FailNow()
	}

	if result.Title != "Star Wars" {
		fmt.Printf("[ERROR]: Title does not match Star Wars\n")
		t.FailNow()
	}

}
