package services

import (
	"os"
	"runtime"

	"github.com/siddhantmadhur/pequod/internal/config"
	"github.com/siddhantmadhur/pequod/internal/domain"
)

type ServerService interface {
	GetServerInfo() (*domain.ServerInformation, error)
	FinishWizard() error
	IsWizardCompleted() bool
}

type serverService struct {
	cfg *config.Config
}

func NewServerService(cfg *config.Config) ServerService {
	return &serverService{
		cfg: cfg,
	}
}

func (s *serverService) GetServerInfo() (*domain.ServerInformation, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	return &domain.ServerInformation{
		Hostname:        hostname,
		ServerVersion:   os.Getenv("ocelot_version"),
		OperatingSystem: runtime.GOOS,
		FinishedWizard:  s.cfg.FinishedWizard,
	}, nil
}

func (s *serverService) FinishWizard() error {
	s.cfg.Mutex.Lock()
	defer s.cfg.Mutex.Unlock()

	s.cfg.FinishedWizard = true
	return s.cfg.Write()
}

func (s *serverService) IsWizardCompleted() bool {
	return s.cfg.FinishedWizard
}
