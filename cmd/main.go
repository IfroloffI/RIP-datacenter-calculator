package main

import (
	"log"

	"datacenter-calc/internal/app"
)

func main() {
	app := app.New()

	log.Println("Сервер запущен на http://localhost:8080")
	if err := app.Run(); err != nil {
		log.Fatal("Ошибка запуска сервера:", err)
	}
}
