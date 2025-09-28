package main

import (
	"log"

	"datacenter-calc/config"
	"datacenter-calc/internal/app"
)

func main() {
	cfg, err := config.LoadConfig("./config")
	if err != nil {
		log.Fatal("Не удалось загрузить конфигурацию:", err)
	}

	app := app.New(cfg)

	log.Println("Сервер запущен на http://localhost:8080")
	if err := app.Run(); err != nil {
		log.Fatal("Ошибка запуска сервера:", err)
	}
}
