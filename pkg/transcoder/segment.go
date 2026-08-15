package transcoder

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

func NewSegment(start, end int64) *Segment {
	return &Segment{
		StartSegment: start,
		EndSegment:   end,
		Lock:         &sync.Mutex{},
	}
}

func (s *Segment) AddNewSegment(newSegmentNo int64) error {
	cur := s.GetCurrentSegment(newSegmentNo)
	if cur == nil {
		return errors.New("segment not found")
	}

	newSegment := &Segment{
		StartSegment: newSegmentNo,
		EndSegment:   cur.EndSegment,
		NextSegment:  cur.NextSegment,
		Lock:         s.Lock,
	}
	cur.EndSegment = newSegmentNo - 1
	cur.NextSegment = newSegment

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
