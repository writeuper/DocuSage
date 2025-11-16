package models

import (
	"docusage/query-service/config"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	DB         *gorm.DB
	dbInstance *gorm.DB
)

// InitDB 初始化数据库连接
func InitDB(cfg *config.Config, log *zap.Logger) error {
	log.Info("Connecting to database",
		zap.String("dsn", maskPassword(cfg.GetDSN())),
		zap.String("host", cfg.DBHost),
		zap.String("database", cfg.DBName),
	)

	// 配置GORM日志
	gormLogger := logger.Default.LogMode(logger.Error)
	if cfg.Debug {
		gormLogger = logger.Default.LogMode(logger.Info)
	}

	// 连接数据库
	db, err := gorm.Open(mysql.Open(cfg.GetDSN()), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		log.Error("Failed to connect to database", zap.Error(err))
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// 获取底层的SQL DB实例以配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		log.Error("Failed to get underlying sql.DB instance", zap.Error(err))
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// 配置连接池
	sqlDB.SetMaxIdleConns(cfg.DBMaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.DBMaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.DBConnMaxLifetime) * time.Minute)
	sqlDB.SetConnMaxIdleTime(time.Duration(cfg.DBConnMaxIdleTime) * time.Minute)

	log.Info("Database connection pool configured",
		zap.Int("max_idle_conns", cfg.DBMaxIdleConns),
		zap.Int("max_open_conns", cfg.DBMaxOpenConns),
		zap.Int("conn_max_lifetime_minutes", int(cfg.DBConnMaxLifetime)),
		zap.Int("conn_max_idle_time_minutes", int(cfg.DBConnMaxIdleTime)),
	)

	// 设置全局DB实例
	dbInstance = db
	DB = db

	// 自动迁移
	if err := migrateDB(log); err != nil {
		return err
	}

	// 进行初始健康检查
	if !CheckDBHealth() {
		return fmt.Errorf("initial database health check failed")
	}

	log.Info("Database connected and migrated successfully")
	return nil
}

// migrateDB 执行数据库迁移
func migrateDB(log *zap.Logger) error {
	log.Info("Running database migrations")

	// 自动迁移所有模型
	err := DB.AutoMigrate(
		&Query{},
		&SearchResult{},
		&SearchHistory{},
		&QueryLog{},
		&QueryStats{},
	)

	if err != nil {
		log.Error("Failed to migrate database", zap.Error(err))
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	log.Info("Database migration completed successfully")
	return nil
}

// GetDB 获取数据库连接
func GetDB() *gorm.DB {
	return DB
}

// CheckDBHealth 检查数据库连接健康状态
func CheckDBHealth() bool {
	if dbInstance == nil {
		return false
	}

	sqlDB, err := dbInstance.DB()
	if err != nil {
		return false
	}

	// 使用Ping检查连接
	if err := sqlDB.Ping(); err != nil {
		return false
	}

	return true
}

// CloseDB 关闭数据库连接
func CloseDB() error {
	if dbInstance == nil {
		return nil
	}

	sqlDB, err := dbInstance.DB()
	if err != nil {
		return err
	}

	return sqlDB.Close()
}

// maskPassword 遮蔽DSN中的密码，用于日志记录
func maskPassword(dsn string) string {
	// 简单的密码遮蔽逻辑，实际项目中可能需要更复杂的解析
	for i := 0; i < len(dsn); i++ {
		if i > 0 && dsn[i-1] == ':' && i < len(dsn)-1 {
			// 查找密码结束位置（@符号）
			for j := i; j < len(dsn); j++ {
				if dsn[j] == '@' {
					return dsn[:i] + "***" + dsn[j:]
				}
			}
			break
		}
	}
	return dsn
}
