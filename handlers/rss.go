package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/srt180/mtRSSConverter/config"
	"github.com/srt180/mtRSSConverter/models"
	"github.com/srt180/mtRSSConverter/services"
)

// RSS 入口：/rss/*url
func RSS(c *gin.Context) {
	fullURL := c.Request.RequestURI
	prefix := "/rss/"
	if !strings.HasPrefix(fullURL, prefix) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid path"})
		return
	}
	rawURL := fullURL[len(prefix):]

	resp, err := http.Get(rawURL)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Read failed"})
		return
	}

	newURLPrefix := fmt.Sprintf("%s/fetch/", config.C.BaseAddr)
	newBody, err := services.ModifyRSSEnclosureURL(body, newURLPrefix)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Modify RSS failed: " + err.Error()})
		return
	}

	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), newBody)
}

// Fetch 入口：/fetch/:guid
func Fetch(c *gin.Context) {
	db := config.C.DB

	guid := c.Param("guid")
	if guid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "GUID is required"})
		return
	}

	var item models.Item
	if err := db.Where("guid = ?", guid).First(&item).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	}

	resp, err := http.Get(item.RawURL)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Read failed"})
		return
	}

	c.Header("Content-Disposition", resp.Header.Get("Content-Disposition"))
	c.Header("Content-Type", resp.Header.Get("Content-Type"))
	c.Header("Content-Length", resp.Header.Get("Content-Length"))
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), body)
}
