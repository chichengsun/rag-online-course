package main

import (
	"flag"
	"path/filepath"
	"strings"

	"rag-online-course/internal/config"
	"rag-online-course/internal/logging"

	"github.com/golang-migrate/migrate/v4"
	"github.com/sirupsen/logrus"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	cfg := config.Load()
	logging.Init(cfg)
	// source 指向迁移文件位置，默认读取项目内 migrations 目录。
	source := flag.String("source", "file://migrations", "migration source")
	// action 控制迁移行为：up 执行、down 回滚、force 强制修正版本号。
	action := flag.String("action", "up", "migration action: up|down|force")
	version := flag.Int("version", 0, "version for force action")
	flag.Parse()

	resolvedSource := resolveMigrationSource(*source)

	m, err := migrate.New(resolvedSource, cfg.Postgres.DSN())
	if err != nil {
		logrus.Fatalf("create migrator failed: %v", err)
	}
	defer func() {
		srcErr, dbErr := m.Close()
		if srcErr != nil {
			logrus.Warnf("close source error: %v", srcErr)
		}
		if dbErr != nil {
			logrus.Warnf("close db error: %v", dbErr)
		}
	}()

	switch *action {
	case "up":
		err = m.Up()
	case "down":
		err = m.Down()
	case "force":
		err = m.Force(*version)
	default:
		logrus.Fatalf("unknown action: %s", *action)
	}

	if err != nil && err != migrate.ErrNoChange {
		logrus.Fatalf("migration failed: %v", err)
	}
	logrus.Infof("migration action %s done", *action)
}

// resolveMigrationSource 将 file:// 的相对路径转换为绝对路径，避免执行目录差异导致找不到迁移文件。
func resolveMigrationSource(src string) string {
	if !strings.HasPrefix(src, "file://") {
		return src
	}
	p := strings.TrimPrefix(src, "file://")
	if p == "" {
		return src
	}
	// 绝对路径直接返回。
	if filepath.IsAbs(p) {
		return src
	}
	abs, err := filepath.Abs(p)
	if err != nil {
		return src
	}
	return "file://" + abs
}
