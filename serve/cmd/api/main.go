package main

import (
	"flag"
	"log"

	"github.com/gin-gonic/gin"

	"mining-pipeline/internal/embedder"
	"mining-pipeline/serve/handler"
	"mining-pipeline/serve/middleware"
)

func main() {
	dataDir := flag.String("data-dir", "data", "data directory")
	port := flag.String("port", "8080", "listen port")
	flag.Parse()

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.APIKey())

	h := &handler.QueryHandler{DataDir: *dataDir, Embed: embedder.NewFromEnv()}
	h.Register(r)

	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	log.Printf("listening :%s data-dir=%s", *port, *dataDir)
	if err := r.Run(":" + *port); err != nil {
		log.Fatal(err)
	}
}
