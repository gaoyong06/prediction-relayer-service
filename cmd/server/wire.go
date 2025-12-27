//go:build wireinject
// +build wireinject

package main

import (
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"gorm.io/gorm"

	"xinyuan_tech/relayer-service/internal/auth"
	"xinyuan_tech/relayer-service/internal/biz"
	"xinyuan_tech/relayer-service/internal/conf"
	"xinyuan_tech/relayer-service/internal/data"
	"xinyuan_tech/relayer-service/internal/executor"
	"xinyuan_tech/relayer-service/internal/fee"
	"xinyuan_tech/relayer-service/internal/kms"
	"xinyuan_tech/relayer-service/internal/monitor"
	"xinyuan_tech/relayer-service/internal/nonce"
	"xinyuan_tech/relayer-service/internal/server"
	"xinyuan_tech/relayer-service/internal/service"
)

func wireApp(c *conf.Bootstrap, logger log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(
		server.ProviderSet,
		service.ProviderSet,
		data.ProviderSet,
		biz.ProviderSet,
		NewEthClient,
		NewChainID,
		NewAuthService,
		NewNonceManager,
		NewExecutor,
		NewFeeTracker,
		NewMonitor,
		wire.FieldsOf(new(*conf.Bootstrap), "Server", "Data", "Chain", "Builder"),
		newApp,
	))
}

// NewEthClient 创建以太坊客户端
func NewEthClient(c *conf.Chain) (*ethclient.Client, func(), error) {
	client, err := ethclient.Dial(c.RpcUrl)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to ethereum client: %w", err)
	}
	cleanup := func() {
		if client != nil {
			client.Close()
		}
	}
	return client, cleanup, nil
}

// NewChainID 创建 Chain ID
func NewChainID(c *conf.Chain) (*big.Int, error) {
	chainID, err := strconv.ParseInt(c.ChainId, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse chain ID: %w", err)
	}
	return big.NewInt(chainID), nil
}

// NewAuthService 创建认证服务
func NewAuthService(
	builderRepo data.BuilderRepo,
	c *conf.Builder,
) auth.AuthService {
	timestampWindow := int64(5 * 60 * 1000) // 默认 5 分钟
	if c != nil && c.TimestampWindowMs > 0 {
		timestampWindow = c.TimestampWindowMs
	}
	return auth.NewAuthService(builderRepo, timestampWindow)
}

// NewNonceManager 创建 Nonce 管理器
func NewNonceManager(
	db *gorm.DB,
	operatorRepo data.OperatorRepo,
	ethClient *ethclient.Client,
) nonce.Manager {
	return nonce.NewManager(db, operatorRepo, ethClient)
}

// NewExecutor 创建交易执行器
func NewExecutor(
	ethClient *ethclient.Client,
	chainID *big.Int,
	nonceMgr nonce.Manager,
	operatorRepo data.OperatorRepo,
	c *conf.Chain,
) executor.Executor {
	gasMultiplier := int64(110) // 默认 110%
	if c != nil && c.GasPriceMultiplier > 0 {
		gasMultiplier = c.GasPriceMultiplier
	}
	return executor.NewExecutor(ethClient, chainID, nonceMgr, operatorRepo, gasMultiplier)
}

// NewFeeTracker 创建费用追踪器
func NewFeeTracker(feeRepo data.BuilderFeeRepo) fee.Tracker {
	return fee.NewTracker(feeRepo)
}

// NewMonitor 创建交易监控器
func NewMonitor(
	ethClient *ethclient.Client,
	txRepo data.TransactionRepo,
	exec executor.Executor,
	logger log.Logger,
) monitor.Monitor {
	pendingTimeout := 30 * time.Second // 默认 30 秒
	return monitor.NewMonitor(ethClient, txRepo, exec, logger, pendingTimeout)
}

// NewKMS 创建 KMS 服务
func NewKMS(c *conf.Security) (kms.KMS, error) {
	kmsType := "local"
	if c != nil && c.KmsType != "" {
		kmsType = c.KmsType
	}
	kmsConfig := ""
	if c != nil {
		kmsConfig = c.KmsConfig
	}
	return kms.NewKMS(kmsType, kmsConfig)
}


