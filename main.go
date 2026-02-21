package main

import (
	"context"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/rl404/verniy"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MediaStatus string

const (
	StatusPlanning  MediaStatus = "Plan to Watch"
	StatusWatching  MediaStatus = "Watching"
	StatusCompleted MediaStatus = "Completed"
	StatusDropped   MediaStatus = "Dropped"
	StatusPaused    MediaStatus = "Paused"
)

type TrackerItem struct {
	gorm.Model
	Title           string      `json:"title"`
	Category        string      `json:"category" gorm:"uniqueIndex:idx_external"`
	ExternalID      int         `json:"external_id" gorm:"uniqueIndex:idx_external"`
	Status          MediaStatus `json:"status"`
	ProgressCurrent float64     `json:"progress_current"`
	ProgressTotal   float64     `json:"progress_total"`
	ProgressUnit    string      `json:"progress_unit"` // ep, ch, percent, min
}

var DB *gorm.DB

func initDatabase() {
	_ = os.Mkdir("data", 0755)
	var err error
	DB, err = gorm.Open(sqlite.Open("data/tracker.sqlite"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	err = DB.AutoMigrate(&TrackerItem{})
	if err != nil {
		panic("failed to migrate database")
	}
}

func fetchAniList(username string) ([]TrackerItem, error) {
	v := verniy.New()
	ctx := context.Background()

	collection, err := v.GetUserAnimeListWithContext(ctx, username)
	if err != nil {
		return nil, err
	}

	var items []TrackerItem

	for _, list := range collection {
		for _, entry := range list.Entries {
			// for not yet released anime, ProgressTotal is not available, so we set it to 0
			progress_total := 0.0
			if entry.Media.Episodes != nil {
				progress_total = float64(*entry.Media.Episodes)
			}
			title, err := fetchAniListAnimeName(entry.Media.ID)
			if err != nil {
				fmt.Printf("failed to fetch anime name for id %d: %v\n", entry.ID, err)
				title = "Unknown Title"
			}

			item := TrackerItem{
				Title:           title,
				Category:        "Anime",
				ExternalID:      entry.ID,
				Status:          mapAniListStatus(string(*entry.Status)),
				ProgressCurrent: float64(*entry.Progress),
				ProgressTotal:   progress_total,
				ProgressUnit:    "ep",
			}
			items = append(items, item)
		}
	}
	return items, nil
}

func fetchAniListAnimeName(id int) (string, error) {
	v := verniy.New()
	ctx := context.Background()

	media, err := v.GetAnimeWithContext(ctx, id)
	if err != nil {
		return "", err
	}
	if media.Title.Romaji == nil {
		return "", fmt.Errorf("title not found for media id %d", id)
	}
	return *media.Title.English, nil
}

func mapAniListStatus(status string) MediaStatus {
	switch status {
	case "CURRENT":
		return StatusWatching
	case "PLANNING":
		return StatusPlanning
	case "COMPLETED":
		return StatusCompleted
	case "DROPPED":
		return StatusDropped
	case "PAUSED":
		return StatusPaused
	default:
		return StatusPlanning
	}
}

func main() {
	initDatabase()

	r := gin.Default()
	r.GET("/items", func(c *gin.Context) {
		var items []TrackerItem
		DB.Find(&items)
		c.JSON(200, items)
	})

	r.POST("/items", func(c *gin.Context) {
		var item TrackerItem
		if err := c.ShouldBindJSON(&item); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		// upsert logic
		err := DB.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "category"}, {Name: "external_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"title", "status", "progress_current", "updated_at"}),
		}).Create(&item).Error

		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		// fetch the updated or created item to return in response
		DB.Where(
			"category = ? AND external_id = ?",
			item.Category,
			item.ExternalID,
		).First(&item)

		c.JSON(201, item)
	})

	r.POST("/sync/anilist", func(c *gin.Context) {
		username := c.Query("username")
		if username == "" {
			c.JSON(400, gin.H{"error": "username query parameter is required"})
			return
		}

		data, err := fetchAniList(username)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		for _, item := range data {
			// upsert logic
			err := DB.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "category"}, {Name: "external_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"title", "status", "progress_current", "updated_at"}),
			}).Create(&item).Error

			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}

			c.JSON(200, gin.H{"message": "Sync Complete", "item": item})
		}

	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	err := r.Run(fmt.Sprintf(":%s", port))
	if err != nil {
		panic("failed to start server")
	}
}
