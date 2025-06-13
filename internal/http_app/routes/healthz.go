package routes

import (
	"github.com/gin-gonic/gin"
)

// LivenessHandler  is an HTTP handler that checks the liveness of the service.
func LivenessHandler(ctx *gin.Context) {
	ctx.JSON(200, gin.H{
		"status": "alive",
	})
}
