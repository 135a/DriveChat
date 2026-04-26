package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nym/go-gateway/pkg/redis"
)

// SysConfig keys
const (
	KeyAutoBlacklistEnabled = "global:sys_config:auto_blacklist_enabled"
	KeySlightlyFreqThreshold = "global:sys_config:slightly_freq_threshold"
	KeyTooFreqThreshold      = "global:sys_config:too_freq_threshold"
)

type SysConfigReq struct {
	AutoBlacklistEnabled  bool `json:"auto_blacklist_enabled"`
	SlightlyFreqThreshold int  `json:"slightly_freq_threshold"` // e.g., 50 reqs / min
	TooFreqThreshold      int  `json:"too_freq_threshold"`      // e.g., 100 reqs / min
}

// GetSysConfig retrieves the global parameters from Redis.
func GetSysConfig(c *gin.Context) {
	enabledStr, _ := redis.Client.Get(redis.Ctx, KeyAutoBlacklistEnabled).Result()
	slightlyStr, _ := redis.Client.Get(redis.Ctx, KeySlightlyFreqThreshold).Result()
	tooStr, _ := redis.Client.Get(redis.Ctx, KeyTooFreqThreshold).Result()

	// Default values
	enabled := enabledStr == "1"
	slightly, _ := strconv.Atoi(slightlyStr)
	if slightly == 0 {
		slightly = 50 // default
	}
	too, _ := strconv.Atoi(tooStr)
	if too == 0 {
		too = 100 // default
	}

	c.JSON(http.StatusOK, gin.H{
		"auto_blacklist_enabled":  enabled,
		"slightly_freq_threshold": slightly,
		"too_freq_threshold":      too,
	})
}

// UpdateSysConfig updates the global parameters in Redis.
func UpdateSysConfig(c *gin.Context) {
	var req SysConfigReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	enabledVal := "0"
	if req.AutoBlacklistEnabled {
		enabledVal = "1"
	}

	pipe := redis.Client.Pipeline()
	pipe.Set(redis.Ctx, KeyAutoBlacklistEnabled, enabledVal, 0)
	pipe.Set(redis.Ctx, KeySlightlyFreqThreshold, req.SlightlyFreqThreshold, 0)
	pipe.Set(redis.Ctx, KeyTooFreqThreshold, req.TooFreqThreshold, 0)
	if _, err := pipe.Exec(redis.Ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update sys config"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "System configuration updated successfully"})
}
