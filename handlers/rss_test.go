package handlers

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/srt180/mtRSSConverter/config"
	"github.com/srt180/mtRSSConverter/models"
	"github.com/srt180/mtRSSConverter/services"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestDownloadStationRSS(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "test.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.Item{}); err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	t.Cleanup(func() { sqlDB.Close() })
	old := *config.C
	t.Cleanup(func() { *config.C = old })
	config.C.DB = db
	config.C.BaseAddr = "http://converter.example/"
	torrent := "d4:infod4:name4:test6:lengthi1e12:piece lengthi1e6:pieces20:01234567890123456789ee"
	var upstream *httptest.Server
	upstream = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/feed":
			if r.URL.Query().Get("sign") != "secret" {
				t.Error("lost feed signature")
			}
			fmt.Fprintf(w, `<rss version="2.0"><channel><item><title>One</title><link>https://tracker.example/details?id=1</link><enclosure url="%s/redirect"/></item><item><title>Two</title><link>https://tracker.example/details?id=2</link><enclosure url="%s/missing"/></item></channel></rss>`, upstream.URL, upstream.URL)
		case "/redirect":
			http.Redirect(w, r, "/torrent?sign=secret", http.StatusFound)
		case "/torrent":
			if r.URL.Query().Get("sign") != "secret" {
				t.Error("lost torrent signature")
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			fmt.Fprint(w, torrent)
		default:
			http.NotFound(w, r)
		}
	}))
	defer upstream.Close()
	router := gin.New()
	router.GET("/rss/*url", RSS)
	router.GET("/fetch/:guid", Fetch)
	router.HEAD("/fetch/:guid", Fetch)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest("GET", "/rss/"+upstream.URL+"/feed?sign=secret", nil))
	if w.Code != 200 {
		t.Fatalf("feed: %d %s", w.Code, w.Body.String())
	}
	var feed services.RSS
	if err := xml.Unmarshal(w.Body.Bytes(), &feed); err != nil {
		t.Fatal(err)
	}
	if len(feed.Channel.Items) != 2 {
		t.Fatalf("items: %d", len(feed.Channel.Items))
	}
	first, second := feed.Channel.Items[0], feed.Channel.Items[1]
	if first.Link == second.Link {
		t.Fatal("GUID-less items share a mapping")
	}
	if first.Link != first.Enclosure.URL || !strings.HasSuffix(first.Link, ".torrent") || !strings.HasPrefix(first.Link, "http://converter.example/fetch/") || first.Enclosure.Type != "application/x-bittorrent" {
		t.Fatalf("invalid download metadata: %+v", first)
	}
	for _, target := range []string{first.Link, strings.TrimSuffix(first.Link, ".torrent")} {
		for _, method := range []string{"GET", "HEAD"} {
			w := httptest.NewRecorder()
			router.ServeHTTP(w, httptest.NewRequest(method, target, nil))
			if w.Code != 200 || w.Header().Get("Content-Type") != "application/x-bittorrent" || !strings.Contains(w.Header().Get("Content-Disposition"), ".torrent") || w.Header().Get("Content-Length") != strconv.Itoa(len(torrent)) {
				t.Fatalf("%s: %d %v", method, w.Code, w.Header())
			}
			body := torrent
			if method == "HEAD" {
				body = ""
			}
			if w.Body.String() != body {
				t.Fatalf("%s body: %q", method, w.Body.String())
			}
		}
	}
	w = httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest("GET", second.Link, nil))
	if w.Code != http.StatusBadGateway || w.Header().Get("Content-Disposition") != "" {
		t.Fatalf("upstream failure: %d %v", w.Code, w.Header())
	}
}
