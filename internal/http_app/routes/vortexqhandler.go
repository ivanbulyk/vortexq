package routes

import (
	"github.com/ivanbulyk/vortexq/broker"
	"github.com/ivanbulyk/vortexq/internal/version"
	"github.com/prometheus/client_golang/prometheus"
	"log/slog"
	"sync/atomic"
)

type VortexQHandler struct {
	funcs          broker.VortexQFuncs
	IsShuttingDown *atomic.Bool
	CustomRegistry *prometheus.Registry
	Version        *version.Version
	Logger         *slog.Logger
}

// NewVortexQHandler creates a new VortexQHandler with the provided vortexQFuncs implementation.
func NewVortexQHandler(funcs broker.VortexQFuncs) *VortexQHandler {
	return &VortexQHandler{
		funcs:          funcs,
		IsShuttingDown: &atomic.Bool{},
		CustomRegistry: prometheus.NewRegistry(),
		Version:        version.NewVersion(),
		Logger:         slog.Default(),
	}
}
