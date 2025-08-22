package config

import (
	"log/slog"
	"os"

	"github.com/srt180/mtRSSConverter/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitDB() {

	// Initialize DB
	var err error

	C.DB, err = gorm.Open(sqlite.Open(C.SQLitePath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		slog.Error("failed to connect sqlite database", "error", err)
		os.Exit(1)
	}

	// init database tables
	if err := C.DB.AutoMigrate(&models.Item{}); err != nil {
		slog.Error("failed to auto migrate Party model", "error", err)
	}

}
