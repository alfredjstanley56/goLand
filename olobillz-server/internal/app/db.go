package app

import (
	"log"
	"os"

	_ "github.com/denisenkom/go-mssqldb" // Import the Azure SQL driver
	"github.com/jmoiron/sqlx"
)

var db *sqlx.DB

// InitDB initializes the database connection
func InitDB() {
	// Retrieve the connection string from environment variables
	connectionString := os.Getenv("AZURE_SQL_CONNECTION_STRING")
	if connectionString == "" {
		log.Fatal("AZURE_SQL_CONNECTION_STRING not set in environment")
	}

	// Connect to the database
	var err error
	db, err = sqlx.Connect("sqlserver", connectionString)
	if err != nil {
		log.Fatalf("Failed to connect to Azure SQL: %s", err)
	}

	log.Println("Successfully connected to Azure SQL")
}

// GetDB returns the instance of the database
func GetDB() *sqlx.DB {
	return db
}
