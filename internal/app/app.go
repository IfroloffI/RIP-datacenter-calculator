package app

import (
	appHttp "datacenter-calc/internal/delivery/http"
	"datacenter-calc/internal/repo"
	"datacenter-calc/internal/usecase"
	"net/http"
)

type App struct {
	httpServer *http.Server
}

func New() *App {
	deviceRepo := &repo.DeviceRepository{}
	orderRepo := &repo.OrderRepository{}

	calculator := usecase.NewPowerCalculator(deviceRepo, orderRepo)
	handler := appHttp.NewHandler(calculator)

	// Routes
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler.Devices)
	mux.HandleFunc("/device/", handler.DeviceDetail)
	mux.HandleFunc("/power-calc", handler.PowerCalc)
	mux.Handle("/static/", http.StripPrefix("/static/", handler.ServeStatic()))

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	return &App{
		httpServer: server,
	}
}

func (a *App) Run() error {
	return a.httpServer.ListenAndServe()
}
