package config

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
)

type AppConfig struct {
	ServiceName string
	Version     string
	Environment string
	Port        string
	DatabaseURL string
}

func New() *AppConfig {
	return &AppConfig{
		ServiceName: os.Getenv("SERVICE_NAME"),
		Version:     os.Getenv("VERSION"),
		Environment: os.Getenv("ENVIRONMENT"),
		Port:        os.Getenv("PORT"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}
}

func (c *AppConfig) ConnectDB() *pgx.Conn {
	conn, err := pgx.Connect(context.Background(), c.DatabaseURL)
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	return conn
}
