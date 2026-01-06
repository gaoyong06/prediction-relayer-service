package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"prediction-relayer-service/internal/auth"
	"prediction-relayer-service/internal/data"
	"prediction-relayer-service/internal/executor"
	"prediction-relayer-service/internal/fee"

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

	// SubmitMatch 提交订单匹配结果
	SubmitMatch(ctx context.Context, req *SubmitMatchRequest) (*SubmitMatchReply, error)

	// GetTransactionHashByOrderID 根据订单 ID 获取交易哈希
	GetTransactionHashByOrderID(ctx context.Context, orderID string) (string, error)
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

// SubmitMatchRequest 提交匹配请求
type SubmitMatchRequest struct {
	MakerOrder *MatchOrder
	TakerOrder *MatchOrder
	Price      string // 匹配价格（BigInt as string）
	Size       string // 匹配数量（BigInt as string）
	TokenID    string
	Timestamp  int64
}

// MatchOrder 匹配订单信息
type MatchOrder struct {
	ID            string
	Maker         string
	Signer        string
	Taker         string
	TokenID       string
	MakerAmount   string
	TakerAmount   string
	Side          string
	Price         string
	Size          string
	Remaining     string
	Expiration    int64
	Salt          string
	Nonce         string
	FeeRateBps    string
	Signature     string
	SignatureType int32
	Funder        string
	OrderType     string
	Owner         string
}

// SubmitMatchReply 提交匹配响应
type SubmitMatchReply struct {
	TaskID  string
	Success bool
	Message string
}

// SubmitMatch 提交订单匹配结果
func (s *relayerService) SubmitMatch(ctx context.Context, req *SubmitMatchRequest) (*SubmitMatchReply, error) {
	// 1. 构建交易数据（将订单信息编码到 data 字段）
	// 注意：这里简化处理，实际应该调用 CLOB 合约的 matchOrders 函数
	// 2. 创建 CLOB_ORDER 类型的交易
	// 3. 将订单 ID 信息存储到 Signature 字段（JSON 格式）以便后续查询

	// 将订单 ID 信息编码为 JSON 存储在 Signature 字段中
	orderIDs := map[string]string{
		"maker_order_id": req.MakerOrder.ID,
		"taker_order_id": req.TakerOrder.ID,
	}
	orderIDsJSON, err := json.Marshal(orderIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal order IDs: %w", err)
	}

	// 构建交易请求
	// 注意：这里简化处理，实际应该：
	// 1. 调用 CLOB 合约的 matchOrders 函数，编码交易数据
	// 2. 使用正确的 to 地址（CLOB 合约地址）
	// 3. 设置正确的 gas limit

	// 简化实现：创建一个 CLOB_ORDER 类型的交易
	// 实际应该从配置中获取 CLOB 合约地址
	clobContractAddress := "" // TODO: 从配置获取 CLOB 合约地址

	// 构建交易数据（这里简化，实际应该编码 matchOrders 函数调用）
	txData := "" // TODO: 编码 matchOrders(makerOrder, takerOrder) 函数调用

	// 创建交易请求
	txReq := &SubmitTransactionRequest{
		To:              clobContractAddress,
		Data:            txData,
		Signature:       string(orderIDsJSON), // 将订单 ID 存储在 Signature 字段中
		Forwarder:       "",
		GasLimit:        500000, // 默认 Gas Limit
		TransactionType: "CLOB_ORDER",
		Value:           "0x0",
		AuthRequest:     nil, // CLOB 订单不需要 Builder 认证
	}

	// 提交交易
	reply, err := s.SubmitTransaction(ctx, txReq)
	if err != nil {
		return nil, fmt.Errorf("failed to submit match transaction: %w", err)
	}

	return &SubmitMatchReply{
		TaskID:  reply.TaskID,
		Success: reply.Success,
		Message: reply.Message,
	}, nil
}

// GetTransactionHashByOrderID 根据订单 ID 获取交易哈希
func (s *relayerService) GetTransactionHashByOrderID(ctx context.Context, orderID string) (string, error) {
	tx, err := s.txRepo.GetByOrderID(ctx, orderID)
	if err != nil {
		return "", fmt.Errorf("failed to get transaction by order ID: %w", err)
	}
	if tx == nil {
		return "", nil // 未找到交易，返回空字符串
	}
	return tx.TxHash, nil
}
