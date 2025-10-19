package models

import (
	"time"

	"gorm.io/gorm"
)

type MatchResult struct {
	Id           int64          `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	MatchId      int64          `gorm:"column:match_id;unique" json:"match_id"`
	HomeScore    int            `gorm:"column:home_score" json:"home_score"`
	AwayScore    int            `gorm:"column:away_score" json:"away_score"`
	WinnerTeamId int64          `gorm:"column:winner_team_id" json:"winner_team_id"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	PlayerScored []PlayerScored `gorm:"foreignKey:MatchResultId" json:"player_scored"`
}

type MatchResultDetail struct {
	MatchResult
	HomeTeamName      string `gorm:"column:home_team_name" json:"home_team_name"`
	AwayTeamName      string `gorm:"column:away_team_name" json:"away_team_name"`
	MatchStatus       string `gorm:"column:match_status" json:"match_status"`
	MVP               string `gorm:"column:mvp" json:"mvp"`
	HomeTeamTotalWins int64  `gorm:"column:home_team_total_wins" json:"home_team_total_wins"`
	AwayTeamTotalWins int64  `gorm:"column:away_team_total_wins" json:"away_team_total_wins"`
}

type PlayerScored struct {
	Id            int64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	MatchId       int64 `gorm:"column:match_id" json:"match_id"`
	PlayerId      int64 `gorm:"column:player_id" json:"player_id"`
	TeamId        int64 `gorm:"column:team_id" json:"team_id"`
	TimeScored    int   `gorm:"column:time_scored" json:"time_scored"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
	MatchResultId int64          `gorm:"column:match_result_id"`
}

type MatchResultRequest struct {
	MatchId      int64          `json:"match_id" binding:"required"`
	HomeScore    int            `json:"home_score"`
	AwayScore    int            `json:"away_score"`
	PlayerScored []PlayerScored `json:"player_scored"`
}
