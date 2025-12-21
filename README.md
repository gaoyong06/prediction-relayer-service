# Relayer Service (交易转发服务)

Relayer Service 是实现 Gasless Trading (无 Gas 交易) 的基础设施。它充当“元交易 (Meta-Transaction)”的中继者，接收用户签名的指令，并由平台控制的 EOA (外部拥有账户) 支付 Matic 将交易上链。

## 特性
- **EIP-2771 / Gnosis Safe 兼容**: 支持多种元交易标准。
- **Nonce 管理**: 本地管理 Nonce 队列，防止交易阻塞。
- **Gas 估算与调整**: 动态调整 Gas Price 以确保交易及时确认。
- **交易重发**: 自动检测卡顿交易并加速 (Speed Up)。

## 技术栈
- **语言**: Go / Node.js (OpenZeppelin Defender 风格)
- **库**: `go-ethereum`
- **存储**: Redis (Nonce 锁), MySQL (交易记录)
