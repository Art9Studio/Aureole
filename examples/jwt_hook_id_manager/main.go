package main

import (
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/tern/migrate"
	"github.com/joho/godotenv"
	"os"
)

func main() {
	_ = godotenv.Load("./.env")
	manager, err := newIDManager()
	if err != nil {
		panic(err)
	}

	conn, err := manager.pool.Acquire(context.Background())
	if err != nil {
		panic(err)
	}
	defer conn.Release()
	err = runDBMigrations(conn.Conn())
	if err != nil {
		panic(err)
	}

	err = runServer(manager)
	if err != nil {
		panic(err)
	}
}

func runDBMigrations(conn *pgx.Conn) error {
	migrator, err := migrate.NewMigrator(context.Background(), conn, "schema_version")
	if err != nil {
		return fmt.Errorf("unable to create a migrator: %v", err)
	}

	err = migrator.LoadMigrations("db/migrations")
	if err != nil {
		return fmt.Errorf("unable to load migrations: %v", err)
	}
	return migrator.Migrate(context.Background())
}

func runServer(m *IDManager) error {
	fiberApp := fiber.New()
	fiberApp.Use(cors.New())
	fiberApp.Use(logger.New())
	fiberApp.Add(fiber.MethodPost, "/", handleRequest(m))

	var (
		addr string
		ok   bool
	)
	addr, ok = os.LookupEnv("ADDRESS")
	if !ok {
		addr = "localhost:3001"
	}

	return fiberApp.Listen(addr)
}
