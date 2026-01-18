package models

import (
	"time"

	"gorm.io/gorm"
)

// EventType represents the type of user interaction
type EventType string

const (
	EventTypeView  EventType = "view"
	EventTypeClick EventType = "click"
)

// Event represents a simulated user interaction with an article
type Event struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	ArticleID  string    `gorm:"index" json:"article_id"`
	EventType  EventType `gorm:"index" json:"event_type"`
	Latitude   float64   `json:"latitude"`
	Longitude  float64   `json:"longitude"`
	Timestamp  time.Time `gorm:"index" json:"timestamp"`
	CreatedAt  time.Time `json:"-"`
}

func (Event) TableName() string {
	return "events"
}

// BeforeCreate hook to set timestamps
func (e *Event) BeforeCreate(tx *gorm.DB) error {
	if e.Timestamp.IsZero() {
		e.Timestamp = time.Now()
	}
	e.CreatedAt = time.Now()
	return nil
}
