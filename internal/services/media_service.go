package services

import (
	"context"
	"fmt"
	"math"
	"sync"

	"github.com/siddhantmadhur/pequod/internal/config"
	"github.com/siddhantmadhur/pequod/internal/domain"
	"github.com/siddhantmadhur/pequod/internal/repository"
	"github.com/siddhantmadhur/pequod/pkg/transcoder"
)

type MediaService interface {
	GetPlaybackInfo(ctx context.Context, mediaID int64, preset string, playbackSeconds int64, directPlay bool, userID int64) (*domain.PlaybackInfoResponse, error)
	GetMasterPlaylist(sessionID string) (string, error)
	GetStreamFile(sessionID string, segmentNo int64) (string, error)
	GetDirectPlayFilePath(ctx context.Context, mediaID int64) (string, error)
	GetAllSessions() []domain.StreamingSessionInfo
}

type mediaService struct {
	contentRepo repository.ContentRepository
	cfg         *config.Config
	sessions    map[string]*transcoder.FfmpegSession
	mu          sync.RWMutex
}

func NewMediaService(contentRepo repository.ContentRepository, cfg *config.Config) MediaService {
	return &mediaService{
		contentRepo: contentRepo,
		cfg:         cfg,
		sessions:    make(map[string]*transcoder.FfmpegSession),
	}
}

func (s *mediaService) GetPlaybackInfo(ctx context.Context, mediaID int64, preset string, playbackSeconds int64, directPlay bool, userID int64) (*domain.PlaybackInfoResponse, error) {
	content, err := s.contentRepo.GetByID(ctx, mediaID)
	if err != nil {
		return nil, domain.ErrNotFound
	}

	if directPlay {
		return &domain.PlaybackInfoResponse{
			MediaID:   fmt.Sprint(mediaID),
			StreamURL: fmt.Sprintf("/media/%d/direct/stream", mediaID),
			UserID:    fmt.Sprint(userID),
		}, nil
	}

	lengthOfFile, err := transcoder.GetLengthOfFile(content.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file duration: %w", err)
	}

	session, err := transcoder.NewFfmpegSession(preset, content.FilePath, playbackSeconds, s.cfg.CacheDir, mediaID)
	if err != nil {
		return nil, err
	}

	session.StopPlaybackSecond = int64(math.Ceil(lengthOfFile))
	session.SegmentBuffer = transcoder.NewSegment(session.CurrentPlaybackSecond/2, session.StopPlaybackSecond/2)

	s.mu.Lock()
	s.sessions[session.ID] = session
	s.mu.Unlock()

	go session.Start()
	for !session.DoesSegmentExist(playbackSeconds / 2) {
	}
	go session.TrackSegmentList()

	return &domain.PlaybackInfoResponse{
		SessionID: session.ID,
		MediaID:   fmt.Sprint(mediaID),
		Preset:    session.Preset,
		StreamURL: session.StreamURL,
		UserID:    fmt.Sprint(userID),
	}, nil
}

func (s *mediaService) GetMasterPlaylist(sessionID string) (string, error) {
	s.mu.RLock()
	session, exists := s.sessions[sessionID]
	s.mu.RUnlock()

	if !exists || session == nil {
		return "", domain.ErrSessionNotFound
	}

	return transcoder.CreatePlaylistHLSFile(session.MediaID, session.ID, session.CurrentPath)
}

func (s *mediaService) GetStreamFile(sessionID string, segmentNo int64) (string, error) {
	s.mu.RLock()
	session, exists := s.sessions[sessionID]
	s.mu.RUnlock()

	if !exists || session == nil {
		return "", domain.ErrSessionNotFound
	}

	if !session.DoesSegmentExist(segmentNo) {
		_ = session.SkipTo(segmentNo)
	}

	for !session.DoesSegmentExist(segmentNo) {
	}

	return fmt.Sprintf("%s/master%d.ts", session.TranscodePath, segmentNo), nil
}

func (s *mediaService) GetDirectPlayFilePath(ctx context.Context, mediaID int64) (string, error) {
	content, err := s.contentRepo.GetByID(ctx, mediaID)
	if err != nil {
		return "", domain.ErrNotFound
	}
	return content.FilePath, nil
}

func (s *mediaService) GetAllSessions() []domain.StreamingSessionInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []domain.StreamingSessionInfo
	for id, sess := range s.sessions {
		if sess != nil {
			result = append(result, domain.StreamingSessionInfo{
				SessionID: id,
				MediaID:   fmt.Sprint(sess.MediaID),
			})
		}
	}
	return result
}
