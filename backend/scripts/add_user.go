package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Database connection
	dsn := "waapp:waapp123@tcp(127.0.0.1:3307)/wa_platform"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	username := "ara"
	password := "ara321"
	role := "admin"

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	// Upsert user (insert or update if exists)
	query := `
		INSERT INTO users (username, password_hash, role)
		VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE
			password_hash = VALUES(password_hash),
			role = VALUES(role),
			updated_at = CURRENT_TIMESTAMP
	`

	result, err := db.Exec(query, username, string(hashedPassword), role)
	if err != nil {
		log.Fatalf("Failed to insert/update user: %v", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		log.Fatalf("Failed to get rows affected: %v", err)
	}

	fmt.Printf("Successfully added user '%s' with role '%s' (rows affected: %d)\n", username, role, rows)
}
