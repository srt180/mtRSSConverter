package services

import (
	"encoding/xml"
	"fmt"

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
		newItem := models.Item{
			GUID:   item.GUID,
			RawURL: rss.Channel.Items[i].Enclosure.URL,
		}
		if err := config.C.DB.Save(&newItem).Error; err != nil {
			// 忽略单条错误，继续处理其它条目
		}

		rss.Channel.Items[i].Enclosure.URL = newURLPrefix + item.GUID
	}

	modifiedXML, err := xml.MarshalIndent(rss, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("序列化RSS失败: %v", err)
	}
	return modifiedXML, nil
}
