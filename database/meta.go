package database

import (
	"gorm.io/gorm"
	"time"
)

type MovieInformation struct {
	gorm.Model
	Title   string
	Cover   string
	Release time.Time
	Credits []MovieCredit
}
type MovieCredit struct {
	gorm.Model
	Name               string
	Pic                string
	Character          string
	MovieInformationID uint
}
