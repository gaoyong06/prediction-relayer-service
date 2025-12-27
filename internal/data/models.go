package data

import (
	"time"
)

// Transaction 交易记录
type Transaction struct {
	ID              uint64    `gorm:"primaryKey;autoIncrement"`
	TaskID          string    `gorm:"type:varchar(36);uniqueIndex;not null"` // UUID
	TxHash          string    `gorm:"type:varchar(66);uniqueIndex"`
	BuilderAPIKey   string    `gorm:"type:varchar(255);not null;index:idx_builder_api_key"`
	FromAddress     string    `gorm:"type:varchar(42);not null;index:idx_from_address"`
	ToAddress       string    `gorm:"type:varchar(42);not null"`
	TargetContract  string    `gorm:"type:varchar(42);not null"`
	TransactionType string    `gorm:"type:varchar(50);not null"`
	Data            string    `gorm:"type:text;not null"`
	Value           string    `gorm:"type:varchar(78);not null;default:'0x0'"`
	Signature       string    `gorm:"type:text"`
	Forwarder       string    `gorm:"type:varchar(42)"`
	Nonce           int64     `gorm:"type:bigint;not null"`
	GasLimit        int64     `gorm:"type:bigint;not null"`
	GasPrice        string    `gorm:"type:varchar(78);not null"`
	Status          string    `gorm:"type:varchar(20);not null;default:'PENDING';index:idx_status"`
	BlockNumber     *int64    `gorm:"type:bigint"`
	GasUsed         *int64    `gorm:"type:bigint"`
	ErrorMessage    string    `gorm:"type:text"`
	CreatedAt       time.Time `gorm:"autoCreateTime;index:idx_created_at"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime;index:idx_status_created_at,priority:2"`
}

// TableName 指定表名
func (Transaction) TableName() string {
	return "transactions"
}

// Builder Builder 认证信息
type Builder struct {
	ID             uint64    `gorm:"primaryKey;autoIncrement"`
	APIKey         string    `gorm:"type:varchar(255);uniqueIndex;not null"`
	SecretHash     string    `gorm:"type:varchar(255);not null"` // 加密存储
	PassphraseHash string    `gorm:"type:varchar(255);not null"` // 加密存储
	Name           string    `gorm:"type:varchar(255)"`
	Status         string    `gorm:"type:varchar(20);not null;default:'ACTIVE';index:idx_status"`
	CreatedAt      time.Time `gorm:"autoCreateTime"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime"`
}

// TableName 指定表名
func (Builder) TableName() string {
	return "builders"
}

// BuilderFee Builder 费用统计
type BuilderFee struct {
	ID              uint64    `gorm:"primaryKey;autoIncrement"`
	BuilderAPIKey   string    `gorm:"type:varchar(255);not null;index:idx_builder_api_key"`
	TransactionType string    `gorm:"type:varchar(50);not null;index:idx_transaction_type"`
	TransactionID   string    `gorm:"type:varchar(36);not null"` // 关联 transactions.task_id
	GasUsed         int64     `gorm:"type:bigint;not null"`
	GasPrice        string    `gorm:"type:varchar(78);not null"`
	TotalCost       string    `gorm:"type:varchar(78);not null"` // MATIC 金额（字符串）
	CreatedAt       time.Time `gorm:"autoCreateTime;index:idx_created_at"`
}

// TableName 指定表名
func (BuilderFee) TableName() string {
	return "builder_fees"
}

// Operator Operator 钱包管理
type Operator struct {
	ID                  uint64    `gorm:"primaryKey;autoIncrement"`
	Address             string    `gorm:"type:varchar(42);uniqueIndex;not null"`
	PrivateKeyEncrypted string    `gorm:"type:text;not null"` // 加密存储
	Status              string    `gorm:"type:varchar(20);not null;default:'ACTIVE';index:idx_status"`
	BalanceThreshold    string    `gorm:"type:varchar(78);not null;default:'1000000000000000000'"`
	CurrentNonce        int64     `gorm:"type:bigint;not null;default:0"`
	CreatedAt           time.Time `gorm:"autoCreateTime"`
	UpdatedAt           time.Time `gorm:"autoUpdateTime"`
}

// TableName 指定表名
func (Operator) TableName() string {
	return "operators"
}


