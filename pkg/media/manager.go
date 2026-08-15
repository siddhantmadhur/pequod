package media

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"sync"

	"github.com/siddhantmadhur/pequod/internal/auth"
	"github.com/siddhantmadhur/pequod/internal/config"
	"github.com/siddhantmadhur/pequod/internal/database"

	"github.com/labstack/echo/v4"
)

type Manager struct {
	Sessions map[string]*Ffmpeg
	Config   *config.Config
}

func NewManager(cfg *config.Config) (*Manager, error) {
	var m Manager
	m.Config = cfg
	m.Sessions = make(map[string]*Ffmpeg)

	return &m, nil
}

// /media/:mediaId/playback/info
func (m *Manager) GetPlaybackInfo(c echo.Context, u *auth.User, cfg *config.Config) error {

	var request struct {
		Preset          string `json:"preset"`
		PlaybackSeconds int64  `json:"playback_seconds"`
		DirectPlay      bool   `json:"direct_play"`
	}

	c.Bind(&request)

	mediaId, err := strconv.Atoi(c.Param("mediaId"))
	if err != nil {
		return c.String(500, err.Error())
	}

	conn, queries, err := database.GetConn(cfg)
	if err != nil {
		return c.String(500, err.Error())
	}
	defer conn.Close()

	content, err := queries.GetContentInfo(context.Background(), int64(mediaId))
	if err != nil {
		return c.String(500, err.Error())
	}
	if request.DirectPlay {
		var response = map[string]string{
			"media_id":   fmt.Sprint(mediaId),
			"stream_url": fmt.Sprintf("/media/%d/direct/stream", mediaId),
			"user_id":    fmt.Sprint(u.UID),
		}

		return c.JSON(200, response)
	}

	lengthOfFile, err := GetLengthOfFile(content.FilePath)
	if err != nil {
		return c.String(500, err.Error())
	}

	ffmpeg, err := NewFfmpeg(request.Preset, content.FilePath, request.PlaybackSeconds, m.Config, int64(mediaId))
	if err != nil {
		return c.String(500, err.Error())
	}

	ffmpeg.StopPlaybackSecond = int64(math.Ceil(lengthOfFile))
	var seg Segment
	seg.StartSegment = ffmpeg.CurrentPlaybackSecond / 2
	seg.EndSegment = ffmpeg.StopPlaybackSecond / 2
	var lock sync.Mutex
	seg.Lock = &lock
	ffmpeg.SegmentBuffer = &seg

	m.Sessions[ffmpeg.Id] = ffmpeg

	var response = map[string]string{
		"session_id": ffmpeg.Id,
		"media_id":   fmt.Sprint(mediaId),
		"preset":     ffmpeg.Preset,
		"stream_url": ffmpeg.StreamUrl,
		"user_id":    fmt.Sprint(u.UID),
	}

	go ffmpeg.Start()
	for !doesSegmentExist(ffmpeg, request.PlaybackSeconds/2) {
	}
	go ffmpeg.TrackSegmentList()
	return c.JSON(201, response)
}

// /media/:mediaId/streams/:streamId/master.m3u8
func (m *Manager) GetMasterPlaylist(c echo.Context) error {
	sessionId := c.Param("sessionId")
	session := m.Sessions[sessionId]
	if session == nil {
		return c.String(500, "Session not found")
	}

	playlist, err := CreatePlaylistHLSFile(session)
	if err != nil {
		return c.String(500, err.Error())
	}

	return c.String(200, playlist)
}

// /media/:mediaId/streams/:sessionId/:segment/stream.ts
func (m *Manager) GetStreamFile(c echo.Context) error {
	sessionId := c.Param("sessionId")
	segment := c.Param("segment")
	session := m.Sessions[sessionId]
	path := fmt.Sprintf("%s/master%s.ts", session.TranscodePath, segment)
	if session == nil {
		return c.String(500, "Session does not exist")
	}
	segmentNo, err := strconv.Atoi(segment)
	if err != nil {
		return c.String(500, err.Error())
	}
	if !doesSegmentExist(session, int64(segmentNo)) {
		session.SkipTo(int64(segmentNo))
	}

	for !doesSegmentExist(session, int64(segmentNo)) {
	}

	return c.File(path)
}

// /media/:mediaId/direct/master.mp4?
func (m *Manager) GetDirectPlayVideo(c echo.Context, cfg *config.Config) error {

	mediaId, err := strconv.Atoi(c.Param("mediaId"))
	if err != nil {
		return c.String(500, err.Error())
	}

	conn, queries, err := database.GetConn(cfg)
	if err != nil {
		return c.String(500, err.Error())
	}
	defer conn.Close()

	content, err := queries.GetContentInfo(context.Background(), int64(mediaId))
	if err != nil {
		return c.String(500, err.Error())
	}

	return c.File(content.FilePath)
}
