package main

import (
	"fmt"
	"log"
	"sports-backend-api/database"
	"sports-backend-api/migrations"
	"sports-backend-api/routes"
	"sports-backend-api/util"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// Load environment variables
	util.LoadEnv()

	// Initialize database connection
	dbConfig := database.BuildConfig()
	dsn := database.DBUrl(dbConfig)
	var err error
	// Add parseTime=true to handle time.Time fields correctly.
	// It ensures the driver parses MySQL DATETIME/TIMESTAMP columns into time.Time structs.
	database.DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true, // Improves performance by avoiding auto-transactions.
		PrepareStmt:            true, // Caches compiled statements for performance and helps prevent SQL injection.
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	fmt.Println("Database connection successful.")

	// Run migrations
	migrations.Migrate(database.DB)

	// Setup and run the router
	routes.SetupRoutes()
}
