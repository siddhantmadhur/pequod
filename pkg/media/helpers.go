package media

import (
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func GetLengthOfFile(path string) (float64, error) {
	probe := exec.Command("ffprobe", "-i", path, "-show_entries", "format=duration", "-v", "quiet", "-of", `csv=p=0`)
	output, err := probe.Output()
	if err != nil {
		return 0, err
	}
	size, err := strconv.ParseFloat(strings.ReplaceAll(string(output), "\n", ""), 64)
	return size, err
}

// Returns content of m3u8 file as a string or error
func CreatePlaylistHLSFile(session *Ffmpeg) (string, error) {
	if session == nil {
		return "", errors.New("Session not found")
	}
	size, err := GetLengthOfFile(session.CurrentPath)
	counter := size
	if err != nil {
		return "", err
	}
	content := "#EXTM3U\n"
	content += "#EXT-X-VERSION:3\n"
	content += "#EXT-X-TARGETDURATION:2\n"
	content += "#EXT-X-MEDIA-SEQUENCE:0\n"
	content += "#EXT-X-PLAYLIST-TYPE:VOD\n"

	idx := 0
	for counter > 0.0 {
		newTime := 0.0
		if counter >= 2.002 {
			newTime = 2.002
		} else {
			newTime = counter
		}

		///media/:mediaId/streams/:streamId/master.m3u8
		content += fmt.Sprintf("#EXTINF:%.6f,\n", newTime)
		content += fmt.Sprintf("http://localhost:8080/media/%d/streams/%s/%d/stream.ts\n", session.MediaId, session.Id, idx)
		counter -= newTime
		idx += 1
	}

	content += "#EXT-X-ENDLIST"
	return content, nil
}

func getTimeStamp(timestamp int64) string {
	currentTime := time.Duration(timestamp) * time.Second
	hours := math.Floor(currentTime.Hours())
	minutes := math.Floor(currentTime.Minutes()) - (hours * 60)
	seconds := currentTime.Seconds() - (hours * 3600) - (minutes * 60)
	formatTimestamp := fmt.Sprintf("%.2d:%.2d:%.2d", int(hours), int(minutes), int(seconds))

	return formatTimestamp
}

func doesSegmentExist(session *Ffmpeg, segment int64) bool {
	path := fmt.Sprintf("%s/master%d.ts", session.TranscodePath, segment)
	_, err := os.ReadFile(path)
	return err == nil
}
