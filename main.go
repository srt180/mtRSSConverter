package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/srt180/mtRSSConverter/config"
	"github.com/srt180/mtRSSConverter/models"

	"github.com/gin-gonic/gin"
)

var BaseAddr string

func main() {

	config.InitDB()

	BaseAddr = os.Getenv("BASE_ADDR")
	if BaseAddr == "" {
		BaseAddr = "http://localhost:8080" // 默认值
	}

	r := gin.Default()

	r.GET(("/rss/*url"), rss)
	r.GET("/fetch/:guid", fetch)
	r.HEAD("/fetch/:guid", fetch)

	r.Run(":8080")
}

func rss(c *gin.Context) {
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
	newURLPrefix := fmt.Sprintf("%s/fetch/", BaseAddr)

	newBody, err := modifyRSSEnclosureURL(body, newURLPrefix)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Modify RSS failed: " + err.Error()})
		return
	}

	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), newBody)
}

func fetch(c *gin.Context) {

	db := config.C.DB

	// select from database
	guid := c.Param("guid")
	if guid == "" {
		slog.Error("Fetch: Missing guid")
		c.JSON(http.StatusBadRequest, gin.H{"error": "GUID is required"})
		return
	}
	var item models.Item
	if err := db.Where("guid = ?", guid).First(&item).Error; err != nil {
		slog.Error("Fetch: Missing item")
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	}

	rawURL := item.RawURL

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

	c.Header("Content-Disposition", resp.Header.Get("Content-Disposition"))
	c.Header("Content-Type", resp.Header.Get("Content-Type"))
	c.Header("Content-Length", resp.Header.Get("Content-Length"))

	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), body)
}

func modifyRSSEnclosureURL(xmlData []byte, prefix string) ([]byte, error) {
	var rss RSS

	// 解析XML
	err := xml.Unmarshal(xmlData, &rss)
	if err != nil {
		return nil, fmt.Errorf("解析RSS失败: %v", err)
	}

	// 修改每个item的enclosure URL
	for i, item := range rss.Channel.Items {

		db := config.C.DB

		newItem := models.Item{
			GUID:   item.GUID,
			RawURL: rss.Channel.Items[i].Enclosure.URL,
		}
		if err := db.Save(&newItem); err != nil {
			slog.Error("保存Item到数据库失败", "error", err, "item", newItem)
		}

		rss.Channel.Items[i].Enclosure.URL = prefix + item.GUID
	}

	// 重新序列化为XML
	modifiedXML, err := xml.MarshalIndent(rss, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("序列化RSS失败: %v", err)
	}

	return modifiedXML, nil
}

// RSS结构体定义
type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	Channel Channel  `xml:"channel"`
}

type Channel struct {
	Title         string `xml:"title"`
	Link          string `xml:"link"`
	Description   string `xml:"description"`
	Language      string `xml:"language"`
	Copyright     string `xml:"copyright"`
	PubDate       string `xml:"pubDate"`
	LastBuildDate string `xml:"lastBuildDate"`
	TTL           string `xml:"ttl"`
	Docs          string `xml:"docs"`
	Generator     string `xml:"generator"`
	Items         []Item `xml:"item"`
}

type Item struct {
	GUID        string    `xml:"guid"`
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	Author      string    `xml:"author"`
	Comments    string    `xml:"comments"`
	PubDate     string    `xml:"pubDate"`
	Category    Category  `xml:"category"`
	Enclosure   Enclosure `xml:"enclosure"`
}

type Category struct {
	Domain string `xml:"domain,attr"`
	Value  string `xml:",chardata"`
}

type Enclosure struct {
	Length string `xml:"length,attr"`
	Type   string `xml:"type,attr"`
	URL    string `xml:"url,attr"`
}
