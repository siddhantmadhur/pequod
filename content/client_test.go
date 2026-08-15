package content

import (
	"fmt"
	"os"
	"testing"

	"github.com/siddhantmadhur/pequod/content/tmdb"
)

func TestTMDBClient(t *testing.T) {
	client, err := NewClient(tmdb.Client{
		ApiKey: os.Getenv("TMDB_READ_TOKEN"),
	})

	if err != nil {
		fmt.Printf("[ERROR]: %s\n", err.Error())
		t.FailNow()
	}

	authenticate := client.Authenticate()

	if !authenticate {
		fmt.Printf("[ERROR]: Could not authenticate client\n")
		t.FailNow()
	}
}
