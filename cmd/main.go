package main

import (
	appHttp "datacenter-calc/internal/delivery/http"
	"datacenter-calc/internal/repo"
	"datacenter-calc/internal/usecase"
	"log"
	"net/http"
)

func main() {
	deviceRepo := &repo.DeviceRepository{}
	calculator := usecase.NewPowerCalculator(deviceRepo)
	handler := appHttp.NewHandler(calculator)

	mux := http.NewServeMux()

	mux.HandleFunc("/", handler.Devices)
	mux.HandleFunc("/device/", handler.DeviceDetail)
	mux.HandleFunc("/power-calc", handler.PowerCalc)

	mux.Handle("/static/", http.StripPrefix("/static/", handler.ServeStatic()))

	log.Println("Сервер запущен на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
