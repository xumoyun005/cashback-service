package main

import (
	"cashback-app/config"
	"cashback-app/models"
	"cashback-app/routes"
	"cashback-app/worker"

	"github.com/gin-gonic/gin"
)

func main() {
	config.LoadEnv()
	config.ConnectDb()
	config.DB.AutoMigrate(&models.Cashback{}, &models.CashbackHistory{})
	go worker.StartCashbackWorker()
	go worker.StartDescreaseCashbackWorker()
	r := gin.Default()
	routes.RegisterRoutes(r)
	r.Run(":8080")
}
