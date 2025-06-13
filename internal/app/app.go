package app

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/ivanbulyk/vortexq/broker"
	"github.com/ivanbulyk/vortexq/internal/config"
	"github.com/ivanbulyk/vortexq/internal/http_app"
	"github.com/ivanbulyk/vortexq/internal/http_app/routes"
	"github.com/ivanbulyk/vortexq/internal/logging"
	"github.com/ivanbulyk/vortexq/internal/version"
	"net"
	"net/http"
	"os"

	"golang.org/x/sync/errgroup"
	"log/slog"
	"os/signal"
	"syscall"
	"time"
)

const (
	_shutdownPeriod      = 15 * time.Second
	_shutdownHardPeriod  = 3 * time.Second
	_readinessDrainDelay = 5 * time.Second
)

type App struct {
	HTTPApp *http_app.App
}

// New returns an App instance.
func New(log *slog.Logger, httpServer *http.Server) *App {

	httpApp := http_app.New(log, httpServer)
	return &App{
		HTTPApp: httpApp,
	}
}

// MustRun is wrapper around run() and it panics if any error occurs.
func MustRun() {
	if err := run(); err != nil {
		panic(err)
	}
}

// Run runs App instance.
func run() error {
	const op = "app.run"

	// Setup signal context
	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := &config.ServerAppConfig{}
	cfg.LoadFromEnv()
	log := logging.SetupLogger(cfg.LogLevel)

	router := gin.Default()
	router.Use(gin.Logger())
	router.Use(routes.RequestMetricsMiddleware())

	// Initialize the broker
	vq := broker.NewVortexQ[any]()
	vq.Logger = log

	// Set up the VortexQ handler
	vortexqHandler := routes.NewVortexQHandler(vq)
	vortexqHandler.Logger = log
	vortexqHandler.Version = &version.Version{
		Project:   cfg.Project,
		BuildTime: cfg.BuildTime,
		Commit:    cfg.Commit,
		Release:   cfg.Release,
	}

	// Register custom metrics
	vortexqHandler.CustomRegistry.MustRegister(routes.HttpRequestTotal, routes.HttpRequestErrorTotal)

	// Set up routes
	SetUpRoutes(router, vortexqHandler)

	// Ensure in-flight requests aren't canceled immediately on SIGTERM
	ongoingCtx, stopOngoingGracefully := context.WithCancel(context.Background())

	server := &http.Server{
		Addr: cfg.Host + ":" + cfg.Port,
		BaseContext: func(_ net.Listener) context.Context {
			return ongoingCtx
		},
		Handler: router.Handler(),
	}

	application := New(log, server)

	g, ctx := errgroup.WithContext(ongoingCtx)

	g.Go(func() error {
		log.With(slog.String("op", op)).
			Info("server starting ..", slog.String("project:", cfg.Project), slog.String("commit:", cfg.Commit),
				slog.String("build time:", cfg.BuildTime), slog.String("release:", cfg.Release))
		application.HTTPApp.MustRun()
		return ctx.Err()
	})
	g.Go(func() error {
		for {
			select {
			case <-time.Tick(time.Second):
				err := vq.Swirl()
				if err != nil {
					log.With(slog.String("op", op)).Error("failed to swirl messages", logging.Err(err))
				}

			case <-ctx.Done():
				return ctx.Err()
			}
		}
	})

	// Wait for a signal
	<-rootCtx.Done()
	stop()
	vortexqHandler.IsShuttingDown.Store(true)
	log.With(slog.String("op", op)).Info("received shutdown signal, shutting down..")

	// Give time for readiness check to propagate
	time.Sleep(_readinessDrainDelay)
	log.With(slog.String("op", op)).Info("readiness check propagated, now waiting for ongoing requests to finish..")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), _shutdownPeriod)
	defer cancel()
	err := application.HTTPApp.Stop(shutdownCtx)
	stopOngoingGracefully()
	if err != nil {
		log.Error("failed to wait for ongoing requests to finish, waiting for forced cancellation", logging.Err(err))
		time.Sleep(_shutdownHardPeriod)
	}

	log.With(slog.String("op", op)).Info("server shut down gracefully")

	// wait for shutdown
	if errWait := g.Wait(); errWait != nil {
		log.With(slog.String("op", op)).Error("error during shutdown: ", logging.Err(errWait))
		os.Exit(1)
	}
	return nil
}

func SetUpRoutes(router *gin.Engine, vortexqHandler *routes.VortexQHandler) {

	router.GET("/", vortexqHandler.IndexHandler)
	router.POST("/publish", vortexqHandler.PublishHandler)
	router.POST("/subscribe", vortexqHandler.SubscribeHandler)
	router.GET("/healthz", routes.LivenessHandler)
	router.GET("/readyz", vortexqHandler.ReadinessHandler)
	router.GET("/metrics", vortexqHandler.PrometheusHandler())
}
