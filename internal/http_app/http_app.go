package http_app

import (
	"context"
	"errors"
	"fmt"
	"github.com/ivanbulyk/vortexq/internal/logging"
	"log/slog"
	"net/http"
)

type App struct {
	log        *slog.Logger
	httpServer *http.Server
}

// New creates new http server app.
func New(log *slog.Logger, httpServer *http.Server) *App {

	return &App{
		log:        log,
		httpServer: httpServer,
	}
}

// MustRun runs HTTP server and panics if any error occurs.
func (a *App) MustRun() {
	if err := a.run(); err != nil {
		panic(err)
	}
}

// Run runs HTTP server.
func (a *App) run() error {
	const op = "http_app.App.run"

	a.log.With(slog.String("op", op)).Info("server listening at ", slog.String("addr", a.httpServer.Addr))
	if err := a.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		a.log.With(slog.String("op", op)).Error("failed to run server: \n", logging.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// Stop stops HTTP server.
func (a *App) Stop(timeoutCtx context.Context) error {
	const op = "http_app.App.Stop"

	a.log.With(slog.String("op", op)).
		Info("server shutdown")

	err := a.httpServer.Shutdown(timeoutCtx)
	if err != nil {
		return err
	}
	return nil
}
