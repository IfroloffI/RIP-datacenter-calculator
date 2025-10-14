package main

import (
	"log"

	_ "datacenter-calc/docs"

	"datacenter-calc/config"
	"datacenter-calc/internal/app"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Datacenter Calculator API
// @version 1.0
// @description API для расчёта мощности ЦОД (лабораторная №4)
// @host localhost:8080
// @BasePath /api
// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	cfg, err := config.LoadConfig("./config")
	if err != nil {
		log.Fatal("Не удалось загрузить конфигурацию:", err)
	}

	app := app.New(cfg)

	r := app.GetEngine()
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	log.Println("Сервер запущен на http://localhost:8080")
	if err := app.Run(); err != nil {
		log.Fatal("Ошибка запуска сервера:", err)
	}
}
