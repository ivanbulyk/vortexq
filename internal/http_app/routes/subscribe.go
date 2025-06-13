package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/ivanbulyk/vortexq/broker"
	"github.com/ivanbulyk/vortexq/internal/logging"
	"log/slog"
	"net/http"
)

func (vh VortexQHandler) SubscribeHandler(ctx *gin.Context) {
	const op = "http_app.App.SubscribeHandler"
	// Unmarshal the JSON request body into the Subscription struct
	var subscription broker.Subscription

	if err := ctx.ShouldBindJSON(&subscription); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "invalid request body", "error": err.Error()})
		return
	}

	if err := vh.funcs.Subscribe(subscription); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "failed to subscribe", "error": err.Error()})
		return
	}

	vh.Logger.With(slog.String("op", op)).Info("subscription received", logging.Attr("subscription", subscription))
	ctx.JSON(http.StatusOK, gin.H{"message": "subscription processed successfully"})

}
