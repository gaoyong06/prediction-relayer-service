package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	v1 "xinyuan_tech/relayer-service/api/relayer/v1"
	"xinyuan_tech/relayer-service/internal/auth"
	"xinyuan_tech/relayer-service/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/grpc/metadata"
)

// RelayerService Relayer 服务实现
type RelayerService struct {
	v1.UnimplementedRelayerServer

	bizService  biz.RelayerService
	authService auth.AuthService
	logger      log.Logger
}

// NewRelayerService 创建 Relayer 服务
func NewRelayerService(
	bizService biz.RelayerService,
	authService auth.AuthService,
	logger log.Logger,
) *RelayerService {
	return &RelayerService{
		bizService:  bizService,
		authService: authService,
		logger:      logger,
	}
}

// extractAuthHeaders 从 gRPC metadata 中提取 Builder 认证头
func (s *RelayerService) extractAuthHeaders(ctx context.Context, method, path string, body []byte) (*auth.AuthRequest, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("no metadata found")
	}

	// 从 HTTP 头中提取（gRPC-Gateway 会将 HTTP 头转换为 metadata）
	apiKey := ""
	signature := ""
	timestamp := ""
	passphrase := ""

	if values := md.Get("poly-builder-api-key"); len(values) > 0 {
		apiKey = values[0]
	}
	if values := md.Get("poly-builder-signature"); len(values) > 0 {
		signature = values[0]
	}
	if values := md.Get("poly-builder-timestamp"); len(values) > 0 {
		timestamp = values[0]
	}
	if values := md.Get("poly-builder-passphrase"); len(values) > 0 {
		passphrase = values[0]
	}

	// 构建 body 字符串（用于签名验证）
	bodyStr := string(body)

	return &auth.AuthRequest{
		APIKey:     apiKey,
		Signature:  signature,
		Timestamp:  timestamp,
		Passphrase: passphrase,
		Method:     method,
		Path:       path,
		Body:       bodyStr,
	}, nil
}

// SubmitTransaction 提交单笔交易
func (s *RelayerService) SubmitTransaction(ctx context.Context, req *v1.SubmitTransactionRequest) (*v1.SubmitTransactionReply, error) {
	// 1. 提取 Builder 认证信息
	// 注意：对于 gRPC，需要从请求中构建 body
	body, _ := json.Marshal(req)
	authReq, err := s.extractAuthHeaders(ctx, "POST", "/v1/submit", body)
	if err != nil {
		return nil, fmt.Errorf("failed to extract auth headers: %w", err)
	}

	// 2. 构建业务请求
	bizReq := &biz.SubmitTransactionRequest{
		To:              req.To,
		Data:            req.Data,
		Signature:       req.Signature,
		Forwarder:       req.Forwarder,
		GasLimit:        req.GasLimit,
		TransactionType: req.TransactionType.String(),
		Value:           req.Value,
		AuthRequest:     authReq,
	}

	// 3. 调用业务服务
	reply, err := s.bizService.SubmitTransaction(ctx, bizReq)
	if err != nil {
		return nil, err
	}

	return &v1.SubmitTransactionReply{
		TaskId:  reply.TaskID,
		Success: reply.Success,
		Message: reply.Message,
	}, nil
}

// SubmitBatchTransaction 提交批量交易
func (s *RelayerService) SubmitBatchTransaction(ctx context.Context, req *v1.SubmitBatchTransactionRequest) (*v1.SubmitBatchTransactionReply, error) {
	// 1. 提取 Builder 认证信息（使用第一个交易的认证信息）
	if len(req.Transactions) == 0 {
		return nil, fmt.Errorf("no transactions provided")
	}

	body, _ := json.Marshal(req)
	authReq, err := s.extractAuthHeaders(ctx, "POST", "/v1/submit/batch", body)
	if err != nil {
		return nil, fmt.Errorf("failed to extract auth headers: %w", err)
	}

	// 2. 构建业务请求
	bizTransactions := make([]*biz.SubmitTransactionRequest, 0, len(req.Transactions))
	for _, tx := range req.Transactions {
		bizTransactions = append(bizTransactions, &biz.SubmitTransactionRequest{
			To:              tx.To,
			Data:            tx.Data,
			Signature:       tx.Signature,
			Forwarder:       tx.Forwarder,
			GasLimit:        tx.GasLimit,
			TransactionType: tx.TransactionType.String(),
			Value:           tx.Value,
			AuthRequest:     authReq,
		})
	}

	bizReq := &biz.SubmitBatchTransactionRequest{
		Transactions:  bizTransactions,
		BuilderAPIKey: req.BuilderApiKey,
	}

	// 3. 调用业务服务
	reply, err := s.bizService.SubmitBatchTransaction(ctx, bizReq)
	if err != nil {
		return nil, err
	}

	return &v1.SubmitBatchTransactionReply{
		TaskIds: reply.TaskIDs,
		Success: reply.Success,
		Message: reply.Message,
	}, nil
}

// DeployWallet 部署钱包
func (s *RelayerService) DeployWallet(ctx context.Context, req *v1.DeployWalletRequest) (*v1.DeployWalletReply, error) {
	// TODO: 实现钱包部署
	// 1. 提取 Builder 认证信息
	// 2. 调用钱包部署器
	// 3. 返回部署的钱包地址

	return &v1.DeployWalletReply{
		Success: false,
		Message: "Wallet deployment not implemented yet",
	}, nil
}

// GetTransactionStatus 获取交易状态
func (s *RelayerService) GetTransactionStatus(ctx context.Context, req *v1.GetTransactionStatusRequest) (*v1.GetTransactionStatusReply, error) {
	// 调用业务服务
	status, err := s.bizService.GetTransactionStatus(ctx, req.TaskId)
	if err != nil {
		return nil, err
	}

	return &v1.GetTransactionStatusReply{
		Status: &v1.TransactionStatus{
			TaskId:      status.TaskID,
			TxHash:      status.TxHash,
			Status:      status.Status,
			GasPrice:    status.GasPrice,
			BlockNumber: status.BlockNumber,
			GasUsed:     status.GasUsed,
			CreatedAt:   status.CreatedAt,
			UpdatedAt:   status.UpdatedAt,
		},
	}, nil
}

// GetBuilderFeeStats 获取 Builder 费用统计
func (s *RelayerService) GetBuilderFeeStats(ctx context.Context, req *v1.GetBuilderFeeStatsRequest) (*v1.GetBuilderFeeStatsReply, error) {
	// 1. 提取 Builder 认证信息
	body, _ := json.Marshal(req)
	authReq, err := s.extractAuthHeaders(ctx, "GET", "/v1/builder/fees", body)
	if err != nil {
		return nil, fmt.Errorf("failed to extract auth headers: %w", err)
	}

	// 2. 验证认证（确保只能查询自己的费用）
	builder, err := s.authService.ValidateBuilderAuth(ctx, authReq)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// 3. 解析时间范围
	startTime := time.Unix(req.StartTime, 0)
	endTime := time.Unix(req.EndTime, 0)

	// 4. 调用业务服务
	stats, err := s.bizService.GetBuilderFeeStats(ctx, builder.APIKey, startTime, endTime)
	if err != nil {
		return nil, err
	}

	// 5. 转换为响应格式
	byType := make(map[string]*v1.FeeStatsByType)
	for k, v := range stats.ByType {
		byType[k] = &v1.FeeStatsByType{
			Count:   v.Count,
			GasUsed: v.GasUsed,
			Cost:    v.Cost,
		}
	}

	return &v1.GetBuilderFeeStatsReply{
		TotalTransactions: stats.TotalTransactions,
		TotalGasUsed:      stats.TotalGasUsed,
		TotalCost:         stats.TotalCost,
		ByType:            byType,
	}, nil
}

// GetOperatorBalance 获取 Operator 余额
func (s *RelayerService) GetOperatorBalance(ctx context.Context, req *v1.GetOperatorBalanceRequest) (*v1.GetOperatorBalanceReply, error) {
	// TODO: 实现 Operator 余额查询
	// 1. 从以太坊客户端查询余额
	// 2. 转换为 MATIC 单位

	return &v1.GetOperatorBalanceReply{
		OperatorAddress: req.OperatorAddress,
		Balance:         "0",
		BalanceMatic:    "0",
	}, nil
}


