package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/nym/go-gateway/internal/config"
	"github.com/nym/go-gateway/internal/routes"
	"github.com/nym/go-gateway/internal/services"
	"github.com/nym/go-gateway/pkg/mysql"
	"github.com/nym/go-gateway/pkg/redis"
)

func main() {
	// 1. Load Configuration
	config.LoadConfig()

	// 2. Initialize Infrastructure
	mysql.InitDB()
	redis.InitRedis()

	// 3. Initialize Route Sync (load routes into Trie + start Pub/Sub listener)
	services.InitRouteSync()

	// 4. Setup Gin
	r := gin.Default()

	// 5. Register Routes
	routes.InitRoutes(r)

	// 6. Start Server
	log.Println("Go Gateway is running on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
