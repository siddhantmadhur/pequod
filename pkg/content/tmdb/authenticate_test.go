package tmdb

import (
	"fmt"
	"os"
	"testing"
)

func TestAuthenticate(t *testing.T) {

	readToken := os.Getenv("TMDB_READ_TOKEN")
	if readToken == "" {
		fmt.Printf("[ERROR]: TMDB_READ_TOKEN env variable not provided.\n")
		t.FailNow()
	}

	var tmdb Client
	tmdb.ApiKey = readToken

	isAuthenticated := tmdb.Authenticate()
	if !isAuthenticated {
		fmt.Printf("[ERROR] Could not authenticate user\n")
		t.FailNow()
	}

}
