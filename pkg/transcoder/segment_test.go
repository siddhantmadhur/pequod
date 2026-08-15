package transcoder

import (
	"fmt"
	"testing"
)

func TestNewSegment(t *testing.T) {
	initialSegment := NewSegment(0, 100)

	err := initialSegment.AddNewSegment(30)
	if err != nil {
		t.Fatalf("unexpected error adding segment 30: %v", err)
	}

	err = initialSegment.AddNewSegment(13)
	if err != nil {
		t.Fatalf("unexpected error adding segment 13: %v", err)
	}

	cur := initialSegment
	idx := 0
	expected := []int64{0, 12, 13, 29, 30, 100}
	for cur != nil {
		if cur.StartSegment != expected[idx] || cur.EndSegment != expected[idx+1] {
			fmt.Printf("Expected: [%d, %d], got: [%d, %d]\n", expected[idx], expected[idx+1], cur.StartSegment, cur.EndSegment)
			t.FailNow()
		}
		idx += 2
		cur = cur.NextSegment
	}
}
