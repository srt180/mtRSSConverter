package services

import (
	"crypto/sha256"
	"encoding/xml"
	"fmt"
	"net/url"
	"strings"

	"github.com/srt180/mtRSSConverter/config"
	"github.com/srt180/mtRSSConverter/models"
)

// RSS 及其相关结构体用于解析与重建 RSS XML
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

// ModifyRSSEnclosureURL 解析 RSS，保存条目到数据库，并把 enclosure URL 替换为本服务的 fetch 地址
func ModifyRSSEnclosureURL(xmlData []byte, newURLPrefix string) ([]byte, error) {
	var rss RSS

	if err := xml.Unmarshal(xmlData, &rss); err != nil {
		return nil, fmt.Errorf("解析RSS失败: %v", err)
	}

	for i, item := range rss.Channel.Items {
		rawURL := strings.TrimSpace(item.Enclosure.URL)
		if rawURL == "" {
			return nil, fmt.Errorf("RSS item %d is missing an enclosure URL", i+1)
		}
		identity := strings.TrimSpace(item.GUID)
		if identity == "" {
			identity = stableItemURL(item.Link)
		}
		if identity == "" {
			identity = stableItemURL(rawURL)
		}
		sum := sha256.Sum256([]byte(newURLPrefix + "\x00" + identity))
		mappingID := fmt.Sprintf("%x", sum[:16])
		newItem := models.Item{GUID: mappingID, RawURL: rawURL}
		if err := config.C.DB.Save(&newItem).Error; err != nil {
			return nil, fmt.Errorf("save torrent mapping: %w", err)
		}
		torrentURL := newURLPrefix + mappingID + ".torrent"
		rss.Channel.Items[i].Link = torrentURL
		rss.Channel.Items[i].Enclosure.URL = torrentURL
		rss.Channel.Items[i].Enclosure.Type = "application/x-bittorrent"
	}

	modifiedXML, err := xml.MarshalIndent(rss, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("序列化RSS失败: %v", err)
	}
	return modifiedXML, nil
}

func stableItemURL(rawURL string) string {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return ""
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	query := parsed.Query()
	for key := range query {
		switch strings.ToLower(key) {
		case "t", "sign", "token", "timestamp", "_t", "nonce", "r":
			query.Del(key)
		}
	}
	parsed.RawQuery = query.Encode()
	parsed.Fragment = ""
	return parsed.String()
}
