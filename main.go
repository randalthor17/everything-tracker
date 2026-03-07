package main

//go:generate go run github.com/swaggo/swag/cmd/swag@latest init -g main.go -o docs --outputTypes go

import (
	"fmt"
	"os"

	"everythingtracker/anilist"
	"everythingtracker/db"
	_ "everythingtracker/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Everything Tracker API
// @version 1.0
// @description REST API for tracking anime and manga data, with AniList sync and search support.
// @BasePath /

func main() {
	db.InitDatabase()
	err := db.MigrateModels(&anilist.Anime{}, &anilist.Manga{})
	if err != nil {
		panic("failed to migrate database")
	}

	r := gin.Default()
	
	// Add CORS middleware to allow Swagger UI requests
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	})
	
	anilist.RegisterRoutes(r)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	err = r.Run(fmt.Sprintf(":%s", port))
	if err != nil {
		panic("failed to start server")
	}
}
