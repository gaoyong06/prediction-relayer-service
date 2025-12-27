package data

import (
	"time"
)

// Transaction 交易记录
type Transaction struct {
	ID              uint64    `gorm:"primaryKey;autoIncrement"`                                     // 主键 ID
	TaskID          string    `gorm:"type:varchar(36);uniqueIndex;not null"`                        // 任务 ID（UUID）
	TxHash          string    `gorm:"type:varchar(66);uniqueIndex"`                                 // 交易哈希（0x 开头的 66 字符）
	BuilderAPIKey   string    `gorm:"type:varchar(255);not null;index:idx_builder_api_key"`         // Builder API Key（用于费用追踪）
	FromAddress     string    `gorm:"type:varchar(42);not null;index:idx_from_address"`             // 发送方地址（Operator 地址）
	ToAddress       string    `gorm:"type:varchar(42);not null"`                                    // 接收方地址（目标合约或转发器）
	TargetContract  string    `gorm:"type:varchar(42);not null"`                                    // 目标合约地址
	TransactionType string    `gorm:"type:varchar(50);not null"`                                    // 交易类型（WALLET_DEPLOYMENT, TOKEN_APPROVAL, CTF_SPLIT 等）
	Data            string    `gorm:"type:text;not null"`                                           // 交易数据（hex 编码的函数调用数据）
	Value           string    `gorm:"type:varchar(78);not null;default:'0x0'"`                      // 交易金额（hex 格式，通常为 "0x0"）
	Signature       string    `gorm:"type:text"`                                                    // 用户签名（消息签名，不是交易签名）
	Forwarder       string    `gorm:"type:varchar(42)"`                                             // 转发器合约地址（可选）
	Nonce           int64     `gorm:"type:bigint;not null"`                                         // 交易 nonce
	GasLimit        int64     `gorm:"type:bigint;not null"`                                         // Gas 限制
	GasPrice        string    `gorm:"type:varchar(78);not null"`                                    // Gas 价格（字符串，支持大整数）
	Status          string    `gorm:"type:varchar(20);not null;default:'PENDING';index:idx_status"` // 交易状态（PENDING, MINED, FAILED, REPLACED）
	BlockNumber     *int64    `gorm:"type:bigint"`                                                  // 区块号（交易被打包后才有值）
	GasUsed         *int64    `gorm:"type:bigint"`                                                  // 实际使用的 Gas（交易被打包后才有值）
	ErrorMessage    string    `gorm:"type:text"`                                                    // 错误信息（交易失败时记录）
	CreatedAt       time.Time `gorm:"autoCreateTime;index:idx_created_at"`                          // 创建时间
	UpdatedAt       time.Time `gorm:"autoUpdateTime;index:idx_status_created_at,priority:2"`        // 更新时间
}

// TableName 指定表名
func (Transaction) TableName() string {
	return "transaction"
}

// Builder Builder 认证信息
type Builder struct {
	ID             uint64    `gorm:"primaryKey;autoIncrement"`                                    // 主键 ID
	APIKey         string    `gorm:"type:varchar(255);uniqueIndex;not null"`                      // API Key（唯一标识）
	SecretHash     string    `gorm:"type:varchar(255);not null"`                                  // Secret 哈希值（加密存储）
	PassphraseHash string    `gorm:"type:varchar(255);not null"`                                  // Passphrase 哈希值（加密存储）
	Name           string    `gorm:"type:varchar(255)"`                                           // Builder 名称（可选）
	Status         string    `gorm:"type:varchar(20);not null;default:'ACTIVE';index:idx_status"` // 状态（ACTIVE, INACTIVE）
	CreatedAt      time.Time `gorm:"autoCreateTime"`                                              // 创建时间
	UpdatedAt      time.Time `gorm:"autoUpdateTime"`                                              // 更新时间
}

// TableName 指定表名
func (Builder) TableName() string {
	return "builder"
}

// BuilderFee Builder 费用统计
type BuilderFee struct {
	ID              uint64    `gorm:"primaryKey;autoIncrement"`                             // 主键 ID
	BuilderAPIKey   string    `gorm:"type:varchar(255);not null;index:idx_builder_api_key"` // Builder API Key
	TransactionType string    `gorm:"type:varchar(50);not null;index:idx_transaction_type"` // 交易类型（用于按类型统计）
	TransactionID   string    `gorm:"type:varchar(36);not null"`                            // 关联 transaction.task_id
	GasUsed         int64     `gorm:"type:bigint;not null"`                                 // Gas 消耗量
	GasPrice        string    `gorm:"type:varchar(78);not null"`                            // Gas 价格（字符串，支持大整数）
	TotalCost       string    `gorm:"type:varchar(78);not null"`                            // 总成本（MATIC，字符串格式）
	CreatedAt       time.Time `gorm:"autoCreateTime;index:idx_created_at"`                  // 创建时间
}

// TableName 指定表名
func (BuilderFee) TableName() string {
	return "builder_fee"
}

// Operator Operator 钱包管理
type Operator struct {
	ID                  uint64    `gorm:"primaryKey;autoIncrement"`                                    // 主键 ID
	Address             string    `gorm:"type:varchar(42);uniqueIndex;not null"`                       // Operator 钱包地址
	PrivateKeyEncrypted string    `gorm:"type:text;not null"`                                          // 私钥（加密存储）
	Status              string    `gorm:"type:varchar(20);not null;default:'ACTIVE';index:idx_status"` // 状态（ACTIVE, INACTIVE）
	BalanceThreshold    string    `gorm:"type:varchar(78);not null;default:'1000000000000000000'"`     // 余额告警阈值（wei，默认 1 MATIC）
	CurrentNonce        int64     `gorm:"type:bigint;not null;default:0"`                              // 当前 nonce（用于交易排序）
	CreatedAt           time.Time `gorm:"autoCreateTime"`                                              // 创建时间
	UpdatedAt           time.Time `gorm:"autoUpdateTime"`                                              // 更新时间
}

// TableName 指定表名
func (Operator) TableName() string {
	return "operator"
}
