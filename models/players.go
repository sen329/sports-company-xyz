package models

import (
	"time"

	"gorm.io/gorm"
)

type Player struct {
	Id         int64          `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name       string         `gorm:"column:name" json:"name"`
	Weight     int            `gorm:"column:weight" json:"weight"`
	Height     int            `gorm:"column:height" json:"height"`
	Position   string         `gorm:"column:position" json:"position"`
	BackNumber int            `gorm:"column:back_number" json:"back_number"`
	TeamId     int64          `gorm:"column:team_id" json:"team_id"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

type PlayerRequest struct {
	Name       string `form:"name" json:"name"`
	Weight     int    `form:"weight" json:"weight"`
	Height     int    `form:"height" json:"height"`
	Position   string `form:"position" json:"position"`
	BackNumber int    `form:"back_number" json:"back_number"`
	TeamId     int64  `form:"team_id" json:"team_id"`
	TeamName   string `form:"team_name" json:"team_name"`
	Status     string `form:"status" json:"status"`
	Page       int    `form:"page"`
	Limit      int    `form:"limit"`
}

// PlayerDetail is used to hold the result of a join query between players and team_hqs.
type PlayerDetail struct {
	Player
	TeamName string `gorm:"column:team_name" json:"team_name"`
}

type PlayerResponse struct {
	Id         int64  `json:"id"`
	Name       string `json:"name"`
	Weight     int    `json:"weight"`
	Height     int    `json:"height"`
	Position   string `json:"position"`
	BackNumber int    `json:"back_number"`
	TeamName   string `json:"team_name"`
}

type PaginatedPlayerResponse struct {
	Data         []PlayerResponse `json:"data"`
	TotalRecords int64            `json:"total_records"`
	CurrentPage  int              `json:"current_page"`
	PageSize     int              `json:"page_size"`
	TotalPages   int              `json:"total_pages"`
}
