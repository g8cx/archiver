package main

import (
	"archive-app/db"
	"archive-app/ui"
	"log"
)

func main() {
	// Подключение к базе данных
	err := db.InitDB()
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}
	defer db.Close()

	// Запуск графического интерфейса
	ui.StartUI()
}
