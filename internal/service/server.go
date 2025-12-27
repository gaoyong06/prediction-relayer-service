package service

import (
	"github.com/google/wire"
)

// ProviderSet 服务层依赖注入
var ProviderSet = wire.NewSet(
	NewRelayerService,
)


