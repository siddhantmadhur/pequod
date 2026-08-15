package library

import (
	"log"
	"testing"
)

func TestGetDepth(t *testing.T) {
	inputRoots := []string{
		"/library/",
		"/library/show-1",
		"/library/show-1",
		"/library/show-1/season-1",
		"/library/show-1/season-1/episode-1",
		"/library/show-2/season-2",
	}
	inputComplete := []string{
		"/library/show-1/season-1",
		"/library/show-1/season-1",
		"/library/show-1/season-1/episode-2",
		"/library/show-1/season-1/episode-2",
		"/library/show-1/season-1/episode-1",
		"/library/show-2/season-2/episode-1/",
	}

	expectedOutput := []int{
		2,
		1,
		2,
		1,
		0,
		1,
	}

	for i := range len(inputRoots) {
		depth := getFileDepth(inputRoots[i], inputComplete[i])
		if depth != expectedOutput[i] {
			t.Fail()
			log.Printf("Expected: %d, Got: %d\n", expectedOutput[i], depth)
		}
	}

}
