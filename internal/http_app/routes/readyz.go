package routes

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func (vh VortexQHandler) ReadinessHandler(ctx *gin.Context) {
	if vh.IsShuttingDown.Load() {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{
			"message": "service is shutting down",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "service is ready",
	})
}
