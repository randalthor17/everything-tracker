package main

import (
	"fmt"
	"os"

	"everythingtracker/anilist"
	"everythingtracker/db"

	"github.com/gin-gonic/gin"
)

func main() {
	db.InitDatabase()
	err := db.MigrateModels(&anilist.Anime{}, &anilist.Manga{})
	if err != nil {
		panic("failed to migrate database")
	}

	r := gin.Default()
	anilist.RegisterRoutes(r)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	err = r.Run(fmt.Sprintf(":%s", port))
	if err != nil {
		panic("failed to start server")
	}
}
