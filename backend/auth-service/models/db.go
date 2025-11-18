package models

import (
	"docusage/auth-service/config"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDB 初始化数据库连接
func InitDB(cfg *config.Config, log *zap.Logger) error {
	// 设置GORM日志
	gormLogger := logger.Default.LogMode(logger.Error)
	if cfg.LogLevel == "debug" {
		gormLogger = logger.Default.LogMode(logger.Info)
	}

	// 连接数据库
	db, err := gorm.Open(mysql.Open(cfg.GetDSN()), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		log.Error("Failed to connect to database", zap.Error(err))
		return err
	}

	// 设置全局DB变量
	DB = db

	// 自动迁移数据库表
	err = migrateDB(log)
	if err != nil {
		log.Error("Failed to migrate database", zap.Error(err))
		return err
	}

	// 初始化默认数据
	err = initDefaultData(log)
	if err != nil {
		log.Error("Failed to initialize default data", zap.Error(err))
		return err
	}

	log.Info("Database initialized successfully")
	return nil
}

// migrateDB 迁移数据库表
func migrateDB(log *zap.Logger) error {
	err := DB.AutoMigrate(
		&User{},
		&Role{},
		&Permission{},
	)
	if err != nil {
		log.Error("Database migration failed", zap.Error(err))
		return err
	}

	log.Info("Database migration completed")
	return nil
}

// initDefaultData 初始化默认数据
func initDefaultData(log *zap.Logger) error {
	// 初始化默认角色
	roles := []Role{
		{Name: "admin", Description: "系统管理员", Permissions: "{\"all\":true}"},
		{Name: "user", Description: "普通用户", Permissions: "{\"read\":true,\"write\":true}"},
	}

	for _, role := range roles {
		var existingRole Role
		result := DB.Where("name = ?", role.Name).First(&existingRole)
		if result.Error == gorm.ErrRecordNotFound {
			if err := DB.Create(&role).Error; err != nil {
				log.Error("Failed to create default role", zap.String("role", role.Name), zap.Error(err))
				return err
			}
			log.Info("Created default role", zap.String("role", role.Name))
		}
	}

	// 检查是否存在默认管理员用户
	var adminCount int64
	DB.Model(&User{}).Where("role = ?", "admin").Count(&adminCount)

	if adminCount == 0 {
		// 创建默认管理员用户
		adminUser := User{
			Username:     "admin",
			Email:        "admin@example.com",
			PasswordHash: "$2a$10$VZ1vU0KjYq0QZQZqZQZQZeQZQZQZQZQZQZQZQZQZQZQZQZQZQZQ", // admin123
			FullName:     "System Administrator",
			Role:         "admin",
			Status:       "active",
		}

		if err := DB.Create(&adminUser).Error; err != nil {
			log.Error("Failed to create default admin user", zap.Error(err))
			return err
		}
		log.Info("Created default admin user")
	}

	return nil
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return DB
}
