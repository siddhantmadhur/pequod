package content_test

import (
	"os"
	"testing"

	"github.com/siddhantmadhur/pequod/pkg/content"
	"github.com/siddhantmadhur/pequod/pkg/content/tmdb"
)

func TestTMDBClient(t *testing.T) {
	readToken := os.Getenv("TMDB_READ_TOKEN")
	if readToken == "" {
		t.Skip("TMDB_READ_TOKEN env variable not provided, skipping external integration test.")
	}

	client, err := content.NewClient(tmdb.Client{
		ApiKey: readToken,
	})
	if err != nil {
		t.Fatalf("unexpected client init error: %v", err)
	}

	if !client.Authenticate() {
		t.Fatal("could not authenticate client with TMDB")
	}
}
