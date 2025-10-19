package repositories

import (
	"gorm.io/gorm"

	"sports-backend-api/models"
)

// TeamHQRepository defines the interface for data operations on the TeamHQ model.
type TeamHQRepository interface {
	CreateTeamHQ(team *models.TeamHQ) error
	GetTeamHQByID(id int64) (*models.TeamHQ, error)
	UpdateTeamHQ(team *models.TeamHQ) error
	DeleteTeamHQ(id int64) error
	GetTeamHQsByFilter(filter models.TeamHQRequest) ([]models.TeamHQ, int64, error)
}

type teamHQRepository struct {
	db *gorm.DB
}

// NewTeamHQRepository creates a new instance of TeamHQRepository.
func NewTeamHQRepository(db *gorm.DB) TeamHQRepository {
	return &teamHQRepository{db: db}
}

// CreateTeamHQ adds a new team headquarters to the database.
func (r *teamHQRepository) CreateTeamHQ(team *models.TeamHQ) error {
	return r.db.Create(team).Error
}

// GetTeamHQByID retrieves a single team headquarters by its ID.
func (r *teamHQRepository) GetTeamHQByID(id int64) (*models.TeamHQ, error) {
	var team models.TeamHQ
	err := r.db.First(&team, id).Error
	return &team, err
}

// UpdateTeamHQ updates an existing team headquarters' details.
func (r *teamHQRepository) UpdateTeamHQ(team *models.TeamHQ) error {
	return r.db.Model(team).Updates(team).Error
}

// DeleteTeamHQ removes a team headquarters from the database by its ID.
func (r *teamHQRepository) DeleteTeamHQ(id int64) error {
	return r.db.Delete(&models.TeamHQ{}, id).Error
}

// GetTeamHQsByFilter retrieves a paginated list of team headquarters based on filter criteria.
func (r *teamHQRepository) GetTeamHQsByFilter(filter models.TeamHQRequest) ([]models.TeamHQ, int64, error) {
	var total int64
	var teams []models.TeamHQ
	query := r.db.Model(&models.TeamHQ{})

	if filter.Name != "" {
		query = query.Where("name LIKE ?", "%"+filter.Name+"%")
	}
	if filter.Location != "" {
		query = query.Where("location LIKE ?", "%"+filter.Location+"%")
	}
	if filter.City != "" {
		query = query.Where("city LIKE ?", "%"+filter.City+"%")
	}

	// First, count the total records matching the filter
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Then, apply pagination to get the records for the current page
	offset := (filter.Page - 1) * filter.Limit
	if err := query.Offset(offset).Limit(filter.Limit).Find(&teams).Error; err != nil {
		return nil, 0, err
	}

	return teams, total, nil
}
