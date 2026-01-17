package models

import (
	"time"
)

type EventType string

const (
	EventTypeView  EventType = "view"
	EventTypeClick EventType = "click"
)

type Event struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ArticleID string    `gorm:"index" json:"article_id"`
	EventType EventType `gorm:"index" json:"event_type"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	CreatedAt time.Time `gorm:"index" json:"created_at"`
}
