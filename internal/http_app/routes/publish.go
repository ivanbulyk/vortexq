package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/ivanbulyk/vortexq/broker"
	"github.com/ivanbulyk/vortexq/internal/logging"
	"log/slog"
	"net/http"
)

func (vh VortexQHandler) PublishHandler(ctx *gin.Context) {
	const op = "http_app.App.PublishHandler"
	message := broker.Message[any]{}

	if err := ctx.ShouldBindJSON(&message); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body", "error": err.Error()})
		return
	}

	// Perform the publish
	vh.funcs.Publish(message)
	vh.Logger.With(slog.String("op", op)).Info("published message", logging.Attr("message", message))

	// Send JSON response indicating success
	ctx.JSON(http.StatusOK, gin.H{"message": "message published", "data": message})
}
