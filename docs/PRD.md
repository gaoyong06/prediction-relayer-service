# Relayer Service 产品需求文档 (PRD)

## 1. 简介
为了降低用户门槛，我们不要求用户持有 Matic 代币支付 Gas 费。所有上链操作（撮合结算、部署代理、赎回）都通过 Relayer Service 转发。

**参考实现**：完全参考 Polymarket 的 Relayer Service 设计
- [Polymarket Builder Introduction](https://docs.polymarket.com/developers/builders/builder-intro)
- [Polymarket Relayer Client](https://docs.polymarket.com/developers/builders/relayer-client)

## 2. 核心功能

### 2.1 Builder 认证机制
- **认证方式**：HMAC-SHA256 签名认证
- **认证头**：
  - `POLY_BUILDER_SIGNATURE`：HMAC 签名
  - `POLY_BUILDER_TIMESTAMP`：时间戳（Unix 毫秒）
  - `POLY_BUILDER_API_KEY`：Builder API Key
  - `POLY_BUILDER_PASSPHRASE`：密码
- **签名算法**：
  ```
  signature = HMAC-SHA256(secret, timestamp + method + path + body)
  ```
- **安全验证**：
  - 验证时间戳（允许 5 分钟时间窗口，防止重放攻击）
  - 验证 API Key 有效性
  - 验证签名正确性

### 2.2 接收元交易 (Meta-Tx Ingestion)
- **API**: `POST /v1/submit`
- **输入**:
  - `to`: 目标合约地址 (通常是 CTF Exchange 或 Proxy)
  - `data`: 编码后的函数调用数据（hex 格式）
  - `signature`: 用户签名 (消息签名，不是交易签名)
  - `forwarder`: 转发器合约地址（可选）
  - `gas_limit`: 预估 Gas Limit（可选）
  - `transaction_type`: 交易类型（WALLET_DEPLOYMENT, TOKEN_APPROVAL, CTF_SPLIT, CTF_MERGE, CTF_REDEEM, CLOB_ORDER, CUSTOM）
  - `value`: 交易金额（hex，通常为 "0x0"）

### 2.3 交易执行器 (Transaction Executor)
- 维护一个或多个 Operator Wallet (带私钥)。
- 收到请求后：
  1. 验证 Builder 认证（HMAC 签名）
  2. 验证用户签名有效性
  3. 验证目标合约白名单
  4. 估算 Gas Limit（如果未提供）
  5. 获取当前 Gas Price (加权 10% 以加速)
  6. 从 Operator 地址池中选择可用地址
  7. 使用 Operator 私钥签名以太坊交易
  8. 广播到 RPC 节点
  9. 记录交易信息（用于费用追踪）

### 2.4 钱包部署支持
- **Safe Wallet 部署**：
  - API: `POST /v1/wallet/deploy`
  - 显式部署 Gnosis Safe 钱包
  - 返回部署的 Safe 地址
- **Proxy Wallet 部署**：
  - 首次交易时自动部署
  - 无需显式调用部署接口

### 2.5 批量交易支持
- **API**: `POST /v1/submit/batch`
- **输入**:
  - `transactions`: 交易数组
  - `builder_api_key`: Builder API Key（用于费用追踪）
- **处理**：
  - 批量验证签名
  - 批量执行交易
  - 返回批量交易结果

### 2.6 状态监控
- 监控 `Pending` 状态的交易
- 如果超过 30秒未确认，自动发一笔相同 Nonce 但 Gas Price 更高的替换交易 (RBF - Replace By Fee)
- 交易超时处理（超过 5 分钟未确认，标记为失败）

### 2.7 费用追踪
- 记录每个 Builder 的交易费用
- 统计信息：
  - 交易类型分布
  - Gas 消耗统计
  - 总成本统计
- **API**: `GET /v1/builder/fees?api_key={api_key}&start_time={timestamp}&end_time={timestamp}`

## 3. 支持的交易类型

### 3.1 Wallet Deployment（钱包部署）
- 部署 Gnosis Safe Wallets
- 部署 Custom Proxy Wallets
- Gas 费用由 Relayer 承担

### 3.2 Token Approvals（代币授权）
- USDC 代币授权
- Outcome Token 授权
- 用于后续交易操作

### 3.3 CTF Operations（条件代币框架操作）
- **Split**：拆分条件代币
- **Merge**：合并条件代币
- **Redeem**：赎回条件代币

### 3.4 CLOB Order Execution（订单执行）
- 通过 CLOB API 提交的订单执行
- 撮合后的结算交易

### 3.5 Custom Transactions（自定义交易）
- 其他符合白名单的合约调用

## 4. 运维需求

### 4.1 余额监控
- 监控所有 Operator Wallet 的 MATIC 余额
- 如果余额低于阈值（例如 1 MATIC），发送 Slack/Telegram 报警
- 支持自动充值（可选）

### 4.2 安全要求
- **私钥管理**：必须加密存储 (AWS KMS / HashiCorp Vault)
- **API Key 管理**：支持 API Key 的创建、撤销、轮换
- **白名单机制**：只允许调用白名单中的合约地址
- **速率限制**：每个 Builder 的请求速率限制

### 4.3 监控指标
- 交易成功率
- 平均确认时间
- Gas 费用统计
- Builder 使用情况
- Operator 钱包余额

## 5. API 接口设计

### 5.1 提交交易
```
POST /v1/submit
Headers:
  POLY_BUILDER_SIGNATURE: {hmac_signature}
  POLY_BUILDER_TIMESTAMP: {timestamp}
  POLY_BUILDER_API_KEY: {api_key}
  POLY_BUILDER_PASSPHRASE: {passphrase}
Body:
  {
    "to": "0x...",
    "data": "0x...",
    "signature": "0x...",
    "forwarder": "0x...",
    "gas_limit": 100000,
    "transaction_type": "CTF_SPLIT",
    "value": "0x0"
  }
```

### 5.2 批量提交交易
```
POST /v1/submit/batch
Body:
  {
    "transactions": [
      { "to": "...", "data": "...", ... },
      { "to": "...", "data": "...", ... }
    ],
    "builder_api_key": "..."
  }
```

### 5.3 部署钱包
```
POST /v1/wallet/deploy
Body:
  {
    "wallet_type": "SAFE" | "PROXY",
    "owners": ["0x..."]  // 仅 Safe Wallet 需要
  }
```

### 5.4 查询交易状态
```
GET /v1/status/{task_id}
```

### 5.5 查询 Builder 费用统计
```
GET /v1/builder/fees?api_key={api_key}&start_time={timestamp}&end_time={timestamp}
```

### 5.6 查询 Operator 余额
```
GET /v1/operator/balance?address={operator_address}
```

  - `signature`: 用户签名 (消息签名，不是交易签名)
  - `forwarder`: 转发器合约地址（可选）
  - `gas_limit`: 预估 Gas Limit（可选）
  - `transaction_type`: 交易类型（WALLET_DEPLOYMENT, TOKEN_APPROVAL, CTF_SPLIT, CTF_MERGE, CTF_REDEEM, CLOB_ORDER, CUSTOM）
  - `value`: 交易金额（hex，通常为 "0x0"）

### 2.3 交易执行器 (Transaction Executor)
- 维护一个或多个 Operator Wallet (带私钥)。
- 收到请求后：
  1. 验证 Builder 认证（HMAC 签名）
  2. 验证用户签名有效性
  3. 验证目标合约白名单
  4. 估算 Gas Limit（如果未提供）
  5. 获取当前 Gas Price (加权 10% 以加速)
  6. 从 Operator 地址池中选择可用地址
  7. 使用 Operator 私钥签名以太坊交易
  8. 广播到 RPC 节点
  9. 记录交易信息（用于费用追踪）

### 2.4 钱包部署支持
- **Safe Wallet 部署**：
  - API: `POST /v1/wallet/deploy`
  - 显式部署 Gnosis Safe 钱包
  - 返回部署的 Safe 地址
- **Proxy Wallet 部署**：
  - 首次交易时自动部署
  - 无需显式调用部署接口

### 2.5 批量交易支持
- **API**: `POST /v1/submit/batch`
- **输入**:
  - `transactions`: 交易数组
  - `builder_api_key`: Builder API Key（用于费用追踪）
- **处理**：
  - 批量验证签名
  - 批量执行交易
  - 返回批量交易结果

### 2.6 状态监控
- 监控 `Pending` 状态的交易
- 如果超过 30秒未确认，自动发一笔相同 Nonce 但 Gas Price 更高的替换交易 (RBF - Replace By Fee)
- 交易超时处理（超过 5 分钟未确认，标记为失败）

### 2.7 费用追踪
- 记录每个 Builder 的交易费用
- 统计信息：
  - 交易类型分布
  - Gas 消耗统计
  - 总成本统计
- **API**: `GET /v1/builder/fees?api_key={api_key}&start_time={timestamp}&end_time={timestamp}`

## 3. 支持的交易类型

### 3.1 Wallet Deployment（钱包部署）
- 部署 Gnosis Safe Wallets
- 部署 Custom Proxy Wallets
- Gas 费用由 Relayer 承担

### 3.2 Token Approvals（代币授权）
- USDC 代币授权
- Outcome Token 授权
- 用于后续交易操作

### 3.3 CTF Operations（条件代币框架操作）
- **Split**：拆分条件代币
- **Merge**：合并条件代币
- **Redeem**：赎回条件代币

### 3.4 CLOB Order Execution（订单执行）
- 通过 CLOB API 提交的订单执行
- 撮合后的结算交易

### 3.5 Custom Transactions（自定义交易）
- 其他符合白名单的合约调用

## 4. 运维需求

### 4.1 余额监控
- 监控所有 Operator Wallet 的 MATIC 余额
- 如果余额低于阈值（例如 1 MATIC），发送 Slack/Telegram 报警
- 支持自动充值（可选）

### 4.2 安全要求
- **私钥管理**：必须加密存储 (AWS KMS / HashiCorp Vault)
- **API Key 管理**：支持 API Key 的创建、撤销、轮换
- **白名单机制**：只允许调用白名单中的合约地址
- **速率限制**：每个 Builder 的请求速率限制

### 4.3 监控指标
- 交易成功率
- 平均确认时间
- Gas 费用统计
- Builder 使用情况
- Operator 钱包余额

## 5. API 接口设计

### 5.1 提交交易
```
POST /v1/submit
Headers:
  POLY_BUILDER_SIGNATURE: {hmac_signature}
  POLY_BUILDER_TIMESTAMP: {timestamp}
  POLY_BUILDER_API_KEY: {api_key}
  POLY_BUILDER_PASSPHRASE: {passphrase}
Body:
  {
    "to": "0x...",
    "data": "0x...",
    "signature": "0x...",
    "forwarder": "0x...",
    "gas_limit": 100000,
    "transaction_type": "CTF_SPLIT",
    "value": "0x0"
  }
```

### 5.2 批量提交交易
```
POST /v1/submit/batch
Body:
  {
    "transactions": [
      { "to": "...", "data": "...", ... },
      { "to": "...", "data": "...", ... }
    ],
    "builder_api_key": "..."
  }
```

### 5.3 部署钱包
```
POST /v1/wallet/deploy
Body:
  {
    "wallet_type": "SAFE" | "PROXY",
    "owners": ["0x..."]  // 仅 Safe Wallet 需要
  }
```

### 5.4 查询交易状态
```
GET /v1/status/{task_id}
```

### 5.5 查询 Builder 费用统计
```
GET /v1/builder/fees?api_key={api_key}&start_time={timestamp}&end_time={timestamp}
```

### 5.6 查询 Operator 余额
```
GET /v1/operator/balance?address={operator_address}
```

  - `signature`: 用户签名 (消息签名，不是交易签名)
  - `forwarder`: 转发器合约地址（可选）
  - `gas_limit`: 预估 Gas Limit（可选）
  - `transaction_type`: 交易类型（WALLET_DEPLOYMENT, TOKEN_APPROVAL, CTF_SPLIT, CTF_MERGE, CTF_REDEEM, CLOB_ORDER, CUSTOM）
  - `value`: 交易金额（hex，通常为 "0x0"）

### 2.3 交易执行器 (Transaction Executor)
- 维护一个或多个 Operator Wallet (带私钥)。
- 收到请求后：
  1. 验证 Builder 认证（HMAC 签名）
  2. 验证用户签名有效性
  3. 验证目标合约白名单
  4. 估算 Gas Limit（如果未提供）
  5. 获取当前 Gas Price (加权 10% 以加速)
  6. 从 Operator 地址池中选择可用地址
  7. 使用 Operator 私钥签名以太坊交易
  8. 广播到 RPC 节点
  9. 记录交易信息（用于费用追踪）

### 2.4 钱包部署支持
- **Safe Wallet 部署**：
  - API: `POST /v1/wallet/deploy`
  - 显式部署 Gnosis Safe 钱包
  - 返回部署的 Safe 地址
- **Proxy Wallet 部署**：
  - 首次交易时自动部署
  - 无需显式调用部署接口

### 2.5 批量交易支持
- **API**: `POST /v1/submit/batch`
- **输入**:
  - `transactions`: 交易数组
  - `builder_api_key`: Builder API Key（用于费用追踪）
- **处理**：
  - 批量验证签名
  - 批量执行交易
  - 返回批量交易结果

### 2.6 状态监控
- 监控 `Pending` 状态的交易
- 如果超过 30秒未确认，自动发一笔相同 Nonce 但 Gas Price 更高的替换交易 (RBF - Replace By Fee)
- 交易超时处理（超过 5 分钟未确认，标记为失败）

### 2.7 费用追踪
- 记录每个 Builder 的交易费用
- 统计信息：
  - 交易类型分布
  - Gas 消耗统计
  - 总成本统计
- **API**: `GET /v1/builder/fees?api_key={api_key}&start_time={timestamp}&end_time={timestamp}`

## 3. 支持的交易类型

### 3.1 Wallet Deployment（钱包部署）
- 部署 Gnosis Safe Wallets
- 部署 Custom Proxy Wallets
- Gas 费用由 Relayer 承担

### 3.2 Token Approvals（代币授权）
- USDC 代币授权
- Outcome Token 授权
- 用于后续交易操作

### 3.3 CTF Operations（条件代币框架操作）
- **Split**：拆分条件代币
- **Merge**：合并条件代币
- **Redeem**：赎回条件代币

### 3.4 CLOB Order Execution（订单执行）
- 通过 CLOB API 提交的订单执行
- 撮合后的结算交易

### 3.5 Custom Transactions（自定义交易）
- 其他符合白名单的合约调用

## 4. 运维需求

### 4.1 余额监控
- 监控所有 Operator Wallet 的 MATIC 余额
- 如果余额低于阈值（例如 1 MATIC），发送 Slack/Telegram 报警
- 支持自动充值（可选）

### 4.2 安全要求
- **私钥管理**：必须加密存储 (AWS KMS / HashiCorp Vault)
- **API Key 管理**：支持 API Key 的创建、撤销、轮换
- **白名单机制**：只允许调用白名单中的合约地址
- **速率限制**：每个 Builder 的请求速率限制

### 4.3 监控指标
- 交易成功率
- 平均确认时间
- Gas 费用统计
- Builder 使用情况
- Operator 钱包余额

## 5. API 接口设计

### 5.1 提交交易
```
POST /v1/submit
Headers:
  POLY_BUILDER_SIGNATURE: {hmac_signature}
  POLY_BUILDER_TIMESTAMP: {timestamp}
  POLY_BUILDER_API_KEY: {api_key}
  POLY_BUILDER_PASSPHRASE: {passphrase}
Body:
  {
    "to": "0x...",
    "data": "0x...",
    "signature": "0x...",
    "forwarder": "0x...",
    "gas_limit": 100000,
    "transaction_type": "CTF_SPLIT",
    "value": "0x0"
  }
```

### 5.2 批量提交交易
```
POST /v1/submit/batch
Body:
  {
    "transactions": [
      { "to": "...", "data": "...", ... },
      { "to": "...", "data": "...", ... }
    ],
    "builder_api_key": "..."
  }
```

### 5.3 部署钱包
```
POST /v1/wallet/deploy
Body:
  {
    "wallet_type": "SAFE" | "PROXY",
    "owners": ["0x..."]  // 仅 Safe Wallet 需要
  }
```

### 5.4 查询交易状态
```
GET /v1/status/{task_id}
```

### 5.5 查询 Builder 费用统计
```
GET /v1/builder/fees?api_key={api_key}&start_time={timestamp}&end_time={timestamp}
```

### 5.6 查询 Operator 余额
```
GET /v1/operator/balance?address={operator_address}
```
