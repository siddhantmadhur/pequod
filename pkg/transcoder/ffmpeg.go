package transcoder

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"sync"

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
	CurrentBuffer         *bytes.Buffer
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

	if err := os.MkdirAll(session.TranscodePath, 0777); err != nil {
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

	if f.CurrentBuffer == nil {
		var b bytes.Buffer
		f.CurrentBuffer = &b
	}

	proc.Stderr = io.MultiWriter(f.CurrentBuffer)

	if err := proc.Start(); err != nil {
		log.Printf("[ERROR]: ffmpeg process start error: %v\n", err)
	}
	f.Proc = proc
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
	_, err := os.ReadFile(path)
	return err == nil
}

func (f *FfmpegSession) SkipTo(segmentNo int64) error {
	f.SegmentBuffer.Lock.Lock()

	if f.DoesSegmentExist(segmentNo) {
		f.SegmentBuffer.Lock.Unlock()
		return nil
	}

	if f.LastSegment+5 > int(segmentNo) {
		for !f.DoesSegmentExist(segmentNo) {
		}
		f.SegmentBuffer.Lock.Unlock()
		return nil
	}

	if err := f.SegmentBuffer.AddNewSegment(segmentNo); err != nil {
		f.SegmentBuffer.Lock.Unlock()
		return err
	}

	segmentToSkipTo := f.SegmentBuffer.GetCurrentSegment(segmentNo)
	f.CurrentPlaybackSecond = segmentToSkipTo.StartSegment * 2
	f.StopPlaybackSecond = segmentToSkipTo.EndSegment * 2

	f.Start()
	f.SegmentBuffer.Lock.Unlock()

	for !f.DoesSegmentExist(segmentNo) {
	}

	for f.LastSegment < int(segmentNo) {
	}

	return nil
}

func (f *FfmpegSession) TrackSegmentList() {
	re := regexp.MustCompile(`master([0-9]+)\.ts`)
	num := regexp.MustCompile(`[0-9]+`)
	var line []byte

	for {
		b, err := f.CurrentBuffer.ReadByte()
		if err != nil {
			if err == io.EOF {
				return
			}
			return
		}
		if b == '\n' {
			lineStr := string(line)
			if match := re.FindString(lineStr); match != "" {
				if segStr := num.FindString(match); segStr != "" {
					if segVal, parseErr := strconv.Atoi(segStr); parseErr == nil {
						f.LastSegment = segVal
					}
				}
			}
			line = line[:0]
		} else {
			line = append(line, b)
		}
	}
}
