package data

import (
	"context"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type DB struct {
	*gorm.DB
}

func NewDB(dsn string) (*DB, error) {
	gdb, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		PrepareStmt:            true, // 可提升频繁相同SQL的性能
		SkipDefaultTransaction: true, // 批量写入性能更好
	})
	if err != nil {
		return nil, err
	}

	// 设置连接池（必须通过 sqlDB 完成）
	sqlDB, err := gdb.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(60 * time.Minute)
	sqlDB.SetConnMaxIdleTime(10 * time.Minute)

	// 启动时健康检查
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, err
	}

	return &DB{gdb}, nil
}

func (db *DB) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
