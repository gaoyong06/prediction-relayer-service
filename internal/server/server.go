package server

import (
	"context"

	v1 "xinyuan_tech/relayer-service/api/relayer/v1"
	"xinyuan_tech/relayer-service/internal/conf"
	"xinyuan_tech/relayer-service/internal/monitor"
	"xinyuan_tech/relayer-service/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/google/wire"
)

// ProviderSet 服务器依赖注入
var ProviderSet = wire.NewSet(
	NewHTTPServer,
	NewGRPCServer,
	NewMonitorRunner,
)

// NewHTTPServer 创建 HTTP 服务器
func NewHTTPServer(c *conf.Server, relayerService *service.RelayerService, logger log.Logger) *http.Server {
	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
		),
	}
	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}
	srv := http.NewServer(opts...)
	v1.RegisterRelayerHTTPServer(srv, relayerService)
	return srv
}

// NewGRPCServer 创建 gRPC 服务器
func NewGRPCServer(c *conf.Server, relayerService *service.RelayerService, logger log.Logger) *grpc.Server {
	var opts = []grpc.ServerOption{
		grpc.Middleware(
			recovery.Recovery(),
		),
	}
	if c.Grpc.Network != "" {
		opts = append(opts, grpc.Network(c.Grpc.Network))
	}
	if c.Grpc.Addr != "" {
		opts = append(opts, grpc.Address(c.Grpc.Addr))
	}
	if c.Grpc.Timeout != nil {
		opts = append(opts, grpc.Timeout(c.Grpc.Timeout.AsDuration()))
	}
	srv := grpc.NewServer(opts...)
	v1.RegisterRelayerServer(srv, relayerService)
	return srv
}

// NewMonitorRunner 创建监控器运行器
func NewMonitorRunner(m monitor.Monitor, logger log.Logger) *MonitorRunner {
	return &MonitorRunner{
		monitor: m,
		logger:  logger,
	}
}

// MonitorRunner 监控器运行器
type MonitorRunner struct {
	monitor monitor.Monitor
	logger  log.Logger
}

// Start 启动监控器（在应用启动时运行）
func (r *MonitorRunner) Start(ctx context.Context) error {
	go func() {
		if err := r.monitor.Start(ctx); err != nil {
			r.logger.Log(log.LevelError, "msg", "monitor stopped", "error", err)
		}
	}()
	return nil
}

