// Package db initializes and works on the database
package db

import (
	"os"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var DB *gorm.DB

// UpsertMedia performs an upsert operation on any media item
func UpsertMedia(item any, updateColumns []string) error {
	return DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "username"}, {Name: "external_id"}},
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
	err := DB.AutoMigrate(models...)
	if err != nil {
		return err
	}

	// Create composite unique indexes for each table
	if err := DB.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_animes_user_external ON animes(username, external_id)").Error; err != nil {
		return err
	}
	if err := DB.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_mangas_user_external ON mangas(username, external_id)").Error; err != nil {
		return err
	}

	return nil
}
