package server

import (
	"time"

	"lingolift/api/middleware"
	"lingolift/api/routers"
	"lingolift/config"
	"lingolift/pkg/log"

	"github.com/labstack/echo/v4"
	"github.com/tylerb/graceful"
	"go.uber.org/zap"
)

// NewHTTPServer
func NewHTTPServerWithConfig(cfg *config.AppConfig, logger *zap.Logger, listenAddress string) {
	// Initialize the access log to record all HTTP interface requests
	config.AccessLogger = log.NewLogger(func(option *log.Options) {
		option.LogFileDir = cfg.Log.LogFileDir
		option.AppName = cfg.Log.AppName
	})

	e := echo.New()

	// e.Use(middleware.AccessLogger)

	e.Use(middleware.Recover())

	// Manually specified port takes precedence over the port in the configuration file
	if len(listenAddress) > 0 {
		cfg.HTTP.Address = listenAddress
	}

	e.Server.Addr = cfg.HTTP.Address
	e.Server.IdleTimeout = time.Duration(cfg.HTTP.IdleTimeout) * time.Second
	// 从受理一个链接请求开始，到读取一个完整请求报文后结束
	//（HTTP协议的请求报文，可能只有报文头，例如GET，所以，也可以是读取请求报文头后）。
	// 是在net/http的Accept方法中，通过调用SetReadline来设置的。
	e.Server.ReadTimeout = time.Duration(cfg.HTTP.ReadTimeout) * time.Second
	// 从读取请求报文头后开始，到返回响应报文后结束（也可以称为：ServeHTTP生命周期）。
	// 在readRequest方法结束前，通过SetWriteDeadline来设置。
	e.Server.WriteTimeout = time.Duration(cfg.HTTP.WriteTimeout) * time.Second

	e = routers.Load(e)

	if cfg.EnablePProf {
		middleware.Wrap(e)
	}

	err := graceful.ListenAndServe(e.Server, 15*time.Second)
	if err != nil {
		logger.Error("graceful.ListenAndServe error", zap.String("err", err.Error()))
	}
	config.AppLogger.Info(`The HTTP server has been successfully started.`,
		zap.String(`service`, `kms-uic-prometheus`),
		zap.String(`listen`, cfg.HTTP.Address),
	)
}
