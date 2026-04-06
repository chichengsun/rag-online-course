package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"rag-online-course/internal/app"
	"rag-online-course/internal/config"
	"rag-online-course/internal/logging"
	"rag-online-course/internal/repository/postgres"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()
	logging.Init(cfg)
	ctx := context.Background()
	// 构建依赖注入容器。
	container := app.BuildContainer(ctx)

	var err error
	// 启动服务并监听退出信号，执行优雅关闭。
	if err = container.Invoke(func(db *gorm.DB, server *http.Server) error {
		defer postgres.CloseDB(db)

		serverErr := make(chan error, 1)
		go func() {
			logrus.Infof("服务监听在 %s", server.Addr)
			if serveErr := server.ListenAndServe(); serveErr != nil && serveErr != http.ErrServerClosed {
				serverErr <- serveErr
			}
		}()

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

		select {
		case sig := <-sigCh:
			logrus.Infof("收到退出信号 %s，开始优雅关闭", sig.String())
		case serveErr := <-serverErr:
			return serveErr
		}

		shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	}); err != nil {
		panic(err)
	}
}
