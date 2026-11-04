package handlers

import (
	"io"
	"mime"
	"net/http"
	"strconv"
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

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		c.JSON(http.StatusBadGateway, gin.H{"error": "RSS upstream returned an error"})
		return
	}
	newURLPrefix := strings.TrimRight(config.C.BaseAddr, "/") + "/fetch/"
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

	guid := strings.TrimSuffix(c.Param("guid"), ".torrent")
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

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Torrent upstream returned an error"})
		return
	}
	c.Header("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": guid + ".torrent"}))
	c.Header("Content-Type", "application/x-bittorrent")
	c.Header("Content-Length", strconv.Itoa(len(body)))
	if c.Request.Method == http.MethodHead {
		c.Status(http.StatusOK)
		return
	}
	c.Data(http.StatusOK, "application/x-bittorrent", body)
}
