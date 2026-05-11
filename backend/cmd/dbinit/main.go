package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func main() {
	host := env("DB_HOST", "127.0.0.1")
	port := env("DB_PORT", "3306")
	user := env("DB_USER", "waapp")
	password := env("DB_PASSWORD", "waapp123")
	database := env("DB_NAME", "wa_platform")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/?parseTime=true&charset=utf8mb4&loc=UTC", user, password, host, port)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		panic(err)
	}

	query := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", database)
	if _, err := db.Exec(query); err != nil {
		panic(err)
	}

	if err := applyMigrations(db, database); err != nil {
		panic(err)
	}

	fmt.Printf("database ready: %s\n", database)
}

func applyMigrations(db *sql.DB, database string) error {
	if _, err := db.Exec(fmt.Sprintf("USE `%s`", database)); err != nil {
		return err
	}

	entries, err := os.ReadDir(filepath.Join("migrations"))
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		content, err := os.ReadFile(filepath.Join("migrations", entry.Name()))
		if err != nil {
			return err
		}
		statements := strings.Split(string(content), ";")
		for _, statement := range statements {
			trimmed := strings.TrimSpace(statement)
			if trimmed == "" {
				continue
			}
			if _, err := db.Exec(trimmed); err != nil {
				return fmt.Errorf("migration %s failed: %w", entry.Name(), err)
			}
		}
	}
	return nil
}
