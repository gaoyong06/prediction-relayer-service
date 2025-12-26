package nonce

import (
	"context"
	"errors"
	"fmt"

	"xinyuan_tech/relayer-service/internal/data"

	"gorm.io/gorm"
)

// Manager Nonce 管理器接口
type Manager interface {
	// AcquireNonce 获取并锁定 Nonce
	// 使用数据库事务实现原子操作，防止 Nonce 冲突
	AcquireNonce(ctx context.Context, operator string) (uint64, error)

	// ReleaseNonce 释放 Nonce（交易确认后）
	// 注意：Nonce 不需要释放，因为它是严格递增的
	ReleaseNonce(ctx context.Context, operator string, nonce uint64) error

	// GetPendingNonce 获取当前待处理的 Nonce
	GetPendingNonce(ctx context.Context, operator string) (uint64, error)

	// GetCurrentNonce 获取当前链上 Nonce
	GetCurrentNonce(ctx context.Context, operator string) (uint64, error)
}

// manager Nonce 管理器实现
type manager struct {
	db           *gorm.DB
	operatorRepo data.OperatorRepo
	ethClient    interface{} // 以太坊客户端（用于查询链上 Nonce）
}

// NewManager 创建 Nonce 管理器
func NewManager(db *gorm.DB, operatorRepo data.OperatorRepo, ethClient interface{}) Manager {
	return &manager{
		db:           db,
		operatorRepo: operatorRepo,
		ethClient:    ethClient,
	}
}

// AcquireNonce 获取并锁定 Nonce
// 使用数据库事务实现原子操作，防止 Nonce 冲突
// 注意：Nonce 是严格递增的，不需要队列，只需要原子递增
func (m *manager) AcquireNonce(ctx context.Context, operator string) (uint64, error) {
	// 使用数据库事务确保原子性
	var nextNonce uint64
	err := m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 查询当前 Operator 的 Nonce
		var op data.Operator
		if err := tx.Where("address = ?", operator).First(&op).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("operator not found: %s", operator)
			}
			return fmt.Errorf("failed to get operator: %w", err)
		}

		// 2. 原子递增 Nonce（使用数据库的原子操作）
		nextNonce = uint64(op.CurrentNonce) + 1

		// 3. 更新 Nonce
		if err := tx.Model(&op).Update("current_nonce", int64(nextNonce)).Error; err != nil {
			return fmt.Errorf("failed to update nonce: %w", err)
		}

		return nil
	})

	if err != nil {
		return 0, err
	}

	return nextNonce, nil
}

// ReleaseNonce 释放 Nonce（交易确认后）
// 注意：Nonce 是严格递增的，不需要释放
// 这个方法保留用于未来可能的优化（如 Nonce 回滚等）
func (m *manager) ReleaseNonce(ctx context.Context, operator string, nonce uint64) error {
	// Nonce 是严格递增的，不需要释放
	// 如果未来需要支持 Nonce 回滚，可以在这里实现
	return nil
}

// GetPendingNonce 获取当前待处理的 Nonce
func (m *manager) GetPendingNonce(ctx context.Context, operator string) (uint64, error) {
	// 从数据库获取当前 Nonce
	op, err := m.operatorRepo.GetByAddress(ctx, operator)
	if err != nil {
		return 0, fmt.Errorf("failed to get operator: %w", err)
	}
	if op == nil {
		return 0, fmt.Errorf("operator not found: %s", operator)
	}

	// 待处理的 Nonce = 当前 Nonce + 1（下一个要使用的）
	return uint64(op.CurrentNonce) + 1, nil
}

// GetCurrentNonce 获取当前链上 Nonce
func (m *manager) GetCurrentNonce(ctx context.Context, operator string) (uint64, error) {
	// TODO: 实现从链上获取 Nonce（用于初始化或同步）
	// 这里简化处理，从数据库获取
	op, err := m.operatorRepo.GetByAddress(ctx, operator)
	if err != nil {
		return 0, fmt.Errorf("failed to get operator: %w", err)
	}
	if op == nil {
		return 0, fmt.Errorf("operator not found: %s", operator)
	}

	return uint64(op.CurrentNonce), nil
}

