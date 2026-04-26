package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nym/go-gateway/internal/models"
	"github.com/nym/go-gateway/pkg/mysql"
)

// GetBlockLogs returns paginated interception logs.
func GetBlockLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	ip := c.Query("ip")
	path := c.Query("path")

	var logs []models.BlockLog
	var total int64

	query := mysql.DB.Model(&models.BlockLog{})
	if ip != "" {
		query = query.Where("client_ip = ?", ip)
	}
	if path != "" {
		query = query.Where("req_path LIKE ?", "%"+path+"%")
	}

	query.Count(&total)
	query.Offset((page - 1) * pageSize).Limit(pageSize).Order("created_at desc").Find(&logs)

	c.JSON(http.StatusOK, gin.H{
		"total": total,
		"data":  logs,
	})
}
