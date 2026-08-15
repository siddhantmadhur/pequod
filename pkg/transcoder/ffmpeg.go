package transcoder

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
)

type FfmpegSession struct {
	lock                  sync.Mutex
	ID                    string
	Proc                  *exec.Cmd
	Preset                string
	CurrentPlaybackSecond int64
	StopPlaybackSecond    int64
	CurrentPath           string
	TranscodePath         string
	MediaID               int64
	StreamURL             string
	LastSegment           int
	SegmentBuffer         *Segment
	KillSignal            chan bool
}

func NewFfmpegSession(preset string, sourcePath string, currentPlaybackSecond int64, cacheDir string, mediaID int64) (*FfmpegSession, error) {
	if preset == "" {
		preset = "veryfast"
	}

	session := &FfmpegSession{
		ID:                    uuid.NewString(),
		Preset:                preset,
		CurrentPlaybackSecond: currentPlaybackSecond,
		CurrentPath:           sourcePath,
		MediaID:               mediaID,
		KillSignal:            make(chan bool),
	}

	session.TranscodePath = fmt.Sprintf("%s/%s", cacheDir, session.ID)
	session.StreamURL = fmt.Sprintf("/media/%d/streams/%s/master.m3u8", session.MediaID, session.ID)

	if err := os.MkdirAll(session.TranscodePath, 0755); err != nil {
		log.Printf("[ERROR]: failed to create transcode path %s: %v\n", session.TranscodePath, err)
		return nil, err
	}

	return session, nil
}

func (f *FfmpegSession) Start() {
	f.Stop()
	f.lock.Lock()
	defer f.lock.Unlock()

	proc := exec.Command(
		"ffmpeg",
		"-ss", FormatTimestamp(f.CurrentPlaybackSecond),
		"-to", FormatTimestamp(f.StopPlaybackSecond),
		"-i", f.CurrentPath,
		"-preset", f.Preset,
		"-start_number", fmt.Sprint(f.CurrentPlaybackSecond/2),
		"-hls_playlist_type", "vod",
		"-force_key_frames", "expr:gte(t,n_forced*2.0000)",
		"-hls_time", "2",
		"-hls_list_size", "0",
		"-f", "hls",
		"-y", f.TranscodePath+"/master.m3u8",
	)

	stderrPipe, err := proc.StderrPipe()
	if err != nil {
		log.Printf("[ERROR]: failed to create stderr pipe: %v\n", err)
	}

	if err := proc.Start(); err != nil {
		log.Printf("[ERROR]: ffmpeg process start error: %v\n", err)
		return
	}
	f.Proc = proc

	if stderrPipe != nil {
		go f.trackStderr(stderrPipe)
	}
}

func (f *FfmpegSession) Stop() {
	f.lock.Lock()
	defer f.lock.Unlock()

	if f.Proc != nil && f.Proc.Process != nil {
		_ = f.Proc.Process.Kill()
	}
}

func (f *FfmpegSession) DoesSegmentExist(segment int64) bool {
	path := fmt.Sprintf("%s/master%d.ts", f.TranscodePath, segment)
	info, err := os.Stat(path)
	return err == nil && !info.IsDir() && info.Size() > 0
}

func (f *FfmpegSession) WaitForSegment(segment int64, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if f.DoesSegmentExist(segment) {
			return true
		}
		time.Sleep(50 * time.Millisecond)
	}
	return f.DoesSegmentExist(segment)
}

func (f *FfmpegSession) SkipTo(segmentNo int64) error {
	if f.SegmentBuffer == nil {
		return errors.New("segment buffer not initialized")
	}

	if f.SegmentBuffer.Lock != nil {
		f.SegmentBuffer.Lock.Lock()
		defer f.SegmentBuffer.Lock.Unlock()
	}

	if f.DoesSegmentExist(segmentNo) {
		return nil
	}

	f.lock.Lock()
	lastSeg := f.LastSegment
	f.lock.Unlock()

	if lastSeg+5 > int(segmentNo) {
		if f.WaitForSegment(segmentNo, 5*time.Second) {
			return nil
		}
	}

	if err := f.SegmentBuffer.AddNewSegment(segmentNo); err != nil {
		return err
	}

	segmentToSkipTo := f.SegmentBuffer.GetCurrentSegment(segmentNo)
	if segmentToSkipTo == nil {
		return errors.New("could not find segment range")
	}
	f.CurrentPlaybackSecond = segmentToSkipTo.StartSegment * 2
	f.StopPlaybackSecond = segmentToSkipTo.EndSegment * 2

	f.Start()

	if !f.WaitForSegment(segmentNo, 10*time.Second) {
		return fmt.Errorf("timeout waiting for segment %d", segmentNo)
	}

	return nil
}

func (f *FfmpegSession) trackStderr(r io.Reader) {
	re := regexp.MustCompile(`master([0-9]+)\.ts`)
	num := regexp.MustCompile(`[0-9]+`)
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		lineStr := scanner.Text()
		if match := re.FindString(lineStr); match != "" {
			if segStr := num.FindString(match); segStr != "" {
				if segVal, parseErr := strconv.Atoi(segStr); parseErr == nil {
					f.lock.Lock()
					f.LastSegment = segVal
					f.lock.Unlock()
				}
			}
		}
	}
}

func (f *FfmpegSession) TrackSegmentList() {
	// Preserved for backwards compatibility with interface callers
}
