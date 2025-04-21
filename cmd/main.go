package main

import (
	"cashback-app/config"
	_ "cashback-app/docs"
	"cashback-app/models"
	"cashback-app/routes"
	"cashback-app/worker"

	"github.com/gin-gonic/gin"
)

// @title           Cashback Service API
// @version         1.0
// @description     This cashback for buying tariffs
// @termsOfService  http://swagger.io/terms/

// @contact.name   @xumoyiddin_xolmuminov
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api

// @securityDefinitions.basic  BasicAuth

// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/
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
