package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"xinyuan_tech/relayer-service/internal/data"
)

// AuthService Builder 认证服务接口
type AuthService interface {
	// ValidateBuilderAuth 验证 Builder 认证
	// 验证 HMAC 签名、时间戳、API Key 有效性
	ValidateBuilderAuth(ctx context.Context, req *AuthRequest) (*data.Builder, error)

	// BuildHMACSignature 构建 HMAC 签名（用于测试）
	BuildHMACSignature(secret string, timestamp int64, method, path, body string) string
}

// AuthRequest 认证请求
type AuthRequest struct {
	APIKey     string // POLY_BUILDER_API_KEY
	Signature  string // POLY_BUILDER_SIGNATURE
	Timestamp  string // POLY_BUILDER_TIMESTAMP
	Passphrase string // POLY_BUILDER_PASSPHRASE
	Method     string // HTTP 方法
	Path       string // HTTP 路径
	Body       string // 请求体
}

// authService Builder 认证服务实现
type authService struct {
	builderRepo     data.BuilderRepo
	timestampWindow int64 // 时间戳验证窗口（毫秒）
}

// NewAuthService 创建认证服务
func NewAuthService(builderRepo data.BuilderRepo, timestampWindow int64) AuthService {
	return &authService{
		builderRepo:     builderRepo,
		timestampWindow: timestampWindow,
	}
}

// ValidateBuilderAuth 验证 Builder 认证
func (s *authService) ValidateBuilderAuth(ctx context.Context, req *AuthRequest) (*data.Builder, error) {
	// 1. 验证必填字段
	if req.APIKey == "" {
		return nil, fmt.Errorf("api key is required")
	}
	if req.Signature == "" {
		return nil, fmt.Errorf("signature is required")
	}
	if req.Timestamp == "" {
		return nil, fmt.Errorf("timestamp is required")
	}
	if req.Passphrase == "" {
		return nil, fmt.Errorf("passphrase is required")
	}

	// 2. 验证时间戳（防止重放攻击）
	timestamp, err := strconv.ParseInt(req.Timestamp, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid timestamp format: %w", err)
	}

	now := time.Now().UnixMilli()
	diff := now - timestamp
	if diff < 0 {
		diff = -diff
	}
	if diff > s.timestampWindow {
		return nil, fmt.Errorf("timestamp out of window: diff=%d ms, window=%d ms", diff, s.timestampWindow)
	}

	// 3. 查询 Builder
	builder, err := s.builderRepo.GetByAPIKey(ctx, req.APIKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get builder: %w", err)
	}
	if builder == nil {
		return nil, fmt.Errorf("builder not found")
	}

	// 4. 验证 Builder 状态
	if builder.Status != "ACTIVE" {
		return nil, fmt.Errorf("builder status is not active: %s", builder.Status)
	}

	// 5. 验证 Passphrase（需要解密后比较）
	// TODO: 实现 Passphrase 解密和验证
	// 这里简化处理，实际应该使用加密库解密 passphrase_hash 后比较
	if req.Passphrase != builder.PassphraseHash {
		// 注意：实际应该解密后比较，这里仅作示例
		return nil, fmt.Errorf("invalid passphrase")
	}

	// 6. 验证 HMAC 签名
	// 需要解密 secret_hash 获取原始 secret
	// TODO: 实现 Secret 解密
	// 这里简化处理，假设 secret_hash 就是 secret（实际应该解密）
	secret := builder.SecretHash // 实际应该解密

	expectedSignature := s.BuildHMACSignature(secret, timestamp, req.Method, req.Path, req.Body)
	if !hmac.Equal([]byte(expectedSignature), []byte(req.Signature)) {
		return nil, fmt.Errorf("invalid signature")
	}

	return builder, nil
}

// BuildHMACSignature 构建 HMAC 签名
// 参考：https://docs.polymarket.com/developers/builders/relayer-client
// signature = HMAC-SHA256(secret, timestamp + method + path + body)
func (s *authService) BuildHMACSignature(secret string, timestamp int64, method, path, body string) string {
	// 构建签名内容：timestamp + method + path + body
	signatureContent := strconv.FormatInt(timestamp, 10) + method + path + body

	// 计算 HMAC-SHA256
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signatureContent))
	signature := mac.Sum(nil)

	// 返回 hex 编码的签名
	return hex.EncodeToString(signature)
}

