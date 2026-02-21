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
	StatusPlanningWatch MediaStatus = "Plan to Watch"
	StatusWatching      MediaStatus = "Watching"
	StatusPlanningRead  MediaStatus = "Plan to Read"
	StatusReading       MediaStatus = "Reading"
	StatusCompleted     MediaStatus = "Completed"
	StatusDropped       MediaStatus = "Dropped"
	StatusPaused        MediaStatus = "Paused"
)

type BaseMedia struct {
	gorm.Model
	Title           string      `json:"title"`
	ExternalID      int         `json:"external_id" gorm:"uniqueIndex"`
	Status          MediaStatus `json:"status"`
	ProgressCurrent float64     `json:"progress_current"`
	ProgressTotal   float64     `json:"progress_total"`
	ProgressUnit    string      `json:"progress_unit"` // ep, ch, percent, min
}

type Anime struct {
	BaseMedia
}

type Manga struct {
	BaseMedia
}

var DB *gorm.DB

// upsertMedia performs an upsert operation on any media item
func upsertMedia(item interface{}, updateColumns []string) error {
	return DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "external_id"}},
		DoUpdates: clause.AssignmentColumns(updateColumns),
	}).Create(item).Error
}

func initDatabase() {
	_ = os.Mkdir("data", 0755)
	var err error
	DB, err = gorm.Open(sqlite.Open("data/tracker.sqlite"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	err = DB.AutoMigrate(&Anime{}, &Manga{})
	if err != nil {
		panic("failed to migrate database")
	}
}

// extractTitle extracts the title from AniList media entry
// Prefers English title, falls back to Romaji, then "Unknown Title"
func extractTitle(mediaID int, media *verniy.Media) string {
	title := "Unknown Title"

	if media != nil && media.Title != nil {
		if media.Title.English != nil {
			title = *media.Title.English
		} else if media.Title.Romaji != nil {
			title = *media.Title.Romaji
		} else {
			print("No title found for media id ", mediaID, "\n", "Using 'Unknown Title' as fallback\n")
		}
	}

	return title
}

// mapAniListStatus maps AniList status to internal MediaStatus
func mapAniListStatus(status string, isAnime bool) MediaStatus {
	switch status {
	case "CURRENT":
		if isAnime {
			return StatusWatching
		}
		return StatusReading
	case "PLANNING":
		if isAnime {
			return StatusPlanningWatch
		}
		return StatusPlanningRead
	case "COMPLETED":
		return StatusCompleted
	case "DROPPED":
		return StatusDropped
	case "PAUSED":
		return StatusPaused
	default:
		return StatusPlanningWatch
	}
}

func fetchAniListAnime(username string) ([]Anime, error) {
	v := verniy.New()
	ctx := context.Background()

	collection, err := v.GetUserAnimeListWithContext(ctx, username)
	if err != nil {
		return nil, err
	}

	var items []Anime

	for _, list := range collection {
		for _, entry := range list.Entries {
			// for not yet released anime, ProgressTotal is not available, so we set it to 0
			progressTotal := 0.0

			if entry.Media != nil && entry.Media.Episodes != nil {
				progressTotal = float64(*entry.Media.Episodes)
			} else {
				print("No episode count found for media id ", entry.ID, "\n")
				print("Using 0 as fallback for progress_total\n")
			}

			item := Anime{
				BaseMedia: BaseMedia{
					Title:           extractTitle(entry.ID, entry.Media),
					ExternalID:      entry.ID,
					Status:          mapAniListStatus(string(*entry.Status), true),
					ProgressCurrent: float64(*entry.Progress),
					ProgressTotal:   progressTotal,
					ProgressUnit:    "ep",
				},
			}
			items = append(items, item)
		}
	}
	return items, nil
}

func fetchAniListManga(username string) ([]Manga, error) {
	v := verniy.New()
	ctx := context.Background()

	collection, err := v.GetUserMangaListWithContext(ctx, username)
	if err != nil {
		return nil, err
	}

	var items []Manga

	for _, list := range collection {
		for _, entry := range list.Entries {
			// for ongoing manga, Anilist doesn't track total chapters released, so we set it to chapters read
			progressTotal := 0.0

			if entry.Media != nil && entry.Media.Chapters != nil {
				progressTotal = float64(*entry.Media.Chapters)
			} else {
				print("No chapter count found for media id ", entry.ID, "\n")
				print("Using chapters read as fallback for progress_total\n")
				progressTotal = float64(*entry.Progress)
			}

			item := Manga{
				BaseMedia: BaseMedia{
					Title:           extractTitle(entry.ID, entry.Media),
					ExternalID:      entry.ID,
					Status:          mapAniListStatus(string(*entry.Status), false),
					ProgressCurrent: float64(*entry.Progress),
					ProgressTotal:   progressTotal,
					ProgressUnit:    "ch",
				},
			}
			items = append(items, item)
		}
	}
	return items, nil
}

func main() {
	initDatabase()

	r := gin.Default()
	r.GET("/items/anime", func(c *gin.Context) {
		var items []Anime
		DB.Find(&items)
		c.JSON(200, items)
	})

	r.GET("/items/manga", func(c *gin.Context) {
		var items []Manga
		DB.Find(&items)
		c.JSON(200, items)
	})

	r.POST("/items/anime", func(c *gin.Context) {
		var item Anime
		if err := c.ShouldBindJSON(&item); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		// upsert logic
		err := upsertMedia(&item, []string{"title", "status", "progress_current", "updated_at"})
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		// fetch the updated or created item to return in response
		DB.Where("external_id = ?", item.ExternalID).First(&item)

		c.JSON(201, item)
	})

	r.POST("/items/manga", func(c *gin.Context) {
		var item Manga
		if err := c.ShouldBindJSON(&item); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		// upsert logic
		err := upsertMedia(&item, []string{"title", "status", "progress_current", "updated_at"})
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		// fetch the updated or created item to return in response
		DB.Where("external_id = ?", item.ExternalID).First(&item)

		c.JSON(201, item)
	})

	r.POST("/sync/anilist/anime", func(c *gin.Context) {
		username := c.Query("username")
		if username == "" {
			c.JSON(400, gin.H{"error": "username query parameter is required"})
			return
		}

		data, err := fetchAniListAnime(username)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		for _, item := range data {
			err := upsertMedia(&item, []string{"title", "status", "progress_current", "updated_at"})
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
		}

		c.JSON(200, gin.H{"message": "Sync Complete", "count": len(data)})
	})

	r.POST("/sync/anilist/manga", func(c *gin.Context) {
		username := c.Query("username")
		if username == "" {
			c.JSON(400, gin.H{"error": "username query parameter is required"})
			return
		}

		data, err := fetchAniListManga(username)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		for _, item := range data {
			err := upsertMedia(&item, []string{"title", "status", "progress_current", "progress_total", "updated_at"})
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
		}

		c.JSON(200, gin.H{"message": "Sync Complete", "count": len(data)})
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
