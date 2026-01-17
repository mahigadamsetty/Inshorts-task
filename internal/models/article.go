package models

import (
	"time"

	"gorm.io/gorm"
)

type Article struct {
	ID              string    `gorm:"primaryKey" json:"id"`
	Title           string    `gorm:"index" json:"title"`
	Description     string    `json:"description"`
	URL             string    `json:"url"`
	PublicationDate time.Time `gorm:"index" json:"publication_date"`
	SourceName      string    `gorm:"index" json:"source_name"`
	Category        string    `gorm:"index" json:"-"` // Store as comma-separated for SQL indexing
	CategoryArray   []string  `gorm:"-" json:"category"`
	RelevanceScore  float64   `gorm:"index" json:"relevance_score"`
	Latitude        float64   `gorm:"index" json:"latitude"`
	Longitude       float64   `gorm:"index" json:"longitude"`
	LLMSummary      string    `json:"llm_summary,omitempty"`
	CreatedAt       time.Time `json:"-"`
	UpdatedAt       time.Time `json:"-"`
}

func (a *Article) BeforeSave(tx *gorm.DB) error {
	// Convert array to comma-separated string for storage
	if len(a.CategoryArray) > 0 {
		category := ""
		for i, cat := range a.CategoryArray {
			if i > 0 {
				category += ","
			}
			category += cat
		}
		a.Category = category
	}
	return nil
}

func (a *Article) AfterFind(tx *gorm.DB) error {
	// Convert comma-separated string back to array
	if a.Category != "" && len(a.CategoryArray) == 0 {
		categories := []string{}
		start := 0
		for i, c := range a.Category {
			if c == ',' {
				categories = append(categories, a.Category[start:i])
				start = i + 1
			}
		}
		if start < len(a.Category) {
			categories = append(categories, a.Category[start:])
		}
		a.CategoryArray = categories
	}
	return nil
}
