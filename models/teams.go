package models

import (
	"time"

	"gorm.io/gorm"
)

type TeamHQ struct {
	Id        int64          `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name      string         `gorm:"column:name" json:"name"`
	Logo      string         `gorm:"column:logo" json:"logo"`
	Location  string         `gorm:"column:location" json:"location"`
	City      string         `gorm:"column:city" json:"city"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type TeamHQRequest struct {
	Name     string `form:"name" json:"name"`
	Logo     string `form:"logo" json:"logo"`
	Location string `form:"location" json:"location"`
	City     string `form:"city" json:"city"`
	Page     int    `form:"page"`
	Limit    int    `form:"limit"`
}

type TeamHQResponse struct {
	Id       int64  `json:"id"`
	Name     string `json:"name"`
	Logo     string `json:"logo"`
	Location string `json:"location"`
	City     string `json:"city"`
}

type PaginatedTeamHQResponse struct {
	Data         []TeamHQ `json:"data"`
	TotalRecords int64    `json:"total_records"`
	CurrentPage  int      `json:"current_page"`
	PageSize     int      `json:"page_size"`
	TotalPages   int      `json:"total_pages"`
}
