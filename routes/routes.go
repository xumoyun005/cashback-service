package routes

import "github.com/gin-gonic/gin"
import "cashback-app/handler"

func RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api")
	{
		api.POST("/cashback", handler.HandleCashback)
		api.POST("/cashback/decrease", handler.HandleCashbackDecrease)
		api.GET("/cashback/:id", handler.GetCashbackByCineramaId)
		api.GET("/cashback_history/:id", handler.GetCashbackHistoryByCineramaId)
	}
}
