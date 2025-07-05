package db

import (
	"context"
	"database/sql"
	"time"

	"driftGo/config"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	log "github.com/sirupsen/logrus"
)

func runMigrations() error {
	db, err := sql.Open("postgres", config.DatabaseURL)
	if err != nil {
		return err
	}
	defer db.Close()

	log.Info("Running database migrations...")
	if err := goose.Up(db, "db/goose_migrations"); err != nil {
		return err
	}
	return nil
}

func InitDB() *pgxpool.Pool {
	dbURL := config.DatabaseURL
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set in config")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}

	if err = pool.Ping(ctx); err != nil {
		log.Fatalf("Database ping failed: %v", err)
	}

	log.Info("Successfully connected to PostgreSQL")

	if err := runMigrations(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	return pool
}
