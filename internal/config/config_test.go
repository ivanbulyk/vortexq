package config

import (
	"os"
	"testing"
)

// Test GetCombinedAddress concatenates host and port
func TestGetCombinedAddress(t *testing.T) {
	cfg := &ServerAppConfig{Host: "127.0.0.1", Port: "9090"}
	if got := cfg.GetCombinedAddress(); got != "127.0.0.1:9090" {
		t.Errorf("GetCombinedAddress() = %q; want %q", got, "127.0.0.1:9090")
	}
}

// Test LoadFromEnv sets defaults when environment variables are missing
func TestLoadFromEnvDefaults(t *testing.T) {
	// Unset all related env vars
	vars := []string{
		envServerServiceHost,
		envServerServicePort,
		envServerServiceLogLevel,
		envServerServiceProject,
		envServerServiceRelease,
		envServerServiceBuildTime,
		envServerServiceCommit,
	}
	for _, key := range vars {
		_ = os.Unsetenv(key)
	}

	cfg := &ServerAppConfig{}
	cfg.LoadFromEnv()
	// Check defaults
	if cfg.Host != "0.0.0.0" {
		t.Errorf("default Host = %q; want %q", cfg.Host, "0.0.0.0")
	}
	if cfg.Port != "8085" {
		t.Errorf("default Port = %q; want %q", cfg.Port, "8085")
	}
	if cfg.LogLevel != "local" {
		t.Errorf("default LogLevel = %q; want %q", cfg.LogLevel, "local")
	}
}

// Test LoadFromEnv respects provided environment variables
func TestLoadFromEnvOverrides(t *testing.T) {
	// Set env vars
	t.Setenv(envServerServiceHost, "1.2.3.4")
	t.Setenv(envServerServicePort, "9999")
	t.Setenv(envServerServiceLogLevel, "debug")
	t.Setenv(envServerServiceProject, "proj")
	t.Setenv(envServerServiceRelease, "rel")
	t.Setenv(envServerServiceBuildTime, "bt")
	t.Setenv(envServerServiceCommit, "cm")

	cfg := &ServerAppConfig{}
	cfg.LoadFromEnv()
	if cfg.Host != "1.2.3.4" {
		t.Errorf("Host override = %q; want %q", cfg.Host, "1.2.3.4")
	}
	if cfg.Port != "9999" {
		t.Errorf("Port override = %q; want %q", cfg.Port, "9999")
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("LogLevel override = %q; want %q", cfg.LogLevel, "debug")
	}
	if cfg.Project != "proj" {
		t.Errorf("Project override = %q; want %q", cfg.Project, "proj")
	}
	if cfg.Release != "rel" {
		t.Errorf("Release override = %q; want %q", cfg.Release, "rel")
	}
	if cfg.BuildTime != "bt" {
		t.Errorf("BuildTime override = %q; want %q", cfg.BuildTime, "bt")
	}
	if cfg.Commit != "cm" {
		t.Errorf("Commit override = %q; want %q", cfg.Commit, "cm")
	}
}
