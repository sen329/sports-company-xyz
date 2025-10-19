package repositories

import (
	"sports-backend-api/models"

	"gorm.io/gorm"
)

// MatchResultDetailRepository defines the interface for detailed match result data operations.
type MatchResultDetailRepository interface {
	GetMatchResultDetailByMatchID(matchID int64) (*models.MatchResultDetail, error)
}

type matchResultDetailRepository struct {
	db *gorm.DB
}

// NewMatchResultDetailRepository creates a new instance of MatchResultDetailRepository.
func NewMatchResultDetailRepository(db *gorm.DB) MatchResultDetailRepository {
	return &matchResultDetailRepository{db: db}
}

// GetMatchResultDetailByMatchID retrieves a detailed view of a match result by its associated match ID.
func (r *matchResultDetailRepository) GetMatchResultDetailByMatchID(matchID int64) (*models.MatchResultDetail, error) {
	var detail models.MatchResultDetail

	// Pre-fetch schedule to get team IDs for status determination
	schedule, err := NewMatchScheduleRepository(r.db).GetMatchScheduleByID(matchID)
	if err != nil {
		return nil, err // Could fail if match schedule is deleted after result is created
	}

	// Step 1: Fetch the base MatchResult and join with MatchSchedule and TeamHQs to get team names.
	err = r.db.Model(&models.MatchResult{}).
		Select("match_results.*, home_team.name as home_team_name, away_team.name as away_team_name").
		Joins("JOIN match_schedules ms ON ms.id = match_results.match_id").
		Joins("JOIN team_hqs AS home_team ON home_team.id = ms.home_team_id").
		Joins("JOIN team_hqs AS away_team ON away_team.id = ms.away_team_id").
		Preload("PlayerScored").
		Where("match_results.match_id = ?", matchID).
		First(&detail).Error

	if err != nil {
		return nil, err
	}

	// Step 2: Determine Match Status
	switch detail.WinnerTeamId {
	case 0:
		detail.MatchStatus = "Draw"
	case schedule.HomeTeamId:
		detail.MatchStatus = "Home team wins"
	case schedule.AwayTeamId:
		detail.MatchStatus = "Away team wins"
	}

	// Step 3: Calculate MVP (Most Valuable Player)
	// The MVP is the player who scored the most goals in this match.
	var mvpResult struct {
		Name string
	}
	err = r.db.Model(&models.PlayerScored{}).
		Select("p.name").
		Joins("JOIN players p ON p.id = player_scoreds.player_id").
		Where("player_scoreds.match_id = ?", matchID).
		Group("p.name").
		Order("COUNT(player_scoreds.player_id) DESC").
		Limit(1).
		Scan(&mvpResult).Error

	if err == nil {
		detail.MVP = mvpResult.Name
	}

	// Step 4: Get total wins for both teams
	r.db.Model(&models.MatchResult{}).Where("winner_team_id = ?", schedule.HomeTeamId).Count(&detail.HomeTeamTotalWins)
	r.db.Model(&models.MatchResult{}).Where("winner_team_id = ?", schedule.AwayTeamId).Count(&detail.AwayTeamTotalWins)

	return &detail, nil
}
