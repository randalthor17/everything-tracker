// Package db initializes and works on the database
package db

import (
	"os"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var DB *gorm.DB

// upsertMedia performs an upsert operation on any media item
func UpsertMedia(item any, updateColumns []string) error {
	return DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "external_id"}},
		DoUpdates: clause.AssignmentColumns(updateColumns),
	}).Create(item).Error
}

func InitDatabase() {
	_ = os.Mkdir("data", 0o755)
	var err error
	DB, err = gorm.Open(sqlite.Open("data/tracker.sqlite"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
}

// MigrateModels migrates the provided models into the database
func MigrateModels(models ...any) error {
	return DB.AutoMigrate(models...)
}
