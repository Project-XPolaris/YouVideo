package database

import "gorm.io/gorm"

type Entity struct {
	gorm.Model
	Name      string   `json:"name"`
	Videos    []*Video `json:"videos"`
	LibraryId uint     `json:"library_id"`
	Cover     string   `json:"cover"`
	Summary   string   `json:"summary"`
	Template  string   `json:"template"`
}
