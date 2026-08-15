package config

import (
	"os"
	"runtime"
)

type ServerInformation struct {
	Hostname        string `json:"hostname"`
	ServerVersion   string `json:"server_version"`
	OperatingSystem string `json:"operating_system"`
	FinishedWizard  bool   `json:"finished_wizard"`
}

func getServerInformation(c *Config) (ServerInformation, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return ServerInformation{}, err
	}
	return ServerInformation{
		Hostname:        hostname,
		ServerVersion:   os.Getenv("ocelot_version"),
		OperatingSystem: runtime.GOOS,
		FinishedWizard:  c.FinishedWizard,
	}, nil

}
