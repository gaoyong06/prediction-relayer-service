package biz

import (
	"context"
	"fmt"
	"time"

	"xinyuan_tech/relayer-service/internal/auth"
	"xinyuan_tech/relayer-service/internal/data"
	"xinyuan_tech/relayer-service/internal/executor"
	"xinyuan_tech/relayer-service/internal/fee"

	"github.com/google/uuid"
)

// RelayerService Relayer 业务服务接口
type RelayerService interface {
	// SubmitTransaction 提交单笔交易
	SubmitTransaction(ctx context.Context, req *SubmitTransactionRequest) (*SubmitTransactionReply, error)

	// SubmitBatchTransaction 提交批量交易
	SubmitBatchTransaction(ctx context.Context, req *SubmitBatchTransactionRequest) (*SubmitBatchTransactionReply, error)

	// GetTransactionStatus 获取交易状态
	GetTransactionStatus(ctx context.Context, taskID string) (*TransactionStatus, error)

	// GetBuilderFeeStats 获取 Builder 费用统计
	GetBuilderFeeStats(ctx context.Context, apiKey string, startTime, endTime time.Time) (*BuilderFeeStats, error)
}

// SubmitTransactionRequest 提交交易请求
type SubmitTransactionRequest struct {
	To              string
	Data            string
	Signature       string
	Forwarder       string
	GasLimit        int64
	TransactionType string
	Value           string
	AuthRequest     *auth.AuthRequest
}

// SubmitTransactionReply 提交交易响应
type SubmitTransactionReply struct {
	TaskID  string
	Success bool
	Message string
}

// SubmitBatchTransactionRequest 批量提交交易请求
type SubmitBatchTransactionRequest struct {
	Transactions  []*SubmitTransactionRequest
	BuilderAPIKey string
}

// SubmitBatchTransactionReply 批量提交交易响应
type SubmitBatchTransactionReply struct {
	TaskIDs []string
	Success bool
	Message string
}

// TransactionStatus 交易状态
type TransactionStatus struct {
	TaskID      string
	TxHash      string
	Status      string
	GasPrice    string
	BlockNumber int64
	GasUsed     int64
	CreatedAt   int64
	UpdatedAt   int64
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

// relayerService Relayer 业务服务实现
type relayerService struct {
	authService auth.AuthService
	txRepo      data.TransactionRepo
	executor    executor.Executor
	feeTracker  fee.Tracker
}

// NewRelayerService 创建 Relayer 业务服务
func NewRelayerService(
	authService auth.AuthService,
	txRepo data.TransactionRepo,
	exec executor.Executor,
	feeTracker fee.Tracker,
) RelayerService {
	return &relayerService{
		authService: authService,
		txRepo:      txRepo,
		executor:    exec,
		feeTracker:  feeTracker,
	}
}

// SubmitTransaction 提交单笔交易
func (s *relayerService) SubmitTransaction(ctx context.Context, req *SubmitTransactionRequest) (*SubmitTransactionReply, error) {
	// 1. 验证 Builder 认证
	builder, err := s.authService.ValidateBuilderAuth(ctx, req.AuthRequest)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// 2. 选择 Operator
	operator, err := s.executor.SelectOperator(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to select operator: %w", err)
	}

	// 3. 创建交易记录
	taskID := uuid.New().String()
	tx := &data.Transaction{
		TaskID:          taskID,
		BuilderAPIKey:   builder.APIKey,
		FromAddress:     operator.Address,
		ToAddress:       req.To,
		TargetContract:  req.To,
		TransactionType: req.TransactionType,
		Data:            req.Data,
		Value:           req.Value,
		Signature:       req.Signature,
		Forwarder:       req.Forwarder,
		GasLimit:        req.GasLimit,
		GasPrice:        "0", // 将在执行时设置
		Status:          "PENDING",
	}

	// 4. 保存交易记录
	if err := s.txRepo.Create(ctx, tx); err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	// 5. 执行交易（异步）
	go func() {
		ctx := context.Background()
		result, err := s.executor.Execute(ctx, tx, operator)
		if err != nil {
			// 更新状态为失败
			s.txRepo.UpdateStatus(ctx, taskID, "FAILED")
			return
		}

		// 更新交易哈希
		if err := s.txRepo.UpdateTxHash(ctx, taskID, result.TxHash); err != nil {
			// 记录错误但不影响主流程
		}
	}()

	return &SubmitTransactionReply{
		TaskID:  taskID,
		Success: true,
		Message: "Transaction submitted",
	}, nil
}

// SubmitBatchTransaction 提交批量交易
func (s *relayerService) SubmitBatchTransaction(ctx context.Context, req *SubmitBatchTransactionRequest) (*SubmitBatchTransactionReply, error) {
	// 1. 验证 Builder 认证（使用第一个交易的认证信息）
	if len(req.Transactions) == 0 {
		return nil, fmt.Errorf("no transactions provided")
	}

	_, err := s.authService.ValidateBuilderAuth(ctx, req.Transactions[0].AuthRequest)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// 2. 批量处理交易
	taskIDs := make([]string, 0, len(req.Transactions))
	for _, txReq := range req.Transactions {
		reply, err := s.SubmitTransaction(ctx, txReq)
		if err != nil {
			// 记录错误但继续处理其他交易
			continue
		}
		taskIDs = append(taskIDs, reply.TaskID)
	}

	return &SubmitBatchTransactionReply{
		TaskIDs: taskIDs,
		Success: true,
		Message: fmt.Sprintf("Submitted %d transactions", len(taskIDs)),
	}, nil
}

// GetTransactionStatus 获取交易状态
func (s *relayerService) GetTransactionStatus(ctx context.Context, taskID string) (*TransactionStatus, error) {
	tx, err := s.txRepo.GetByTaskID(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}
	if tx == nil {
		return nil, fmt.Errorf("transaction not found")
	}

	status := &TransactionStatus{
		TaskID:    tx.TaskID,
		TxHash:    tx.TxHash,
		Status:    tx.Status,
		GasPrice:  tx.GasPrice,
		CreatedAt: tx.CreatedAt.Unix(),
		UpdatedAt: tx.UpdatedAt.Unix(),
	}

	if tx.BlockNumber != nil {
		status.BlockNumber = *tx.BlockNumber
	}
	if tx.GasUsed != nil {
		status.GasUsed = *tx.GasUsed
	}

	return status, nil
}

// GetBuilderFeeStats 获取 Builder 费用统计
func (s *relayerService) GetBuilderFeeStats(ctx context.Context, apiKey string, startTime, endTime time.Time) (*BuilderFeeStats, error) {
	// 从费用仓库获取统计
	stats, err := s.txRepo.GetByBuilderAPIKey(ctx, apiKey, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get fee stats: %w", err)
	}

	// 转换为响应格式
	result := &BuilderFeeStats{
		TotalTransactions: int64(len(stats)),
		TotalGasUsed:      "0",
		TotalCost:         "0",
		ByType:            make(map[string]*FeeStatsByType),
	}

	// TODO: 计算总 Gas 和总成本（需要大整数运算）

	return result, nil
}

