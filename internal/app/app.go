package app

import (
	"datacenter-calc/config"
	"datacenter-calc/internal/db"
	appHttp "datacenter-calc/internal/delivery/http"
	"datacenter-calc/internal/repo"
	"datacenter-calc/internal/usecase"
	"log"

	"github.com/gin-gonic/gin"
)

type App struct {
	engine *gin.Engine
}

func New(cfg *config.Config) *App {
	dbConn, err := db.Connect(cfg)
	if err != nil {
		log.Fatal("Не удалось подключиться к БД:", err)
	}

	deviceRepo := &repo.DeviceRepository{DB: dbConn}
	calcRepo := &repo.CalculationRepository{DB: dbConn}

	calculator := usecase.NewPowerCalculator(deviceRepo, calcRepo)
	handler := appHttp.NewHandler(calculator, cfg)

	r := gin.Default()
	r.Static("/static", "./static")

	r.GET("/", handler.Devices)
	r.GET("/device/:id", handler.DeviceDetail)
	r.GET("/calculation/:id", handler.CalculationDetail)
	r.POST("/calculation/add", handler.AddToCalc)
	r.POST("/calculation/:id/add-device", handler.AddDeviceToCalculation)
	r.POST("/calculation/:id/delete", handler.DeleteCalculation)

	return &App{engine: r}
}

func (a *App) Run() error {
	return a.engine.Run(":8080")
}
