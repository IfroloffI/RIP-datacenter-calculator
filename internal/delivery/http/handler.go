package http

import (
	"datacenter-calc/config"
	"datacenter-calc/internal/usecase"
)

const CurrentUserID = 1 // hardcode user

type Handler struct {
	DeviceHandler      *DeviceHandler
	CalculationHandler *CalculationHandler
	CartHandler        *CartHandler
	MMHandler          *MMHandler
	UserHandler        *UserHandler
}

func NewHandler(calculator *usecase.PowerCalculator, cfg *config.Config) *Handler {
	minioURL := cfg.MinIO.URL + "/" + cfg.MinIO.Bucket

	return &Handler{
		DeviceHandler:      NewDeviceHandler(calculator.DeviceRepo, calculator.MinIOClient, minioURL),
		CalculationHandler: NewCalculationHandler(calculator, minioURL),
		CartHandler:        NewCartHandler(calculator.CalculationRepo),
		MMHandler:          NewMMHandler(calculator.CalculationRepo),
		UserHandler:        NewUserHandler(), // или отдельный UserRepo
	}
}
