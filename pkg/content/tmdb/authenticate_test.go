package tmdb

import (
	"os"
	"testing"
)

func TestAuthenticate(t *testing.T) {
	readToken := os.Getenv("TMDB_READ_TOKEN")
	if readToken == "" {
		t.Skip("TMDB_READ_TOKEN env variable not provided, skipping external test.")
	}

	var tmdb Client
	tmdb.ApiKey = readToken

	isAuthenticated := tmdb.Authenticate()
	if !isAuthenticated {
		t.Fatal("Could not authenticate user with TMDB")
	}
}
