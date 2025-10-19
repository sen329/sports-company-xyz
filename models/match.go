package models

import (
	"time"

	"gorm.io/gorm"
)

type MatchSchedule struct {
	Id         int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Date       string `gorm:"column:date" json:"date"`
	Time       string `gorm:"column:time" json:"time"`
	HomeTeamId int64  `gorm:"column:home_team_id" json:"home_team_id"`
	AwayTeamId int64  `gorm:"column:away_team_id" json:"away_team_id"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  gorm.DeletedAt `gorm:"index"`
}

type MatchScheduleRequest struct {
	Date         string `form:"date" json:"date"`
	Time         string `form:"time" json:"time"`
	HomeTeamId   int64  `json:"home_team_id"`
	AwayTeamId   int64  `json:"away_team_id"`
	HomeTeamName string `form:"home_team_name"`
	AwayTeamName string `form:"away_team_name"`
	Page         int    `form:"page"`
	Limit        int    `form:"limit"`
}

// MatchScheduleDetail is used to hold the result of a join query.
type MatchScheduleDetail struct {
	MatchSchedule
	HomeTeamName string `gorm:"column:home_team_name" json:"home_team_name"`
	AwayTeamName string `gorm:"column:away_team_name" json:"away_team_name"`
}

type MatchScheduleResponse struct {
	Id           int64  `json:"id"`
	Date         string `json:"date"`
	Time         string `json:"time"`
	HomeTeamName string `json:"home_team_name"`
	AwayTeamName string `json:"away_team_name"`
}

type PaginatedMatchScheduleResponse struct {
	Data         []MatchScheduleResponse `json:"data"`
	TotalRecords int64                   `json:"total_records"`
	CurrentPage  int                     `json:"current_page"`
	PageSize     int                     `json:"page_size"`
	TotalPages   int                     `json:"total_pages"`
}
