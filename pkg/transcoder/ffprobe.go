package transcoder

import (
	"encoding/json"
	"os/exec"
	"strconv"
	"strings"
)

type FFprobeResponse struct {
	Streams []Stream `json:"streams"`
	Format  Format   `json:"format"`
}

type Format struct {
	Filename   string `json:"filename"`
	NbStreams  int    `json:"nb_streams"`
	FormatName string `json:"format_name"`
	Duration   string `json:"duration"`
	Size       string `json:"size"`
	BitRate    string `json:"bit_rate"`
	ProbeScore int    `json:"probe_score"`
}

type Stream struct {
	Index       int    `json:"index"`
	CodecName   string `json:"codec_name"`
	Profile     string `json:"profile"`
	CodecType   string `json:"codec_type"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	CodedWidth  int    `json:"coded_width"`
	CodedHeight int    `json:"coded_height"`
	Duration    string `json:"duration"`
	BitRate     string `json:"bit_rate"`
}

func FFprobe(path string) (FFprobeResponse, error) {
	var response FFprobeResponse
	cmd := exec.Command("ffprobe", "-v", "quiet", "-print_format", "json", "-show_format", "-show_streams", path)
	output, err := cmd.Output()
	if err != nil {
		return response, err
	}
	err = json.Unmarshal(output, &response)
	return response, err
}

func GetLengthOfFile(path string) (float64, error) {
	probe := exec.Command("ffprobe", "-i", path, "-show_entries", "format=duration", "-v", "quiet", "-of", `csv=p=0`)
	output, err := probe.Output()
	if err != nil {
		return 0, err
	}
	size, err := strconv.ParseFloat(strings.ReplaceAll(string(output), "\n", ""), 64)
	return size, err
}
