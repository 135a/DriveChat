package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nym/go-gateway/internal/models"
	"github.com/nym/go-gateway/pkg/mysql"
	"github.com/nym/go-gateway/pkg/redis"
)

// Blacklist keys in Redis
const (
	KeyBlacklistIPs  = "global:gateway:blacklist:ips"
	KeyBlacklistPaths = "global:gateway:blacklist:paths"
)

type BlacklistReq struct {
	Type  string `json:"type" binding:"required"` // IP, PATH_PREFIX, PATH_EXACT
	Value string `json:"value" binding:"required"`
}

// CreateBlacklist adds a new blacklist rule.
func CreateBlacklist(c *gin.Context) {
	var req BlacklistReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rule := models.BlacklistRule{
		Type:   req.Type,
		Value:  req.Value,
		Status: 1,
	}

	if err := mysql.DB.Create(&rule).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create rule"})
		return
	}

	// Sync to Redis
	updateBlacklistCache(&rule, true)

	c.JSON(http.StatusOK, rule)
}

// GetBlacklist returns all blacklist rules.
func GetBlacklist(c *gin.Context) {
	var rules []models.BlacklistRule
	mysql.DB.Find(&rules)
	c.JSON(http.StatusOK, rules)
}

// DeleteBlacklist removes a blacklist rule.
func DeleteBlacklist(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var rule models.BlacklistRule
	if err := mysql.DB.First(&rule, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Rule not found"})
		return
	}

	updateBlacklistCache(&rule, false)
	mysql.DB.Delete(&rule)

	c.JSON(http.StatusOK, gin.H{"message": "Rule deleted"})
}

func updateBlacklistCache(rule *models.BlacklistRule, add bool) {
	key := ""
	if rule.Type == "IP" {
		key = KeyBlacklistIPs
	} else {
		key = KeyBlacklistPaths
	}

	if add {
		redis.Client.SAdd(redis.Ctx, key, rule.Value)
	} else {
		redis.Client.SRem(redis.Ctx, key, rule.Value)
	}
}
