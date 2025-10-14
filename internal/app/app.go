package app

import (
	"datacenter-calc/config"
	"datacenter-calc/internal/auth"
	"datacenter-calc/internal/db"
	"datacenter-calc/internal/delivery/http"
	"datacenter-calc/internal/minio"
	"datacenter-calc/internal/model"
	"datacenter-calc/internal/redis"
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

	redisClient := redis.NewRedisClient(cfg.Redis.Host, cfg.Redis.Port, cfg.Redis.Password)

	deviceRepo := &repo.DeviceRepository{DB: dbConn}
	calcRepo := &repo.CalculationRepository{DB: dbConn}
	userRepo := &repo.UserRepo{DB: dbConn}

	calculator := usecase.NewPowerCalculator(deviceRepo, calcRepo, minioClient)
	userUsecase := usecase.NewUserUsecase(userRepo, redisClient, cfg.JWT.Exp)

	handler := http.NewHandler(calculator, userUsecase, cfg)

	r := gin.Default()

	api := r.Group("/api")
	{
		devicesPublic := api.Group("/devices")
		{
			devicesPublic.GET("", handler.DeviceHandler.GetDevices)
		}

		usersPublic := api.Group("/users")
		{
			usersPublic.POST("/register", handler.UserHandler.RegisterUser)
			usersPublic.POST("/login", handler.UserHandler.LoginUser)
		}
	}

	authed := api.Group("")
	authed.Use(auth.AuthMiddleware(redisClient, model.RoleUser, model.RoleModerator))
	{
		authed.GET("/devices/:id", handler.DeviceHandler.GetDevice)

		authed.GET("/cart", handler.CartHandler.GetCartInfo)

		calc := authed.Group("/power-calculations")
		{
			calc.GET("", handler.CalculationHandler.GetCalculations)
			calc.GET("/:id", handler.CalculationHandler.GetCalculation)
			calc.PUT("/:id", handler.CalculationHandler.UpdateCalculationFields)
			calc.PUT("/:id/form", handler.CalculationHandler.FormCalculation)
			calc.DELETE("/:id", handler.CalculationHandler.DeleteCalculation)

			calc.POST("/:id/devices", handler.MMHandler.AddDeviceToCalculation)
			calc.PUT("/:id/devices/:device_id", handler.MMHandler.UpdateDeviceInCalculation)
			calc.DELETE("/:id/devices/:device_id", handler.MMHandler.RemoveDeviceFromCalculation)
		}

		users := authed.Group("/users")
		{
			users.GET("/me", handler.UserHandler.GetMe)
			users.PUT("/me", handler.UserHandler.UpdateMe)
			users.POST("/logout", handler.UserHandler.LogoutUser)
		}
	}

	moderator := api.Group("")
	moderator.Use(auth.AuthMiddleware(redisClient, model.RoleModerator))
	{
		devices := moderator.Group("/devices")
		{
			devices.POST("", handler.DeviceHandler.CreateDevice)
			devices.PUT("/:id", handler.DeviceHandler.UpdateDevice)
			devices.DELETE("/:id", handler.DeviceHandler.DeleteDevice)
			devices.POST("/:id/image", handler.DeviceHandler.UploadDeviceImage)
		}

		calc := moderator.Group("/power-calculations")
		{
			calc.PUT("/:id/complete", handler.CalculationHandler.CompleteCalculation)
			calc.PUT("/:id/reject", handler.CalculationHandler.RejectCalculation)
		}
	}

	return &App{engine: r}
}

func (a *App) Run() error {
	return a.engine.Run(":8080")
}
