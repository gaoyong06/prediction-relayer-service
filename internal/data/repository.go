package data

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

// TransactionRepo 交易仓库接口
type TransactionRepo interface {
	Create(ctx context.Context, tx *Transaction) error
	GetByTaskID(ctx context.Context, taskID string) (*Transaction, error)
	GetByTxHash(ctx context.Context, txHash string) (*Transaction, error)
	UpdateStatus(ctx context.Context, taskID string, status string) error
	UpdateTxHash(ctx context.Context, taskID string, txHash string) error
	UpdateGasUsed(ctx context.Context, taskID string, gasUsed int64, blockNumber int64) error
	GetPendingTransactions(ctx context.Context, limit int) ([]*Transaction, error)
	GetByBuilderAPIKey(ctx context.Context, apiKey string, startTime, endTime time.Time) ([]*Transaction, error)
}

// BuilderRepo Builder 仓库接口
type BuilderRepo interface {
	Create(ctx context.Context, builder *Builder) error
	GetByAPIKey(ctx context.Context, apiKey string) (*Builder, error)
	UpdateStatus(ctx context.Context, apiKey string, status string) error
}

// BuilderFeeRepo Builder 费用仓库接口
type BuilderFeeRepo interface {
	Create(ctx context.Context, fee *BuilderFee) error
	GetStatsByBuilder(ctx context.Context, apiKey string, startTime, endTime time.Time) (*BuilderFeeStats, error)
}

// BuilderFeeStats Builder 费用统计
type BuilderFeeStats struct {
	TotalTransactions int64
	TotalGasUsed      string
	TotalCost         string
	ByType            map[string]*FeeStatsByType
}

// FeeStatsByType 按类型统计的费用
type FeeStatsByType struct {
	Count   int64
	GasUsed string
	Cost    string
}

// OperatorRepo Operator 仓库接口
type OperatorRepo interface {
	Create(ctx context.Context, operator *Operator) error
	GetByAddress(ctx context.Context, address string) (*Operator, error)
	GetActiveOperators(ctx context.Context) ([]*Operator, error)
	UpdateNonce(ctx context.Context, address string, nonce int64) error
	UpdateStatus(ctx context.Context, address string, status string) error
}

// transactionRepo 交易仓库实现
type transactionRepo struct {
	data *Data
}

// NewTransactionRepo 创建交易仓库
func NewTransactionRepo(data *Data) TransactionRepo {
	return &transactionRepo{data: data}
}

func (r *transactionRepo) Create(ctx context.Context, tx *Transaction) error {
	return r.data.db.WithContext(ctx).Create(tx).Error
}

func (r *transactionRepo) GetByTaskID(ctx context.Context, taskID string) (*Transaction, error) {
	var tx Transaction
	err := r.data.db.WithContext(ctx).Where("task_id = ?", taskID).First(&tx).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &tx, nil
}

func (r *transactionRepo) GetByTxHash(ctx context.Context, txHash string) (*Transaction, error) {
	var tx Transaction
	err := r.data.db.WithContext(ctx).Where("tx_hash = ?", txHash).First(&tx).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &tx, nil
}

func (r *transactionRepo) UpdateStatus(ctx context.Context, taskID string, status string) error {
	return r.data.db.WithContext(ctx).
		Model(&Transaction{}).
		Where("task_id = ?", taskID).
		Update("status", status).Error
}

func (r *transactionRepo) UpdateTxHash(ctx context.Context, taskID string, txHash string) error {
	return r.data.db.WithContext(ctx).
		Model(&Transaction{}).
		Where("task_id = ?", taskID).
		Update("tx_hash", txHash).Error
}

func (r *transactionRepo) UpdateGasUsed(ctx context.Context, taskID string, gasUsed int64, blockNumber int64) error {
	return r.data.db.WithContext(ctx).
		Model(&Transaction{}).
		Where("task_id = ?", taskID).
		Updates(map[string]interface{}{
			"gas_used":     gasUsed,
			"block_number": blockNumber,
			"status":       "MINED",
		}).Error
}

func (r *transactionRepo) GetPendingTransactions(ctx context.Context, limit int) ([]*Transaction, error) {
	var txs []*Transaction
	err := r.data.db.WithContext(ctx).
		Where("status = ?", "PENDING").
		Order("created_at ASC").
		Limit(limit).
		Find(&txs).Error
	return txs, err
}

func (r *transactionRepo) GetByBuilderAPIKey(ctx context.Context, apiKey string, startTime, endTime time.Time) ([]*Transaction, error) {
	var txs []*Transaction
	err := r.data.db.WithContext(ctx).
		Where("builder_api_key = ? AND created_at >= ? AND created_at <= ?", apiKey, startTime, endTime).
		Find(&txs).Error
	return txs, err
}

// builderRepo Builder 仓库实现
type builderRepo struct {
	data *Data
}

// NewBuilderRepo 创建 Builder 仓库
func NewBuilderRepo(data *Data) BuilderRepo {
	return &builderRepo{data: data}
}

func (r *builderRepo) Create(ctx context.Context, builder *Builder) error {
	return r.data.db.WithContext(ctx).Create(builder).Error
}

func (r *builderRepo) GetByAPIKey(ctx context.Context, apiKey string) (*Builder, error) {
	var builder Builder
	err := r.data.db.WithContext(ctx).Where("api_key = ?", apiKey).First(&builder).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &builder, nil
}

func (r *builderRepo) UpdateStatus(ctx context.Context, apiKey string, status string) error {
	return r.data.db.WithContext(ctx).
		Model(&Builder{}).
		Where("api_key = ?", apiKey).
		Update("status", status).Error
}

// builderFeeRepo Builder 费用仓库实现
type builderFeeRepo struct {
	data *Data
}

// NewBuilderFeeRepo 创建 Builder 费用仓库
func NewBuilderFeeRepo(data *Data) BuilderFeeRepo {
	return &builderFeeRepo{data: data}
}

func (r *builderFeeRepo) Create(ctx context.Context, fee *BuilderFee) error {
	return r.data.db.WithContext(ctx).Create(fee).Error
}

func (r *builderFeeRepo) GetStatsByBuilder(ctx context.Context, apiKey string, startTime, endTime time.Time) (*BuilderFeeStats, error) {
	var fees []*BuilderFee
	err := r.data.db.WithContext(ctx).
		Where("builder_api_key = ? AND created_at >= ? AND created_at <= ?", apiKey, startTime, endTime).
		Find(&fees).Error
	if err != nil {
		return nil, err
	}

	stats := &BuilderFeeStats{
		TotalTransactions: int64(len(fees)),
		TotalGasUsed:      "0",
		TotalCost:         "0",
		ByType:            make(map[string]*FeeStatsByType),
	}

	// 计算总 Gas 和总成本（需要大整数运算，这里简化处理）
	// 实际应该使用 big.Int 进行计算
	for _, fee := range fees {
		if stats.ByType[fee.TransactionType] == nil {
			stats.ByType[fee.TransactionType] = &FeeStatsByType{
				Count:   0,
				GasUsed: "0",
				Cost:    "0",
			}
		}
		stats.ByType[fee.TransactionType].Count++
		// TODO: 使用 big.Int 进行累加
	}

	return stats, nil
}

// operatorRepo Operator 仓库实现
type operatorRepo struct {
	data *Data
}

// NewOperatorRepo 创建 Operator 仓库
func NewOperatorRepo(data *Data) OperatorRepo {
	return &operatorRepo{data: data}
}

func (r *operatorRepo) Create(ctx context.Context, operator *Operator) error {
	return r.data.db.WithContext(ctx).Create(operator).Error
}

func (r *operatorRepo) GetByAddress(ctx context.Context, address string) (*Operator, error) {
	var operator Operator
	err := r.data.db.WithContext(ctx).Where("address = ?", address).First(&operator).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &operator, nil
}

func (r *operatorRepo) GetActiveOperators(ctx context.Context) ([]*Operator, error) {
	var operators []*Operator
	err := r.data.db.WithContext(ctx).
		Where("status = ?", "ACTIVE").
		Find(&operators).Error
	return operators, err
}

func (r *operatorRepo) UpdateNonce(ctx context.Context, address string, nonce int64) error {
	return r.data.db.WithContext(ctx).
		Model(&Operator{}).
		Where("address = ?", address).
		Update("current_nonce", nonce).Error
}

func (r *operatorRepo) UpdateStatus(ctx context.Context, address string, status string) error {
	return r.data.db.WithContext(ctx).
		Model(&Operator{}).
		Where("address = ?", address).
		Update("status", status).Error
}


