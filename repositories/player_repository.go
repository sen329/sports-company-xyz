package repositories

import (
	"gorm.io/gorm"

	"sports-backend-api/models"
)

// PlayerRepository defines the interface for data operations on the Player model.
type PlayerRepository interface {
	CreatePlayer(player *models.Player) error
	GetPlayerByID(id int64) (*models.PlayerDetail, error)
	UpdatePlayer(player *models.Player) error
	DeletePlayer(id int64) error
	GetPlayerByTeamAndBackNumber(teamID int64, backNumber int) (*models.Player, error)
	GetPlayersByFilter(filter models.PlayerRequest) ([]models.PlayerDetail, int64, error)
}

type playerRepository struct {
	db *gorm.DB
}

// NewPlayerRepository creates a new instance of PlayerRepository.
func NewPlayerRepository(db *gorm.DB) PlayerRepository {
	return &playerRepository{db: db}
}

// CreatePlayer adds a new player to the database.
func (r *playerRepository) CreatePlayer(player *models.Player) error {
	return r.db.Create(player).Error
}

// GetPlayerByID retrieves a single player by its ID.
func (r *playerRepository) GetPlayerByID(id int64) (*models.PlayerDetail, error) {
	var player models.PlayerDetail
	err := r.db.Model(&models.Player{}).
		Select("players.*, team_hqs.name as team_name").
		Joins("left join team_hqs on team_hqs.id = players.team_id").
		First(&player, "players.id = ?", id).Error
	return &player, err
}

// UpdatePlayer updates an existing player's details.
func (r *playerRepository) UpdatePlayer(player *models.Player) error {
	return r.db.Model(player).Updates(player).Error
}

// DeletePlayer removes a player from the database by its ID.
func (r *playerRepository) DeletePlayer(id int64) error {
	return r.db.Delete(&models.Player{}, id).Error
}

// GetPlayerByTeamAndBackNumber retrieves a player by their team ID and back number.
func (r *playerRepository) GetPlayerByTeamAndBackNumber(teamID int64, backNumber int) (*models.Player, error) {
	var player models.Player
	err := r.db.Where("team_id = ? AND back_number = ?", teamID, backNumber).First(&player).Error
	return &player, err
}

// GetPlayersByFilter retrieves a paginated list of players based on filter criteria.
func (r *playerRepository) GetPlayersByFilter(filter models.PlayerRequest) ([]models.PlayerDetail, int64, error) {
	var total int64
	var players []models.PlayerDetail
	query := r.db.Model(&models.Player{}).
		Select("players.*, team_hqs.name as team_name").
		Joins("left join team_hqs on team_hqs.id = players.team_id")

	if filter.Name != "" {
		query = query.Where("name LIKE ?", "%"+filter.Name+"%")
	}
	if filter.Position != "" {
		query = query.Where("position = ?", filter.Position)
	}
	if filter.TeamName != "" {
		query = query.Where("team_hqs.name LIKE ?", "%"+filter.TeamName+"%")
	}
	// Allow fetching soft-deleted records if status=inactive is specified
	if filter.Status != "" {
		switch filter.Status {
		case "inactive":
			query = query.Unscoped().Where("players.deleted_at IS NOT NULL")
		case "active":
			// This is the default behavior, but explicit for clarity
			query = query.Where("players.deleted_at IS NULL")
		}
	}

	// Count the total records matching the filter
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	offset := (filter.Page - 1) * filter.Limit
	if err := query.Offset(offset).Limit(filter.Limit).Find(&players).Error; err != nil {
		return nil, 0, err
	}

	return players, total, nil
}
