package postgres

import (
	"context"
	"database/sql"
	"time"

	"rag-online-course/internal/config"

	gormpg "gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// NewDB 根据 Postgres 配置打开 GORM 连接、应用连接池参数并做连通性检查。
func NewDB(ctx context.Context, cfg config.Config) (*gorm.DB, error) {
	// 关闭 GORM 内置 SQL 日志，只将错误回传业务层处理与记录。
	gormDB, err := gorm.Open(gormpg.Open(cfg.Postgres.DSN()), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		return nil, err
	}
	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, err
	}
	applyPoolConfig(sqlDB, cfg.Postgres)
	if err = sqlDB.PingContext(ctx); err != nil {
		_ = sqlDB.Close()
		return nil, err
	}
	return gormDB, nil
}

// CloseDB 关闭底层 sql 连接池，应在进程退出或容器销毁时调用。
func CloseDB(db *gorm.DB) {
	if db == nil {
		return
	}
	sqlDB, err := db.DB()
	if err != nil {
		return
	}
	_ = sqlDB.Close()
}

func applyPoolConfig(sqlDB *sql.DB, cfg config.PostgresConfig) {
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxIdleTime(time.Duration(cfg.ConnMaxIdleMinutes) * time.Minute)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetimeMinutes) * time.Minute)
}
