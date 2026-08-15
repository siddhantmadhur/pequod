package media

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

	"github.com/siddhantmadhur/pequod/internal/config"

	"github.com/google/uuid"
)

type Ffmpeg struct {
	lock                  sync.Mutex
	Id                    string
	Proc                  *exec.Cmd
	Preset                string
	CurrentPlaybackSecond int64
	StopPlaybackSecond    int64
	CurrentPath           string
	TranscodePath         string
	MediaId               int64
	StreamUrl             string
	CurrentBuffer         *bytes.Buffer
	LastSegment           int
	SegmentBuffer         *Segment
	KillSignal            chan bool
	DirectPlay            bool
}

func NewFfmpeg(preset string, sourcePath string, currentPlaybackSecond int64, c *config.Config, mediaId int64) (*Ffmpeg, error) {
	if preset == "" {
		preset = "veryfast"
	}

	var ffmpeg = Ffmpeg{
		Id:                    uuid.NewString(),
		Preset:                preset,
		CurrentPlaybackSecond: currentPlaybackSecond,
		CurrentPath:           sourcePath,
		MediaId:               mediaId,
	}

	ffmpeg.TranscodePath = fmt.Sprintf("%s/%s", c.CacheDir, ffmpeg.Id)
	ffmpeg.StreamUrl = fmt.Sprintf("/media/%d/streams/%s/master.m3u8", ffmpeg.MediaId, ffmpeg.Id)
	ffmpeg.KillSignal = make(chan bool)

	err := os.MkdirAll(ffmpeg.TranscodePath, 0777)

	if err != nil {
		log.Printf("[ERROR]: %s\n", err.Error())
		return &ffmpeg, err
	}

	return &ffmpeg, nil
}

func (f *Ffmpeg) Start() {
	f.Stop()
	f.lock.Lock()
	defer f.lock.Unlock()
	// Add -c:v h264_videotoolbox for macos
	proc := exec.Command("ffmpeg", "-ss", getTimeStamp(f.CurrentPlaybackSecond), "-to", getTimeStamp(f.StopPlaybackSecond), "-i", f.CurrentPath, "-preset", f.Preset, "-start_number", fmt.Sprint(f.CurrentPlaybackSecond/2), "-hls_playlist_type", "vod", "-force_key_frames", "expr:gte(t,n_forced*2.0000)", "-hls_time", "2", "-hls_list_size", "0", "-f", "hls", "-y", f.TranscodePath+"/master.m3u8")

	if f.CurrentBuffer == nil {
		var b bytes.Buffer
		f.CurrentBuffer = &b
	}

	proc.Stderr = io.MultiWriter(f.CurrentBuffer)

	err := proc.Start()
	if err != nil {
		log.Printf("[ERROR]: %s\n", err.Error())
	}
	f.Proc = proc

}

func (f *Ffmpeg) Stop() {
	f.lock.Lock()
	defer f.lock.Unlock()

	if f.Proc != nil {
		f.Proc.Process.Kill()
	}

}

func (f *Ffmpeg) TrackSegmentList() {
	defer fmt.Printf("Go routine ended\n")
	re, _ := regexp.Compile("master[0-9]+.ts")
	num, _ := regexp.Compile("[0-9]+")
	line := ""
	for {
		data, err := f.CurrentBuffer.ReadByte()
		if err != nil {
			if io.EOF != err {
				panic(err)
			}
		}
		if data == '\n' {
			segmentFile := re.FindString(line)
			if segmentFile != "" {
				lastSegment := num.FindString(segmentFile)
				if lastSegment != "" {
					f.LastSegment, err = strconv.Atoi(lastSegment)
					fmt.Printf("Last segment: %d\n", f.LastSegment)
				}
			}
			line = ""
		} else {
			line += string(data)
		}
	}
}
