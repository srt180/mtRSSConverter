package main

import (
	"github.com/srt180/mtRSSConverter/config"
	"github.com/srt180/mtRSSConverter/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	config.InitDB()

	r := gin.Default()
	r.GET("/rss/*url", handlers.RSS)
	r.GET("/fetch/:guid", handlers.Fetch)
	r.HEAD("/fetch/:guid", handlers.Fetch)

	r.Run(":8081")
}
