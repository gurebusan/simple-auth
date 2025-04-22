package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	var storageDSN, migrationsPath, migrationsTable string
	flag.StringVar(&storageDSN, "storage-dsn", os.Getenv("STORAGE_DSN"), "storage DSN")
	flag.StringVar(&migrationsPath, "migrations-path", os.Getenv("MIGRATION_PATH"), "path to migrations")
	flag.StringVar(&migrationsTable, "migrations-table", "migrations", "name of migrations table")
	flag.Parse()

	if storageDSN == "" {
		panic("sotrage dsn is required")
	}
	if migrationsPath == "" {
		panic("migrations-path is required")
	}
	// Добавляем параметр таблицы миграций в DSN
	migrationDSN := fmt.Sprintf("%s&x-migrations-table=%s", storageDSN, migrationsTable)

	m, err := migrate.New(
		"file://"+migrationsPath,
		migrationDSN,
	)
	if err != nil {
		panic(err)
	}
	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("no migrations to apply")
			return
		}
		panic(err)
	}

	fmt.Println("migrations applied")
}
