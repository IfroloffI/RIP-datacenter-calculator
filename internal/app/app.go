package app

import (
	"datacenter-calc/internal/db"
	appHttp "datacenter-calc/internal/delivery/http"
	"datacenter-calc/internal/repo"
	"datacenter-calc/internal/usecase"
	"log"
	"net/http"
)

type App struct {
	httpServer *http.Server
}

func New() *App {
	dbConn, err := db.Connect()
	if err != nil {
		log.Fatal("Не удалось подключиться к БД:", err)
	}

	deviceRepo := &repo.DeviceRepository{DB: dbConn}
	orderRepo := &repo.OrderRepository{DB: dbConn}

	calculator := usecase.NewPowerCalculator(deviceRepo, orderRepo)
	handler := appHttp.NewHandler(calculator)

	mux := http.NewServeMux()
	mux.HandleFunc("/", handler.Devices)
	mux.HandleFunc("/device/", handler.DeviceDetail)
	mux.HandleFunc("/power-calc", handler.PowerCalc)
	mux.HandleFunc("/order/add-device", handler.AddDeviceToOrder)
	mux.HandleFunc("/order/delete", handler.DeleteOrder)
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
