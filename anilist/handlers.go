package anilist

import (
	"strconv"

	"everythingtracker/db"

	"github.com/gin-gonic/gin"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type SyncResponse struct {
	Message string `json:"message"`
	Count   int    `json:"count"`
}

// GetAnimeHandler godoc
// @Summary Get all anime items for a user
// @Description Returns all anime items for the specified user.
// @Tags items
// @Produce json
// @Param username query string true "Username to filter anime items"
// @Success 200 {array} Anime
// @Failure 400 {object} ErrorResponse
// @Router /items/anime [get]
// GetAnimeHandler handles GET requests for anime items
func GetAnimeHandler(c *gin.Context) {
	username := c.Query("username")
	if username == "" {
		c.JSON(400, gin.H{"error": "username query parameter is required"})
		return
	}

	var items []Anime
	db.DB.Where("username = ?", username).Find(&items)
	c.JSON(200, items)
}

// GetMangaHandler godoc
// @Summary Get all manga items for a user
// @Description Returns all manga items for the specified user.
// @Tags items
// @Produce json
// @Param username query string true "Username to filter manga items"
// @Success 200 {array} Manga
// @Failure 400 {object} ErrorResponse
// @Router /items/manga [get]
// GetMangaHandler handles GET requests for manga items
func GetMangaHandler(c *gin.Context) {
	username := c.Query("username")
	if username == "" {
		c.JSON(400, gin.H{"error": "username query parameter is required"})
		return
	}

	var items []Manga
	db.DB.Where("username = ?", username).Find(&items)
	c.JSON(200, items)
}

// PostAnimeHandler godoc
// @Summary Create or update an anime item
// @Description Upserts an anime item using username and external_id as the unique key.
// @Tags items
// @Accept json
// @Produce json
// @Param item body Anime true "Anime item payload (must include username)"
// @Success 201 {object} Anime
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /items/anime [post]
// PostAnimeHandler handles POST requests for anime items
func PostAnimeHandler(c *gin.Context) {
	var item Anime
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if item.Username == "" {
		c.JSON(400, gin.H{"error": "username is required"})
		return
	}

	// upsert logic
	err := db.UpsertMedia(&item, []string{"title", "status", "progress_current", "updated_at"})
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// fetch the updated or created item to return in response
	db.DB.Where("username = ? AND external_id = ?", item.Username, item.ExternalID).First(&item)

	c.JSON(201, item)
}

// PostMangaHandler godoc
// @Summary Create or update a manga item
// @Description Upserts a manga item using username and external_id as the unique key.
// @Tags items
// @Accept json
// @Produce json
// @Param item body Manga true "Manga item payload (must include username)"
// @Success 201 {object} Manga
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /items/manga [post]
// PostMangaHandler handles POST requests for manga items
func PostMangaHandler(c *gin.Context) {
	var item Manga
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if item.Username == "" {
		c.JSON(400, gin.H{"error": "username is required"})
		return
	}

	// upsert logic
	err := db.UpsertMedia(&item, []string{"title", "status", "progress_current", "updated_at"})
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// fetch the updated or created item to return in response
	db.DB.Where("username = ? AND external_id = ?", item.Username, item.ExternalID).First(&item)

	c.JSON(201, item)
}

// SyncAnimeHandler godoc
// @Summary Sync anime from AniList
// @Description Fetches a user's anime list from AniList and upserts all entries into the local database.
// @Tags sync
// @Produce json
// @Param username query string true "AniList username"
// @Success 200 {object} SyncResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /sync/anilist/anime [post]
// SyncAnimeHandler handles anime sync requests from AniList
func SyncAnimeHandler(c *gin.Context) {
	username := c.Query("username")
	if username == "" {
		c.JSON(400, gin.H{"error": "username query parameter is required"})
		return
	}

	data, err := FetchAniListAnime(username)
	if err != nil {
		c.JSON(500, gin.H{
			"error":   err.Error(),
			"message": "Failed to fetch anime list from AniList. Please verify the username exists and the anime list is set to public.",
		})
		return
	}

	for i := range data {
		data[i].Username = username
		err := db.UpsertMedia(&data[i], []string{"title", "status", "progress_current", "updated_at"})
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(200, gin.H{"message": "Sync Complete", "count": len(data)})
}

// SyncMangaHandler godoc
// @Summary Sync manga from AniList
// @Description Fetches a user's manga list from AniList and upserts all entries into the local database.
// @Tags sync
// @Produce json
// @Param username query string true "AniList username"
// @Success 200 {object} SyncResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /sync/anilist/manga [post]
// SyncMangaHandler handles manga sync requests from AniList
func SyncMangaHandler(c *gin.Context) {
	username := c.Query("username")
	if username == "" {
		c.JSON(400, gin.H{"error": "username query parameter is required"})
		return
	}

	data, err := FetchAniListManga(username)
	if err != nil {
		c.JSON(500, gin.H{
			"error":   err.Error(),
			"message": "Failed to fetch manga list from AniList. Please verify the username exists and the manga list is set to public.",
		})
		return
	}

	for i := range data {
		data[i].Username = username
		err := db.UpsertMedia(&data[i], []string{"title", "status", "progress_current", "progress_total", "updated_at"})
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(200, gin.H{"message": "Sync Complete", "count": len(data)})
}

// SearchAnimeHandler godoc
// @Summary Search AniList anime
// @Description Searches AniList anime by query string.
// @Tags search
// @Produce json
// @Param query query string true "Search query"
// @Param search_count query int false "Maximum number of results" default(10)
// @Success 200 {array} Anime
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /search/anilist/anime [get]
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
		return
	}

	c.JSON(200, results)
}

// SearchMangaHandler godoc
// @Summary Search AniList manga
// @Description Searches AniList manga by query string.
// @Tags search
// @Produce json
// @Param query query string true "Search query"
// @Param search_count query int false "Maximum number of results" default(10)
// @Success 200 {array} Manga
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /search/anilist/manga [get]
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
		return
	}

	c.JSON(200, results)
}
