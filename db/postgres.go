package db

import (
	"context"
	"database/sql"
	"log"
	"time"

	"driftGo/config"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

func runMigrations() error {
	db, err := sql.Open("postgres", config.DatabaseURL)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := goose.Up(db, "db/goose_migrations"); err != nil {
		return err
	}
	log.Println("goose: Migrations completed successfully")
	return nil
}

// InitDB initializes and returns a database connection pool
func InitDB() *pgxpool.Pool {
	dbURL := config.DatabaseURL
	if dbURL == "" {
		log.Fatal("postgres: DATABASE_URL is not set in config")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("postgres: Unable to connect to database: %v", err)
	}

	if err = pool.Ping(ctx); err != nil {
		log.Fatalf("postgres: Database ping failed: %v", err)
	}

	log.Println("postgres: Successfully connected to PostgreSQL")

	// Run migrations after successful connection
	if err := runMigrations(); err != nil {
		log.Fatalf("goose: Failed to run migrations: %v", err)
	}

	return pool
}
