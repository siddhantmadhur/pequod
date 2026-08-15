package media

import (
	"fmt"
	"testing"
)

func TestNewSegment(t *testing.T) {
	var initialSegment Segment
	initialSegment.StartSegment = 0
	initialSegment.EndSegment = 100

	initialSegment.AddNewSegment(30)
	initialSegment.AddNewSegment(13)

	cur := &initialSegment
	idx := 0
	expected := []int64{0, 12, 13, 29, 30, 100}
	for cur != nil {
		if cur.StartSegment != expected[idx] && cur.EndSegment != int64(expected[idx+1]) {
			fmt.Printf("Expected: %d, got: %d\n", expected[idx], cur.StartSegment)
			fmt.Printf("Expected: %d, got: %d\n", expected[idx+1], cur.EndSegment)
			t.FailNow()
		}
		idx += 2
		cur = cur.NextSegment
	}
}
