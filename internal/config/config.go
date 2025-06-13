package config

import (
	"fmt"
	"github.com/ivanbulyk/vortexq/internal/version"
	"os"
)

const (
	envServerServiceHost      = "SERVER_SERVICE_HOST"
	envServerServicePort      = "SERVER_SERVICE_PORT"
	envServerServiceLogLevel  = "SERVER_SERVICE_LOG_LEVEL"
	envServerServiceProject   = "SERVER_SERVICE_PROJECT"
	envServerServiceRelease   = "SERVER_SERVICE_RELEASE"
	envServerServiceBuildTime = "SERVER_SERVICE_BUILD_TIME"
	envServerServiceCommit    = "SERVER_SERVICE_COMMIT"
)

// ServerAppConfig ...
type ServerAppConfig struct {
	Host      string
	Port      string
	LogLevel  string
	Project   string
	Release   string
	BuildTime string
	Commit    string
}

// GetCombinedAddress with Host and Port
func (cfg *ServerAppConfig) GetCombinedAddress() string {
	return fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
}

// LoadFromEnv form environment variables
func (cfg *ServerAppConfig) LoadFromEnv() {
	cfg.Host = os.Getenv(envServerServiceHost)
	if len(cfg.Host) == 0 {
		cfg.Host = "0.0.0.0"
	}
	cfg.Port = os.Getenv(envServerServicePort)
	if len(cfg.Port) == 0 {
		cfg.Port = "8085"
	}
	cfg.LogLevel = os.Getenv(envServerServiceLogLevel)
	if len(cfg.LogLevel) == 0 {
		cfg.LogLevel = "local"
	}
	cfg.Project = os.Getenv(envServerServiceProject)
	if len(cfg.Project) == 0 {
		cfg.Project = version.Project
	}
	cfg.Release = os.Getenv(envServerServiceRelease)
	if len(cfg.Release) == 0 {
		cfg.Release = version.Release
	}
	cfg.BuildTime = os.Getenv(envServerServiceBuildTime)
	if len(cfg.BuildTime) == 0 {
		cfg.BuildTime = version.BuildTime
	}
	cfg.Commit = os.Getenv(envServerServiceCommit)
	if len(cfg.Commit) == 0 {
		cfg.Commit = version.Commit
	}

}
