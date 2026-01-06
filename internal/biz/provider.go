package biz

import (
	"github.com/google/wire"
)

// ProviderSet 业务层依赖注入
var ProviderSet = wire.NewSet(
	NewRelayerService,
)




