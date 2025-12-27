package fee

import (
	"context"
	"fmt"
	"math/big"

	"xinyuan_tech/relayer-service/internal/data"
)

// Tracker 费用追踪器接口
type Tracker interface {
	// RecordFee 记录交易费用
	RecordFee(ctx context.Context, tx *data.Transaction, gasUsed uint64) error

	// CalculateCost 计算交易成本
	CalculateCost(gasUsed uint64, gasPrice string) (string, error)
}

// tracker 费用追踪器实现
type tracker struct {
	feeRepo data.BuilderFeeRepo
}

// NewTracker 创建费用追踪器
func NewTracker(feeRepo data.BuilderFeeRepo) Tracker {
	return &tracker{
		feeRepo: feeRepo,
	}
}

// RecordFee 记录交易费用
func (t *tracker) RecordFee(ctx context.Context, tx *data.Transaction, gasUsed uint64) error {
	// 1. 计算总成本
	cost, err := t.CalculateCost(gasUsed, tx.GasPrice)
	if err != nil {
		return fmt.Errorf("failed to calculate cost: %w", err)
	}

	// 2. 创建费用记录
	fee := &data.BuilderFee{
		BuilderAPIKey:   tx.BuilderAPIKey,
		TransactionType: tx.TransactionType,
		TransactionID:   tx.TaskID,
		GasUsed:         int64(gasUsed),
		GasPrice:        tx.GasPrice,
		TotalCost:       cost,
	}

	// 3. 保存到数据库
	if err := t.feeRepo.Create(ctx, fee); err != nil {
		return fmt.Errorf("failed to create fee record: %w", err)
	}

	return nil
}

// CalculateCost 计算交易成本（MATIC）
// cost = gasUsed * gasPrice / 10^18
func (t *tracker) CalculateCost(gasUsed uint64, gasPriceStr string) (string, error) {
	// 解析 Gas Price
	gasPrice, ok := new(big.Int).SetString(gasPriceStr, 10)
	if !ok {
		// 尝试 hex 格式
		if len(gasPriceStr) >= 2 && gasPriceStr[:2] == "0x" {
			gasPrice, ok = new(big.Int).SetString(gasPriceStr[2:], 16)
			if !ok {
				return "", fmt.Errorf("invalid gas price format: %s", gasPriceStr)
			}
		} else {
			return "", fmt.Errorf("invalid gas price format: %s", gasPriceStr)
		}
	}

	// 计算成本：gasUsed * gasPrice
	cost := new(big.Int).Mul(big.NewInt(int64(gasUsed)), gasPrice)

	// 转换为 MATIC（除以 10^18）
	divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	cost = new(big.Int).Div(cost, divisor)

	// 返回字符串格式（保留 6 位小数）
	// 这里简化处理，返回整数部分
	return cost.String(), nil
}


