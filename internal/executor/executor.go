package executor

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"prediction-relayer-service/internal/data"
	"prediction-relayer-service/internal/nonce"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Executor 交易执行器接口
type Executor interface {
	// Execute 执行交易
	Execute(ctx context.Context, tx *data.Transaction, operator *data.Operator) (*ExecutionResult, error)

	// EstimateGas 估算 Gas Limit
	EstimateGas(ctx context.Context, tx *data.Transaction) (uint64, error)

	// SelectOperator 选择可用的 Operator
	SelectOperator(ctx context.Context) (*data.Operator, error)
}

// ExecutionResult 执行结果
type ExecutionResult struct {
	TxHash   string
	GasUsed  uint64
	BlockNum uint64
}

// executor 交易执行器实现
type executor struct {
	ethClient     *ethclient.Client
	chainID       *big.Int
	nonceMgr      nonce.Manager
	operatorRepo  data.OperatorRepo
	gasMultiplier int64 // Gas Price 倍数（例如 110 = 110%）
}

// NewExecutor 创建交易执行器
func NewExecutor(
	ethClient *ethclient.Client,
	chainID *big.Int,
	nonceMgr nonce.Manager,
	operatorRepo data.OperatorRepo,
	gasMultiplier int64,
) Executor {
	return &executor{
		ethClient:     ethClient,
		chainID:       chainID,
		nonceMgr:      nonceMgr,
		operatorRepo:  operatorRepo,
		gasMultiplier: gasMultiplier,
	}
}

// Execute 执行交易
func (e *executor) Execute(ctx context.Context, tx *data.Transaction, operator *data.Operator) (*ExecutionResult, error) {
	// 1. 获取 Nonce
	nonce, err := e.nonceMgr.AcquireNonce(ctx, operator.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire nonce: %w", err)
	}

	// 2. 估算 Gas Limit（如果未提供）
	gasLimit := uint64(tx.GasLimit)
	if gasLimit == 0 {
		gasLimit, err = e.EstimateGas(ctx, tx)
		if err != nil {
			e.nonceMgr.ReleaseNonce(ctx, operator.Address, nonce)
			return nil, fmt.Errorf("failed to estimate gas: %w", err)
		}
	}

	// 3. 获取 Gas Price（加权 10% 以加速）
	gasPrice, err := e.ethClient.SuggestGasPrice(ctx)
	if err != nil {
		e.nonceMgr.ReleaseNonce(ctx, operator.Address, nonce)
		return nil, fmt.Errorf("failed to get gas price: %w", err)
	}
	// 应用倍数
	gasPrice = new(big.Int).Mul(gasPrice, big.NewInt(e.gasMultiplier))
	gasPrice = new(big.Int).Div(gasPrice, big.NewInt(100))

	// 4. 解析目标地址和数据
	toAddr := common.HexToAddress(tx.ToAddress)
	var dataBytes []byte
	if tx.Data != "" {
		dataBytes = common.FromHex(tx.Data)
	}

	// 5. 解析 Value
	value := big.NewInt(0)
	if tx.Value != "" && tx.Value != "0x0" {
		value, _ = new(big.Int).SetString(tx.Value[2:], 16)
	}

	// 6. 解密私钥
	// 注意：这里需要从 KMS 解密私钥
	// TODO: 集成 KMS 解密私钥
	// 临时实现：假设 private_key_encrypted 就是私钥（实际应该解密）
	privateKey, err := crypto.HexToECDSA(operator.PrivateKeyEncrypted)
	if err != nil {
		e.nonceMgr.ReleaseNonce(ctx, operator.Address, nonce)
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	// 7. 创建交易
	rawTx := types.NewTransaction(
		nonce,
		toAddr,
		value,
		gasLimit,
		gasPrice,
		dataBytes,
	)

	// 8. 签名交易
	signer := types.NewEIP155Signer(e.chainID)
	signedTx, err := types.SignTx(rawTx, signer, privateKey)
	if err != nil {
		e.nonceMgr.ReleaseNonce(ctx, operator.Address, nonce)
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	// 9. 广播交易
	if err := e.ethClient.SendTransaction(ctx, signedTx); err != nil {
		e.nonceMgr.ReleaseNonce(ctx, operator.Address, nonce)
		return nil, fmt.Errorf("failed to send transaction: %w", err)
	}

	// 10. 更新交易记录中的 Gas Price（字符串格式）
	// 注意：这里不等待交易确认，交易监控器会处理确认逻辑

	return &ExecutionResult{
		TxHash: signedTx.Hash().Hex(),
	}, nil
}

// EstimateGas 估算 Gas Limit
func (e *executor) EstimateGas(ctx context.Context, tx *data.Transaction) (uint64, error) {
	toAddr := common.HexToAddress(tx.ToAddress)
	var dataBytes []byte
	if tx.Data != "" {
		dataBytes = common.FromHex(tx.Data)
	}

	value := big.NewInt(0)
	if tx.Value != "" && tx.Value != "0x0" {
		value, _ = new(big.Int).SetString(tx.Value[2:], 16)
	}

	// 估算 Gas（使用第一个 Operator 的地址作为 from）
	operators, err := e.operatorRepo.GetActiveOperators(ctx)
	if err != nil || len(operators) == 0 {
		return 0, fmt.Errorf("no active operators available")
	}
	fromAddr := common.HexToAddress(operators[0].Address)

	gasLimit, err := e.ethClient.EstimateGas(ctx, ethereum.CallMsg{
		From:  fromAddr,
		To:    &toAddr,
		Value: value,
		Data:  dataBytes,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to estimate gas: %w", err)
	}

	// 增加 20% 的安全余量
	gasLimit = gasLimit * 120 / 100

	return gasLimit, nil
}

// SelectOperator 选择可用的 Operator
func (e *executor) SelectOperator(ctx context.Context) (*data.Operator, error) {
	operators, err := e.operatorRepo.GetActiveOperators(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get operators: %w", err)
	}
	if len(operators) == 0 {
		return nil, fmt.Errorf("no active operators available")
	}

	// 简单的轮询选择（TODO: 可以实现更智能的选择策略，如基于余额、负载等）
	// 这里使用时间戳取模实现轮询
	index := time.Now().Unix() % int64(len(operators))
	return operators[index], nil
}
