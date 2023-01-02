package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sky/common"
	"sky/config"
	"sky/middleware"
	"sky/repository"
	"sky/routes"
	"syscall"
	"time"
)

func main() {
	//初始化配置信息
	config.InitConfig()
	// 初始化日志
	common.InitLogger()
	// 初始化数据库
	if !common.InitDatabase() {
		common.Log.Error("Failed to initialize the database, exit the program.")
		return
	}
	// 初始化casbin策略管理器
	if !common.InitCasbinEnforcer() {
		common.Log.Error("Failed to initialize Policy Manager, exit the program.")
		return
	}

	// 初始化Validator数据校验
	common.InitValidate()

	// 初始化数据库数据
	common.InitData()

	// 操作日志中间件处理日志时没有将日志发送到rabbitmq或者kafka中, 而是发送到了channel中
	// 这里开启3个goroutine处理channel将日志记录到数据库
	logRepository := repository.NewOperationLogRepository()
	for i := 0; i < 3; i++ {
		go logRepository.SaveOperationLogChannel(middleware.OperationLogChan)
	}

	// 注册所有路由
	r := routes.InitRoutes()

	host := "localhost"
	port := config.Conf.System.Port

	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", host, port),
		Handler: r,
	}

	// Initializing the server in a goroutine so that
	// it won't block the graceful shutdown handling below
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			common.Log.Fatalf("listen: %s\n", err)
		}
	}()

	common.Log.Info(fmt.Sprintf("Server is running at %s:%d/%s", host, port, config.Conf.System.UrlPathPrefix))

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	common.Log.Info("Shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		common.Log.Fatal("Server forced to shutdown:", err)
	}

	common.Log.Info("Server exiting!")

}
