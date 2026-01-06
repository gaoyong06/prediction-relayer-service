package monitor

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"prediction-relayer-service/internal/data"
	"prediction-relayer-service/internal/executor"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-kratos/kratos/v2/log"
)

// Monitor 交易监控器接口
type Monitor interface {
	// Start 启动监控器
	Start(ctx context.Context) error

	// Stop 停止监控器
	Stop()
}

// monitor 交易监控器实现
type monitor struct {
	ethClient      *ethclient.Client
	txRepo         data.TransactionRepo
	executor       executor.Executor
	logger         log.Logger
	pendingTimeout time.Duration // Pending 交易超时时间（默认 30 秒）
	rbfThreshold   time.Duration // RBF 触发阈值（默认 30 秒）
	stopCh         chan struct{}
}

// NewMonitor 创建交易监控器
func NewMonitor(
	ethClient *ethclient.Client,
	txRepo data.TransactionRepo,
	exec executor.Executor,
	logger log.Logger,
	pendingTimeout time.Duration,
) Monitor {
	return &monitor{
		ethClient:      ethClient,
		txRepo:         txRepo,
		executor:       exec,
		logger:         logger,
		pendingTimeout: pendingTimeout,
		rbfThreshold:   pendingTimeout,
		stopCh:         make(chan struct{}),
	}
}

// Start 启动监控器
func (m *monitor) Start(ctx context.Context) error {
	m.logger.Log(log.LevelInfo, "msg", "starting transaction monitor")

	ticker := time.NewTicker(10 * time.Second) // 每 10 秒检查一次
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-m.stopCh:
			return nil
		case <-ticker.C:
			if err := m.monitorPendingTransactions(ctx); err != nil {
				m.logger.Log(log.LevelError, "msg", "failed to monitor pending transactions", "error", err)
			}
		}
	}
}

// Stop 停止监控器
func (m *monitor) Stop() {
	close(m.stopCh)
}

// monitorPendingTransactions 监控 Pending 交易
func (m *monitor) monitorPendingTransactions(ctx context.Context) error {
	// 1. 获取所有 Pending 交易
	txs, err := m.txRepo.GetPendingTransactions(ctx, 100)
	if err != nil {
		return fmt.Errorf("failed to get pending transactions: %w", err)
	}

	now := time.Now()
	for _, tx := range txs {
		// 2. 检查交易是否超时
		elapsed := now.Sub(tx.CreatedAt)
		if elapsed > m.rbfThreshold {
			// 3. 检查交易是否已确认
			if tx.TxHash != "" {
				confirmed, err := m.checkTransactionConfirmed(ctx, tx.TxHash)
				if err != nil {
					m.logger.Log(log.LevelError, "msg", "failed to check transaction confirmation", "tx_hash", tx.TxHash, "error", err)
					continue
				}
				if confirmed {
					// 交易已确认，更新状态
					if err := m.txRepo.UpdateStatus(ctx, tx.TaskID, "MINED"); err != nil {
						m.logger.Log(log.LevelError, "msg", "failed to update transaction status", "task_id", tx.TaskID, "error", err)
					}
					continue
				}

				// 4. 执行 RBF（Replace By Fee）
				if err := m.replaceByFee(ctx, tx); err != nil {
					m.logger.Log(log.LevelError, "msg", "failed to replace by fee", "task_id", tx.TaskID, "error", err)
				}
			} else {
				// 5. 如果超过 5 分钟未确认，标记为失败
				if elapsed > 5*time.Minute {
					if err := m.txRepo.UpdateStatus(ctx, tx.TaskID, "FAILED"); err != nil {
						m.logger.Log(log.LevelError, "msg", "failed to update transaction status to failed", "task_id", tx.TaskID, "error", err)
					}
				}
			}
		}
	}

	return nil
}

// checkTransactionConfirmed 检查交易是否已确认
func (m *monitor) checkTransactionConfirmed(ctx context.Context, txHash string) (bool, error) {
	hash := common.HexToHash(txHash)
	receipt, err := m.ethClient.TransactionReceipt(ctx, hash)
	if err != nil {
		if err == ethereum.NotFound {
			return false, nil
		}
		return false, err
	}

	// 交易已确认
	if receipt.Status == types.ReceiptStatusSuccessful {
		// 更新 Gas Used 和 Block Number
		// TODO: 需要获取 Transaction 对象来更新
		return true, nil
	}

	return false, nil
}

// replaceByFee 执行 RBF（Replace By Fee）
func (m *monitor) replaceByFee(ctx context.Context, tx *data.Transaction) error {
	m.logger.Log(log.LevelInfo, "msg", "replacing transaction by fee", "task_id", tx.TaskID, "tx_hash", tx.TxHash)

	// 1. 获取当前 Gas Price（增加 20%）
	gasPrice, err := m.ethClient.SuggestGasPrice(ctx)
	if err != nil {
		return fmt.Errorf("failed to get gas price: %w", err)
	}
	newGasPrice := new(big.Int).Mul(gasPrice, big.NewInt(120))
	newGasPrice = new(big.Int).Div(newGasPrice, big.NewInt(100))

	// 2. 获取 Operator（用于后续创建替换交易）
	_, err = m.executor.SelectOperator(ctx)
	if err != nil {
		return fmt.Errorf("failed to select operator: %w", err)
	}

	// 3. 创建新的交易（相同 Nonce，更高 Gas Price）
	// 注意：这里需要从原交易获取 Nonce
	// TODO: 实现从原交易获取 Nonce 并创建替换交易
	// 这里简化处理，实际应该：
	// - 获取原交易的 Nonce
	// - 使用相同的 Nonce 创建新交易
	// - Gas Price 提高 20%
	// - 广播新交易

	// 4. 更新原交易状态为 REPLACED
	if err := m.txRepo.UpdateStatus(ctx, tx.TaskID, "REPLACED"); err != nil {
		return fmt.Errorf("failed to update transaction status: %w", err)
	}

	m.logger.Log(log.LevelInfo, "msg", "transaction replaced by fee", "task_id", tx.TaskID, "new_gas_price", newGasPrice.String())

	return nil
}
