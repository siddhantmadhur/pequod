package transcoder

import (
	"errors"
	"fmt"
	"math"
	"time"
)

// CreatePlaylistHLSFile generates standard HLS master playlist text based on total file duration.
func CreatePlaylistHLSFile(mediaID int64, sessionID string, sourcePath string) (string, error) {
	if sourcePath == "" {
		return "", errors.New("source path is required")
	}
	size, err := GetLengthOfFile(sourcePath)
	if err != nil {
		return "", err
	}

	counter := size
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

		content += fmt.Sprintf("#EXTINF:%.6f,\n", newTime)
		content += fmt.Sprintf("http://localhost:8080/media/%d/streams/%s/%d/stream.ts\n", mediaID, sessionID, idx)
		counter -= newTime
		idx++
	}

	content += "#EXT-X-ENDLIST"
	return content, nil
}

func FormatTimestamp(seconds int64) string {
	d := time.Duration(seconds) * time.Second
	hours := math.Floor(d.Hours())
	minutes := math.Floor(d.Minutes()) - (hours * 60)
	secs := d.Seconds() - (hours * 3600) - (minutes * 60)
	return fmt.Sprintf("%.2d:%.2d:%.2d", int(hours), int(minutes), int(secs))
}
