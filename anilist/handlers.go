package anilist

import (
	"strconv"

	"everythingtracker/db"

	"github.com/gin-gonic/gin"
)

// GetAnimeHandler handles GET requests for anime items
func GetAnimeHandler(c *gin.Context) {
	var items []Anime
	db.DB.Find(&items)
	c.JSON(200, items)
}

// GetMangaHandler handles GET requests for manga items
func GetMangaHandler(c *gin.Context) {
	var items []Manga
	db.DB.Find(&items)
	c.JSON(200, items)
}

// PostAnimeHandler handles POST requests for anime items
func PostAnimeHandler(c *gin.Context) {
	var item Anime
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// upsert logic
	err := db.UpsertMedia(&item, []string{"title", "status", "progress_current", "updated_at"})
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// fetch the updated or created item to return in response
	db.DB.Where("external_id = ?", item.ExternalID).First(&item)

	c.JSON(201, item)
}

// PostMangaHandler handles POST requests for manga items
func PostMangaHandler(c *gin.Context) {
	var item Manga
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// upsert logic
	err := db.UpsertMedia(&item, []string{"title", "status", "progress_current", "updated_at"})
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// fetch the updated or created item to return in response
	db.DB.Where("external_id = ?", item.ExternalID).First(&item)

	c.JSON(201, item)
}

// SyncAnimeHandler handles anime sync requests from AniList
func SyncAnimeHandler(c *gin.Context) {
	username := c.Query("username")
	if username == "" {
		c.JSON(400, gin.H{"error": "username query parameter is required"})
		return
	}

	data, err := FetchAniListAnime(username)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	for _, item := range data {
		err := db.UpsertMedia(&item, []string{"title", "status", "progress_current", "updated_at"})
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(200, gin.H{"message": "Sync Complete", "count": len(data)})
}

// SyncMangaHandler handles manga sync requests from AniList
func SyncMangaHandler(c *gin.Context) {
	username := c.Query("username")
	if username == "" {
		c.JSON(400, gin.H{"error": "username query parameter is required"})
		return
	}

	data, err := FetchAniListManga(username)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	for _, item := range data {
		err := db.UpsertMedia(&item, []string{"title", "status", "progress_current", "progress_total", "updated_at"})
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(200, gin.H{"message": "Sync Complete", "count": len(data)})
}

// SearchAnimeHandler handles search requests for AniList anime
func SearchAnimeHandler(c *gin.Context) {
	query := c.Query("query")
	searchCount, _ := strconv.Atoi(c.DefaultQuery("search_count", "10"))

	if query == "" {
		c.JSON(400, gin.H{"error": "query parameter is required"})
		return
	}

	results, err := SearchAnilistAnime(query, searchCount)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
	}

	c.JSON(201, results)
}

// SearchMangaHandler handles search requests for AniList manga
func SearchMangaHandler(c *gin.Context) {
	query := c.Query("query")
	searchCount, _ := strconv.Atoi(c.DefaultQuery("search_count", "10"))

	if query == "" {
		c.JSON(400, gin.H{"error": "query parameter is required"})
		return
	}

	results, err := SearchAnilistManga(query, searchCount)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
	}

	c.JSON(201, results)
}
