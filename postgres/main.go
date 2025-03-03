package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	query := os.Getenv("QUERY")
	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")
	database := os.Getenv("POSTGRES_DATABASE")
	user := os.Getenv("POSTGRES_USERNAME")
	password := os.Getenv("POSTGRES_PASSWORD")
	sslmode := os.Getenv("POSTGRES_SSLMODE")

	if sslmode == "" {
		sslmode = "require"
	}

	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=%s",
		user, password, host, port, database, sslmode)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Printf("Error opening database: %v\n", err)
		os.Exit(1)
	}

	// Check for validate mode
	if len(os.Args) > 1 && os.Args[1] == "validate" {
		// Test the connection with a simple query
		var result int
		err = db.Raw("SELECT 1").Scan(&result).Error
		if err != nil {
			fmt.Printf("Error validating connection: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Database connection validated successfully")
		os.Exit(0)
	}

	// Check if the query returns data
	queryUpper := strings.TrimSpace(strings.ToUpper(query))
	returnsData := strings.Contains(queryUpper, "SELECT") ||
		strings.Contains(queryUpper, "RETURNING") ||
		strings.HasPrefix(queryUpper, "VALUES") ||
		strings.HasPrefix(queryUpper, "SHOW")

	if returnsData {
		var results []map[string]any
		err := db.Raw(query).Scan(&results).Error
		if err != nil {
			fmt.Printf("Error executing query: %v\n", err)
			os.Exit(1)
		}

		jsonData, err := json.Marshal(results)
		if err != nil {
			fmt.Printf("Error marshaling JSON: %v\n", err)
			os.Exit(1)
		}

		fmt.Println(string(jsonData))
	} else {
		result := db.Exec(query)
		if result.Error != nil {
			fmt.Printf("Error executing query: %v\n", result.Error)
			os.Exit(1)
		}

		fmt.Printf("Query executed successfully! Rows affected: %d\n", result.RowsAffected)
	}
}
