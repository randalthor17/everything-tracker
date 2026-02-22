package anilist

import "github.com/gin-gonic/gin"

// RegisterRoutes registers all anilist-related routes to the Gin router
func RegisterRoutes(r *gin.Engine) {
	// Items endpoints
	r.GET("/items/anime", GetAnimeHandler)
	r.POST("/items/anime", PostAnimeHandler)

	r.GET("/items/manga", GetMangaHandler)
	r.POST("/items/manga", PostMangaHandler)

	// Sync endpoints
	r.POST("/sync/anilist/anime", SyncAnimeHandler)
	r.POST("/sync/anilist/manga", SyncMangaHandler)
}
