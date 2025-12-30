package database

import (
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"sonic_fiber_v2/config"
)

// NewGormDB 创建数据库连接
func NewGormDB(cfg *config.Config, logger *zap.Logger) *gorm.DB {
	// 简化实现，使用SQLite
	db, err := gorm.Open(sqlite.Open(cfg.Database.FilePath), &gorm.Config{})
	if err != nil {
		logger.Fatal("Failed to connect database", zap.Error(err))
	}

	// 自动迁移表结构
	// 在实际项目中，这里应该根据模型进行迁移
	// db.AutoMigrate(&models.Post{}, &models.Category{}, &models.Tag{}, ...)

	return db
}
