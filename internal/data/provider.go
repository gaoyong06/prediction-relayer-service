package data

import (
	"github.com/google/wire"
)

// ProviderSet 数据层依赖注入
var ProviderSet = wire.NewSet(
	NewData,
	NewDB,
	NewRedis,
	NewRocketMQ,
	NewTransactionRepo,
	NewBuilderRepo,
	NewBuilderFeeRepo,
	NewOperatorRepo,
)

