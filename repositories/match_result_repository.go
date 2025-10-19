package repositories

import (
	"sports-backend-api/models"

	"gorm.io/gorm"
)

// MatchResultRepository defines the interface for match result data operations.
type MatchResultRepository interface {
	CreateMatchResult(result *models.MatchResult) error
	GetMatchResultByMatchID(matchID int64) (*models.MatchResult, error)
	CheckResultExists(matchID int64) (bool, error)
}

type matchResultRepository struct {
	db *gorm.DB
}

// NewMatchResultRepository creates a new instance of MatchResultRepository.
func NewMatchResultRepository(db *gorm.DB) MatchResultRepository {
	return &matchResultRepository{db: db}
}

// CreateMatchResult adds a new match result to the database.
// It uses a transaction to ensure that the match result and all player scores are created atomically.
func (r *matchResultRepository) CreateMatchResult(result *models.MatchResult) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Create the main match result record
		if err := tx.Create(result).Error; err != nil {
			return err
		}
		return nil
	})
}

// GetMatchResultByMatchID retrieves a match result by its associated match ID, preloading player scores.
func (r *matchResultRepository) GetMatchResultByMatchID(matchID int64) (*models.MatchResult, error) {
	var result models.MatchResult
	err := r.db.Preload("PlayerScored").Where("match_id = ?", matchID).First(&result).Error
	return &result, err
}

// CheckResultExists checks if a result for a given match ID already exists.
func (r *matchResultRepository) CheckResultExists(matchID int64) (bool, error) {
	var count int64
	err := r.db.Model(&models.MatchResult{}).Where("match_id = ?", matchID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
