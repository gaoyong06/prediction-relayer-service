package kms

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

// KMS 密钥管理服务接口
type KMS interface {
	// Decrypt 解密数据
	Decrypt(ctx context.Context, encryptedData string) (string, error)

	// Encrypt 加密数据
	Encrypt(ctx context.Context, plainData string) (string, error)
}

// kmsType KMS 类型
type kmsType string

const (
	KMSTypeLocal kmsType = "local"   // 本地加密（使用配置的密钥）
	KMSTypeAWS   kmsType = "aws-kms" // AWS KMS
	KMSTypeVault kmsType = "vault"   // HashiCorp Vault
)

// localKMS 本地 KMS 实现（使用 AES-256-GCM）
type localKMS struct {
	key []byte // 32 字节密钥（AES-256）
}

// NewKMS 创建 KMS 实例
func NewKMS(kmsType, kmsConfig string) (KMS, error) {
	switch kmsType {
	case "local":
		// 从配置中读取密钥（base64 编码）
		key, err := base64.StdEncoding.DecodeString(kmsConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to decode key: %w", err)
		}
		if len(key) != 32 {
			return nil, fmt.Errorf("key must be 32 bytes for AES-256")
		}
		return &localKMS{key: key}, nil
	case "aws-kms":
		// TODO: 实现 AWS KMS 集成
		return nil, fmt.Errorf("AWS KMS not implemented yet")
	case "vault":
		// TODO: 实现 HashiCorp Vault 集成
		return nil, fmt.Errorf("HashiCorp Vault not implemented yet")
	default:
		return nil, fmt.Errorf("unknown KMS type: %s", kmsType)
	}
}

// Decrypt 解密数据（AES-256-GCM）
func (k *localKMS) Decrypt(ctx context.Context, encryptedData string) (string, error) {
	// 解码 base64
	data, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return "", fmt.Errorf("failed to decode encrypted data: %w", err)
	}

	if len(data) < 12+16 {
		return "", fmt.Errorf("encrypted data too short")
	}

	// 提取 nonce（前 12 字节）和密文
	nonce := data[:12]
	ciphertext := data[12:]

	// 创建 cipher
	block, err := aes.NewCipher(k.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// 解密
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

// Encrypt 加密数据（AES-256-GCM）
func (k *localKMS) Encrypt(ctx context.Context, plainData string) (string, error) {
	// 创建 cipher
	block, err := aes.NewCipher(k.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// 生成 nonce（12 字节）
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// 加密
	ciphertext := aesGCM.Seal(nil, nonce, []byte(plainData), nil)

	// 组合 nonce + ciphertext
	encrypted := append(nonce, ciphertext...)

	// 编码为 base64
	return base64.StdEncoding.EncodeToString(encrypted), nil
}


