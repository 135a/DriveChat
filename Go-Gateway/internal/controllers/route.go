package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nym/go-gateway/internal/models"
	"github.com/nym/go-gateway/pkg/mysql"
	"github.com/nym/go-gateway/pkg/redis"
)

// Route keys in Redis
const KeyRoutesHash = "global:gateway:routes"

type RouteReq struct {
	PathPrefix string `json:"path_prefix" binding:"required"`
	TargetURL  string `json:"target_url" binding:"required"`
	Status     int    `json:"status"`
}

// CreateRoute creates a new routing rule.
func CreateRoute(c *gin.Context) {
	var req RouteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	route := models.RouteRule{
		PathPrefix: req.PathPrefix,
		TargetURL:  req.TargetURL,
		Status:     1,
	}

	if err := mysql.DB.Create(&route).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create route"})
		return
	}

	// Update Redis cache
	redis.Client.HSet(redis.Ctx, KeyRoutesHash, route.PathPrefix, route.TargetURL)

	c.JSON(http.StatusOK, route)
}

// GetRoutes returns all routing rules.
func GetRoutes(c *gin.Context) {
	var routes []models.RouteRule
	mysql.DB.Find(&routes)
	c.JSON(http.StatusOK, routes)
}

// UpdateRoute updates an existing routing rule.
func UpdateRoute(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req RouteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var route models.RouteRule
	if err := mysql.DB.First(&route, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Route not found"})
		return
	}

	// If path prefix changes, remove old from Redis
	if route.PathPrefix != req.PathPrefix {
		redis.Client.HDel(redis.Ctx, KeyRoutesHash, route.PathPrefix)
	}

	route.PathPrefix = req.PathPrefix
	route.TargetURL = req.TargetURL
	route.Status = req.Status

	mysql.DB.Save(&route)

	// Update Redis cache
	if route.Status == 1 {
		redis.Client.HSet(redis.Ctx, KeyRoutesHash, route.PathPrefix, route.TargetURL)
	} else {
		redis.Client.HDel(redis.Ctx, KeyRoutesHash, route.PathPrefix)
	}

	c.JSON(http.StatusOK, route)
}

// DeleteRoute deletes a routing rule.
func DeleteRoute(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var route models.RouteRule
	if err := mysql.DB.First(&route, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Route not found"})
		return
	}

	redis.Client.HDel(redis.Ctx, KeyRoutesHash, route.PathPrefix)
	mysql.DB.Delete(&route)

	c.JSON(http.StatusOK, gin.H{"message": "Route deleted"})
}
