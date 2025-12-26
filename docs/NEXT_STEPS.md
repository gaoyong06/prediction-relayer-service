# Relayer Service 下一步行动

## ✅ 已完成的工作

### 1. 代码生成和编译
- ✅ 生成 API 代码：`make api`
- ✅ 生成配置代码：`make config`
- ✅ 运行 Wire：`cd cmd/server && wire` - 已生成 `wire_gen.go`
- ✅ 编译成功：`go build ./cmd/server` - 无编译错误

### 2. 核心功能实现
- ✅ 服务层实现：所有 gRPC 服务方法已实现
- ✅ Wire 配置：依赖注入配置完成
- ✅ 主程序：`cmd/server/main.go` 已创建
- ✅ Redis 集成：在 `data.go` 中已初始化
- ✅ 以太坊客户端集成：已配置以太坊 RPC 连接
- ✅ RocketMQ 集成：已集成 RocketMQ 生产者（用于异步消息）

### 3. 代码清理
- ✅ 清理了所有文件中的重复代码
- ✅ 修复了所有编译错误
- ✅ 修复了未使用变量的警告

## 📋 下一步行动

### 1. 数据库初始化
```bash
# 创建数据库
mysql -u root -p -e "CREATE DATABASE IF NOT EXISTS relayer_service CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

# 导入表结构
mysql -u root -p relayer_service < docs/db.sql
```

### 2. 配置环境变量
编辑 `configs/config.yaml`，配置以下内容：

- **数据库连接**：`data.database.source`
- **Redis 连接**：`data.redis.addr`
- **RocketMQ 连接**：`data.rocketmq.name_server`
- **以太坊 RPC**：`chain.rpc_url`
- **Chain ID**：`chain.chain_id`
- **Operator 配置**：`operator.address`, `operator.private_key_encrypted`
- **Builder 配置**：`builder.api_key`, `builder.secret_hash`, `builder.passphrase_hash`
- **KMS 配置**：`security.kms_type`, `security.kms_config`

### 3. 初始化数据
在数据库中插入初始数据：

```sql
-- 插入 Operator
INSERT INTO operators (address, private_key_encrypted, status, balance_threshold, current_nonce)
VALUES ('0x...', 'encrypted_private_key', 'ACTIVE', '1000000000000000000', 0);

-- 插入 Builder
INSERT INTO builders (api_key, secret_hash, passphrase_hash, name, status)
VALUES ('your_api_key', 'encrypted_secret', 'encrypted_passphrase', 'Builder Name', 'ACTIVE');
```

### 4. 运行服务
```bash
# 方式 1：直接运行
make run

# 方式 2：构建后运行
make build
./bin/relayer-service -conf ./configs
```

### 5. 测试 API
使用 `api-tester` 或 `curl` 测试各个 API 端点。

## 🔧 待完善的功能（未来工作）

### 1. 钱包部署功能
- [ ] 集成 Gnosis Safe Factory 合约
- [ ] 集成 Proxy Factory 合约
- [ ] 实现完整的钱包部署逻辑

### 2. KMS 集成
- [ ] AWS KMS 集成
- [ ] HashiCorp Vault 集成
- [ ] 在 Executor 中集成 KMS 解密私钥

### 3. Operator 余额查询
- [ ] 实现从以太坊客户端查询余额
- [ ] 实现余额转换为 MATIC 单位

### 4. 其他优化
- [ ] 实现更智能的 Operator 选择策略（基于余额、负载等）
- [ ] 完善 RBF 机制（实现完整的替换交易创建和广播）
- [ ] 实现费用统计的大整数运算
- [ ] 添加更多的监控和日志

## 📝 注意事项

1. **私钥管理**：当前实现中，私钥是直接存储的（未加密）。在生产环境中，必须使用 KMS 进行加密存储和解密。

2. **Nonce 管理**：当前使用数据库事务实现原子操作。如果未来需要更高的并发性能，可以考虑使用 Redis 原子操作。

3. **RocketMQ**：已集成 RocketMQ 用于异步消息，但当前代码中还没有实际使用。可以根据需要添加消息发送逻辑。

4. **监控和告警**：建议添加更多的监控指标和告警机制，确保服务的稳定运行。

