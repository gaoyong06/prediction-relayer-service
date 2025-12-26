# Polymarket Relayer Service 对齐分析

## 1. Polymarket Relayer Service 核心功能

根据 Polymarket 官方文档，Relayer Service 提供以下核心功能：

### 1.1 Gasless Transactions（免 Gas 交易）
- **目标**：用户无需持有 MATIC 代币即可执行链上操作
- **覆盖范围**：
  - 部署 Gnosis Safe Wallets 或 Custom Proxy Wallets
  - Token approvals (USDC, outcome tokens)
  - Conditional Token Framework (CTF) 操作：split, merge, redeem
  - CLOB 订单执行

### 1.2 钱包类型支持
- **Safe Wallets**：基于 Gnosis Safe 的代理钱包，需要显式部署
- **Proxy Wallets**：自定义代理合约，首次交易时自动部署

### 1.3 Builder 认证机制
- **认证方式**：HMAC 签名认证
- **凭证组成**：
  - `apiKey`：Builder API key identifier
  - `secret`：用于签名的密钥
  - `passphrase`：额外的认证密码
- **签名算法**：HMAC-SHA256
- **签名内容**：`timestamp + method + path + body`

### 1.4 交易类型支持
1. **Wallet Deployment**：部署用户钱包
2. **Token Approvals**：代币授权
3. **CTF Operations**：
   - Split：拆分代币
   - Merge：合并代币
   - Redeem：赎回代币
4. **CLOB Order Execution**：订单执行

### 1.5 批量交易支持
- 支持在单次调用中批量执行多个操作

## 2. 当前设计对比分析

### 2.1 已实现功能 ✅
- ✅ 元交易接收和验证
- ✅ 交易执行（使用 Operator Wallet）
- ✅ 状态监控和 RBF（Replace By Fee）
- ✅ 余额监控

### 2.2 缺失功能 ⚠️
- ⚠️ **Builder 认证机制**：缺少 HMAC 签名验证
- ⚠️ **交易类型区分**：未明确支持不同类型的交易
- ⚠️ **批量交易**：不支持批量执行
- ⚠️ **钱包部署**：未明确支持 Safe Wallets 和 Proxy Wallets
- ⚠️ **费用追踪**：缺少 Builder 级别的费用统计

## 3. 设计改进建议

### 3.1 添加 Builder 认证
**参考链接**：https://docs.polymarket.com/developers/builders/relayer-client

**实现要点**：
1. 在 API 请求头中添加认证信息：
   - `POLY_BUILDER_SIGNATURE`：HMAC 签名
   - `POLY_BUILDER_TIMESTAMP`：时间戳
   - `POLY_BUILDER_API_KEY`：API Key
   - `POLY_BUILDER_PASSPHRASE`：密码
2. 服务端验证签名：
   ```go
   signature = HMAC-SHA256(secret, timestamp + method + path + body)
   ```
3. 验证时间戳（防止重放攻击，通常允许 5 分钟时间窗口）

### 3.2 支持交易类型
**交易类型枚举**：
- `WALLET_DEPLOYMENT`：钱包部署
- `TOKEN_APPROVAL`：代币授权
- `CTF_SPLIT`：CTF 拆分
- `CTF_MERGE`：CTF 合并
- `CTF_REDEEM`：CTF 赎回
- `CLOB_ORDER`：CLOB 订单执行
- `CUSTOM`：自定义交易

### 3.3 批量交易支持
**API 设计**：
```protobuf
rpc SubmitBatchTransaction (SubmitBatchTransactionRequest) returns (SubmitBatchTransactionReply);

message SubmitBatchTransactionRequest {
  repeated TransactionRequest transactions = 1;
  string builder_api_key = 2; // 用于费用追踪
}
```

### 3.4 钱包部署支持
**Safe Wallet 部署**：
- 需要显式调用 `deploy()` 方法
- 返回部署的 Safe 地址

**Proxy Wallet 部署**：
- 首次交易时自动部署
- 无需显式调用

### 3.5 费用追踪
**数据模型**：
- `builder_id`：Builder 标识
- `transaction_type`：交易类型
- `gas_used`：Gas 消耗
- `gas_price`：Gas 价格
- `total_cost`：总成本（MATIC）
- `timestamp`：时间戳

## 4. API 设计完善

### 4.1 当前 API 问题
1. 缺少 Builder 认证字段
2. 缺少交易类型字段
3. 缺少批量交易接口
4. 缺少钱包部署接口

### 4.2 建议的 API 结构
```protobuf
service Relayer {
  // 提交单笔交易
  rpc SubmitTransaction (SubmitTransactionRequest) returns (SubmitTransactionReply);
  
  // 提交批量交易
  rpc SubmitBatchTransaction (SubmitBatchTransactionRequest) returns (SubmitBatchTransactionReply);
  
  // 部署钱包（Safe Wallet）
  rpc DeployWallet (DeployWalletRequest) returns (DeployWalletReply);
  
  // 查询交易状态
  rpc GetTransactionStatus (GetTransactionStatusRequest) returns (GetTransactionStatusReply);
  
  // 查询 Builder 费用统计
  rpc GetBuilderFeeStats (GetBuilderFeeStatsRequest) returns (GetBuilderFeeStatsReply);
  
  // 查询 Operator 余额
  rpc GetOperatorBalance (GetOperatorBalanceRequest) returns (GetOperatorBalanceReply);
}
```

## 5. 安全考虑

### 5.1 认证安全
- HMAC 签名防止请求篡改
- 时间戳验证防止重放攻击
- API Key 白名单机制

### 5.2 私钥管理
- 使用 AWS KMS 或 HashiCorp Vault 加密存储
- 支持多个 Operator Wallet 轮询使用
- 定期轮换私钥

### 5.3 交易验证
- 验证用户签名有效性
- 验证交易目标合约白名单
- 验证 Gas Limit 合理性

## 6. 性能优化

### 6.1 Nonce 管理
- 使用 Redis 实现串行化队列
- 支持多个 Operator 地址池提高并发

### 6.2 交易监控
- 实时监控 Pending 交易
- 自动 RBF（Replace By Fee）机制
- 交易超时处理

## 7. 监控和告警

### 7.1 余额监控
- Operator Wallet MATIC 余额低于阈值时告警
- 支持 Slack/Telegram 通知

### 7.2 交易监控
- 交易失败率监控
- 平均确认时间监控
- Builder 费用统计

## 8. 参考链接

- [Polymarket Builder Introduction](https://docs.polymarket.com/developers/builders/builder-intro)
- [Polymarket Relayer Client](https://docs.polymarket.com/developers/builders/relayer-client)
- [Polymarket Builder Profile](https://docs.polymarket.com/developers/builders/builder-profile)
