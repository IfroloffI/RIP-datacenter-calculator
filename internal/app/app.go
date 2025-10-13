package app

import (
	"datacenter-calc/config"
	"datacenter-calc/internal/db"
	"datacenter-calc/internal/delivery/http"
	"datacenter-calc/internal/minio"
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

	// MinIO
	minioClient, err := minio.NewMinIOClient(
		cfg.MinIO.Endpoint,
		cfg.MinIO.AccessKey,
		cfg.MinIO.SecretKey,
		cfg.MinIO.Bucket,
		cfg.MinIO.UseSSL,
	)
	if err != nil {
		log.Fatal("Не удалось подключиться к MinIO:", err)
	}

	deviceRepo := &repo.DeviceRepository{DB: dbConn}
	calcRepo := &repo.CalculationRepository{DB: dbConn}

	calculator := usecase.NewPowerCalculator(deviceRepo, calcRepo, minioClient)
	handler := http.NewHandler(calculator, cfg)

	r := gin.Default()
	// r.Static("/static", "./static")

	api := r.Group("/api")
	{
		devices := api.Group("/devices")
		{
			devices.GET("", handler.DeviceHandler.GetDevices)
			devices.POST("", handler.DeviceHandler.CreateDevice)
			devices.GET("/:id", handler.DeviceHandler.GetDevice)
			devices.PUT("/:id", handler.DeviceHandler.UpdateDevice)
			devices.DELETE("/:id", handler.DeviceHandler.DeleteDevice)
			devices.POST("/:id/image", handler.DeviceHandler.UploadDeviceImage)
		}

		calc := api.Group("/power-calculations")
		{
			calc.GET("", handler.CalculationHandler.GetCalculations)
			// calc.POST("", handler.CalculationHandler.CreateCalculation)
			calc.GET("/:id", handler.CalculationHandler.GetCalculation)
			calc.PUT("/:id", handler.CalculationHandler.UpdateCalculationFields)
			calc.PUT("/:id/form", handler.CalculationHandler.FormCalculation)
			calc.PUT("/:id/complete", handler.CalculationHandler.CompleteCalculation)
			calc.PUT("/:id/reject", handler.CalculationHandler.RejectCalculation)
			calc.DELETE("/:id", handler.CalculationHandler.DeleteCalculation)
		}

		api.GET("/cart", handler.CartHandler.GetCartInfo)

		calc.POST("/:id/devices", handler.MMHandler.AddDeviceToCalculation)
		calc.PUT("/:id/devices/:device_id", handler.MMHandler.UpdateDeviceInCalculation)
		calc.DELETE("/:id/devices/:device_id", handler.MMHandler.RemoveDeviceFromCalculation)

		users := api.Group("/users")
		{
			users.POST("/register", handler.UserHandler.RegisterUser)
			users.POST("/login", handler.UserHandler.LoginUser)
			users.POST("/logout", handler.UserHandler.LogoutUser)
			users.GET("/me", handler.UserHandler.GetMe)
			users.PUT("/me", handler.UserHandler.UpdateMe)
		}
	}

	return &App{engine: r}
}

func (a *App) Run() error {
	return a.engine.Run(":8080")
}
