package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// StringArray is a custom type for handling JSON arrays in SQLite
type StringArray []string

func (a StringArray) Value() (driver.Value, error) {
	return json.Marshal(a)
}

func (a *StringArray) Scan(value interface{}) error {
	if value == nil {
		*a = []string{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, a)
}

// Article represents a news article
type Article struct {
	ID              string      `gorm:"primaryKey" json:"id"`
	Title           string      `gorm:"index" json:"title"`
	Description     string      `json:"description"`
	URL             string      `json:"url"`
	PublicationDate time.Time   `gorm:"index" json:"publication_date"`
	SourceName      string      `gorm:"index" json:"source_name"`
	Category        StringArray `gorm:"type:text" json:"category"`
	RelevanceScore  float64     `gorm:"index" json:"relevance_score"`
	Latitude        float64     `json:"latitude"`
	Longitude       float64     `json:"longitude"`
	LLMSummary      string      `json:"llm_summary,omitempty"`
	CreatedAt       time.Time   `json:"-"`
	UpdatedAt       time.Time   `json:"-"`
}

func (Article) TableName() string {
	return "articles"
}

// BeforeCreate hook to set timestamps
func (a *Article) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	a.CreatedAt = now
	a.UpdatedAt = now
	return nil
}
