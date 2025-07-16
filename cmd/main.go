package main

import (
	"context"
	"log"
	"net/http"
	_ "time/tzdata"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"github.com/tictactoe/internal/app"
	"github.com/tictactoe/internal/config"
	"github.com/tictactoe/internal/infrastructure/postgres"
	httpHandler "github.com/tictactoe/internal/interfaces/http"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Предупреждение: файл .env не найден: %v", err)
	}
}

func main() {
	cfg := config.New()

	db := cfg.ConnectDB()
	defer db.Close(context.Background())

	gameRepo := postgres.NewGameRepository(db)
	gameService := app.NewGameService(gameRepo)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	commandHandler := httpHandler.NewCommandHandler(gameService)
	commandHandler.RegisterRoutes(r)

	log.Printf("Сервер запущен на порту %s, окружение: %s", cfg.Port, cfg.Environment)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
