package migrations

import (
	"fmt"
	"sports-backend-api/models"

	"gorm.io/gorm"
)

// Migrate runs the database migrations.
// It will create or update tables based on the GORM models.
func Migrate(db *gorm.DB) {
	fmt.Println("Running database migrations...")
	err := db.AutoMigrate(
		&models.User{},
		&models.TeamHQ{},
		&models.Player{},
		&models.MatchSchedule{},
		&models.MatchResult{},
		&models.PlayerScored{},
	)
	if err != nil {
		panic("Failed to migrate database: " + err.Error())
	}
	fmt.Println("Database migration completed successfully.")
}
