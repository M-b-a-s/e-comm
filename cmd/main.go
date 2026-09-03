package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dbString := os.Getenv("GOOSE_DBSTRING")

	ctx := context.Background()

	cfg := config{
		addr: ":4000",
		db: dbConfig{
			dsn: dbString,
		},
	}

	// logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Database
	conn, err := pgx.Connect(ctx, cfg.db.dsn)
	if err != nil {
		panic(err)
	}
	defer conn.Close(ctx)

	logger.Info("DB connection established", "dsn", cfg.db.dsn)

	api := application{
		config: cfg,
		db:     conn,
	}

	handler := api.mount()
	if err := api.run(&handler); err != nil {
		slog.Error("Error starting server", "error", err)
		os.Exit(1)
	}
}
