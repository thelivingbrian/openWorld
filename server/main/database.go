package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func connectDB() {
	// Retrieve database credentials from environment variables
	dbUser := os.Getenv("BLOOP_DB_USER")
	dbPassword := os.Getenv("BLOOP_DB_PASSWORD")
	dbHost := os.Getenv("BLOOP_DB_HOST")
	dbName := os.Getenv("BLOOP_DB_NAME")
	dbPort := "5432"
	sslMode := "sslmode=disable"

	// Construct connection string
	connStr := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s %s", dbUser, dbPassword, dbHost, dbPort, dbName, sslMode)

	// Open database connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Error opening database: %s\n", err)
	}
	defer db.Close()

	// Insert a new user with only required fields
	_, err = db.Exec(`
			INSERT INTO "users" (email, username, hashword)
			VALUES ($1, $2, $3)`,
		"user@example.com", "newuser", "hashed_password")
	if err != nil {
		log.Fatalf("Error inserting new user: %s\n", err)
	}

	fmt.Println("New user added successfully")
}
