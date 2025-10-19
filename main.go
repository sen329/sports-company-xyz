package main

import (
	"fmt"
	"log"
	"sports-backend-api/database"
	"sports-backend-api/migrations"
	"sports-backend-api/routes"
	"sports-backend-api/util"
)

func main() {
	// Load environment variables
	util.LoadEnv()

	// Initialize database connection
	err := database.Database()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	fmt.Println("Database connection successful.")

	// Run migrations
	migrations.Migrate(database.DB)

	// Setup and run the router
	routes.SetupRoutes()
}
