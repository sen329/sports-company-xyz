package repositories

import (
	"sports-backend-api/models"

	"gorm.io/gorm"
)

// MatchScheduleRepository defines the interface for match schedule data operations.
type MatchScheduleRepository interface {
	CreateMatchSchedule(match *models.MatchSchedule) error
	GetMatchScheduleByID(id int64) (*models.MatchScheduleDetail, error)
	UpdateMatchSchedule(match *models.MatchSchedule) error
	DeleteMatchSchedule(id int64) error
	GetMatchSchedulesByFilter(filter models.MatchScheduleRequest) ([]models.MatchScheduleDetail, int64, error)
	CheckTeamScheduleConflict(teamID int64, date string, matchIDToExclude int64) (bool, error)
}

type matchScheduleRepository struct {
	db *gorm.DB
}

// NewMatchScheduleRepository creates a new instance of MatchScheduleRepository.
func NewMatchScheduleRepository(db *gorm.DB) MatchScheduleRepository {
	return &matchScheduleRepository{db: db}
}

// CreateMatchSchedule adds a new match schedule to the database.
func (r *matchScheduleRepository) CreateMatchSchedule(match *models.MatchSchedule) error {
	return r.db.Create(match).Error
}

// GetMatchScheduleByID retrieves a match schedule by its ID.
func (r *matchScheduleRepository) GetMatchScheduleByID(id int64) (*models.MatchScheduleDetail, error) {
	var match models.MatchScheduleDetail
	err := r.db.Model(&models.MatchSchedule{}).
		Select("match_schedules.*, home_team.name as home_team_name, away_team.name as away_team_name").
		Joins("left join team_hqs as home_team on home_team.id = match_schedules.home_team_id").
		Joins("left join team_hqs as away_team on away_team.id = match_schedules.away_team_id").
		First(&match, "match_schedules.id = ?", id).Error
	return &match, err
}

// UpdateMatchSchedule updates an existing match schedule in the database.
func (r *matchScheduleRepository) UpdateMatchSchedule(match *models.MatchSchedule) error {
	return r.db.Model(match).Updates(match).Error
}

// DeleteMatchSchedule deletes a match schedule from the database by its ID.
func (r *matchScheduleRepository) DeleteMatchSchedule(id int64) error {
	return r.db.Delete(&models.MatchSchedule{}, id).Error
}

// GetMatchSchedulesByFilter retrieves a paginated list of match schedules based on filter criteria.
func (r *matchScheduleRepository) GetMatchSchedulesByFilter(filter models.MatchScheduleRequest) ([]models.MatchScheduleDetail, int64, error) {
	var total int64
	var matches []models.MatchScheduleDetail
	query := r.db.Model(&models.MatchSchedule{}).
		Select("match_schedules.*, home_team.name as home_team_name, away_team.name as away_team_name").
		Joins("left join team_hqs as home_team on home_team.id = match_schedules.home_team_id").
		Joins("left join team_hqs as away_team on away_team.id = match_schedules.away_team_id")

	if filter.Date != "" {
		query = query.Where("date = ?", filter.Date)
	}

	if filter.HomeTeamName != "" {
		query = query.Where("home_team.name LIKE ?", "%"+filter.HomeTeamName+"%")
	}

	if filter.AwayTeamName != "" {
		query = query.Where("away_team.name LIKE ?", "%"+filter.AwayTeamName+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.Limit
	err := query.Offset(offset).Limit(filter.Limit).Find(&matches).Error
	return matches, total, err
}

// CheckTeamScheduleConflict checks if a team is already scheduled for a match on a given date.
func (r *matchScheduleRepository) CheckTeamScheduleConflict(teamID int64, date string, matchIDToExclude int64) (bool, error) {
	var count int64
	query := r.db.Model(&models.MatchSchedule{}).
		Where("date = ? AND (home_team_id = ? OR away_team_id = ?)", date, teamID, teamID)
	if matchIDToExclude != 0 {
		query = query.Where("id != ?", matchIDToExclude)
	}
	err := query.Count(&count).Error
	return count > 0, err
}
