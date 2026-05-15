package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"org-structure-api/internal/config"  
	"org-structure-api/internal/database"
	"org-structure-api/internal/handlers"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Файл .env не найден")
	}

	db, err := database.NewDB(config.GetDBConfig())
	if err != nil {
		log.Fatalf("Не удалось подключиться к базе данных: %v", err)
	}

	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("Миграция не удалась: %v", err)
	}

	router := handlers.NewRouter(db)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Сервер запущен http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}