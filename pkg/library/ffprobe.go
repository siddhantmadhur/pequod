package library

import (
	"encoding/json"
	"os/exec"
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
	ffprobe := exec.Command("ffprobe", "-v", "quiet", "-print_format", "json", "-show_format", "-show_streams", path)
	output, err := ffprobe.Output()
	if err != nil {
		return response, err
	}
	err = json.Unmarshal(output, &response)
	return response, err
}
