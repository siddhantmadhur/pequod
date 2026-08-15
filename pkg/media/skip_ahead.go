package media

func (f *Ffmpeg) SkipTo(segmentNo int64) error {
	f.SegmentBuffer.Lock.Lock()

	if doesSegmentExist(f, segmentNo) {
		f.SegmentBuffer.Lock.Unlock()
		return nil
	}

	if f.LastSegment+5 > int(segmentNo) {
		for !doesSegmentExist(f, segmentNo) {
		}
		f.SegmentBuffer.Lock.Unlock()
		return nil
	}

	err := f.SegmentBuffer.AddNewSegment(segmentNo)
	if err != nil {
		return err
	}

	segmentToSkipTo := f.SegmentBuffer.GetCurrentSegment(segmentNo)
	f.CurrentPlaybackSecond = segmentToSkipTo.StartSegment * 2
	f.StopPlaybackSecond = segmentToSkipTo.EndSegment * 2

	f.Start()

	f.SegmentBuffer.Lock.Unlock()

	for !doesSegmentExist(f, segmentNo) {
	}

	for f.LastSegment < int(segmentNo) {
	}

	return nil

}
