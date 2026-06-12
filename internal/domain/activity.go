package domain

import (
	"time"
)

type Activity struct {
	ID          string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	Title       string    `gorm:"type:varchar(255);not null" json:"title"`
	StartDate   string    `gorm:"type:date;not null" json:"start_date"` // YYYY-MM-DD
	EndDate     *string   `gorm:"type:date" json:"end_date"`            // YYYY-MM-DD (nullable)
	StartTime   *string   `gorm:"type:varchar(5)" json:"start_time"`    // HH:MM (nullable)
	EndTime     *string   `gorm:"type:varchar(5)" json:"end_time"`      // HH:MM (nullable)
	Location    string    `gorm:"type:varchar(255)" json:"location"`
	ColorLabel  string    `gorm:"type:varchar(50)" json:"color_label"`
	Description *string   `gorm:"type:text" json:"description"` // (nullable)
	Source      string    `gorm:"type:varchar(50);default:'local'" json:"source"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (Activity) TableName() string {
	return "activities"
}

type CreateActivityRequest struct {
	Title       string  `json:"title"`
	StartDate   string  `json:"start_date"`
	EndDate     *string `json:"end_date"`
	StartTime   *string `json:"start_time"`
	EndTime     *string `json:"end_time"`
	Location    string  `json:"location"`
	ColorLabel  string  `json:"color_label"`
	Description *string `json:"description"`
}
