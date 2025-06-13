package routes

import (
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
)

// IndexHandler serves as a health check
func (vh VortexQHandler) IndexHandler(ctx *gin.Context) {
	const op = "http_app.App.IndexHandler"

	info := vh.Version

	vh.Logger.With(slog.String("op", op)).Info("VortexQ Service Version", slog.String("project", info.Project), slog.String("commit", info.Commit),
		slog.String("build time", info.BuildTime), slog.String("release", info.Release))

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Successfully loaded VortexQ Service!",
		"info":    info,
	})
}
