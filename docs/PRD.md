# Relayer Service 产品需求文档 (PRD)

## 1. 简介
为了降低用户门槛，我们不要求用户持有 Matic 代币支付 Gas 费。所有上链操作（撮合结算、部署代理、赎回）都通过 Relayer Service 转发。

## 2. 核心功能

### 2.1 接收元交易 (Meta-Tx Ingestion)
- **API**: `POST /relay`
- **输入**:
  - `to`: 目标合约地址 (通常是 CTF Exchange 或 Proxy)
  - `data`: 编码后的函数调用数据
  - `signature`: 用户签名 (不是交易签名，是消息签名)
  - `forwarder`: 转发器合约地址

### 2.2 交易执行器 (Transaction Executor)
- 维护一个或多个 Operator Wallet (带私钥)。
- 收到请求后：
  1. 验证签名有效性。
  2. 估算 Gas Limit。
  3. 获取当前 Gas Price (加权 10% 以加速)。
  4. 使用 Operator 私钥签名以太坊交易。
  5. 广播到 RPC 节点。

### 2.3 状态监控
- 监控 `Pending` 状态的交易。
- 如果超过 30秒未确认，自动发一笔相同 Nonce 但 Gas Price 更高的替换交易 (RBF - Replace By Fee)。

## 3. 运维需求
- **余额监控**: 如果 Operator 钱包 Matic 余额不足，发送 Slack/Telegram 报警。
- **安全**: 私钥必须加密存储 (AWS KMS / HashiCorp Vault)。
