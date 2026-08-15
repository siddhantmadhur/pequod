package media

import (
	"errors"
	"sync"
)

type Segment struct {
	StartSegment int64
	EndSegment   int64
	NextSegment  *Segment
	Lock         *sync.Mutex
}

func (s *Segment) AddNewSegment(newSegmentNo int64) error {

	cur := s.GetCurrentSegment(newSegmentNo)
	if cur == nil {
		return errors.New("Segment not found.")
	}

	var newSegment Segment

	newSegment.EndSegment = cur.EndSegment
	newSegment.StartSegment = newSegmentNo
	cur.EndSegment = newSegmentNo - 1
	newSegment.NextSegment = cur.NextSegment
	cur.NextSegment = &newSegment

	return nil

}

func (s *Segment) GetCurrentSegment(segmentNo int64) *Segment {
	cur := s

	for cur != nil {
		if cur.StartSegment <= segmentNo && segmentNo <= cur.EndSegment {
			return cur
		}
		cur = cur.NextSegment
	}

	return nil
}
