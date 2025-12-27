/*
 Navicat Premium Dump SQL

 Source Server         : localhost
 Source Server Type    : MySQL
 Source Server Version : 90500 (9.5.0)
 Source Host           : localhost:3306
 Source Schema         : prediction_relayer_service

 Target Server Type    : MySQL
 Target Server Version : 90500 (9.5.0)
 File Encoding         : 65001

 Date: 27/12/2025 16:17:27
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for builder
-- Builder 认证信息表：存储 Builder 的 API Key、认证信息和状态
-- ----------------------------
DROP TABLE IF EXISTS `builder`;
CREATE TABLE `builder` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
  `api_key` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'API Key（唯一标识）',
  `secret_hash` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'Secret 哈希值（加密存储）',
  `passphrase_hash` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'Passphrase 哈希值（加密存储）',
  `name` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'Builder 名称（可选）',
  `status` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'ACTIVE' COMMENT '状态：ACTIVE（激活）, INACTIVE（未激活）',
  `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_builder_api_key` (`api_key`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Builder 认证信息表';

-- ----------------------------
-- Table structure for builder_fee
-- Builder 费用统计表：记录每个 Builder 的交易费用，用于费用统计和结算
-- ----------------------------
DROP TABLE IF EXISTS `builder_fee`;
CREATE TABLE `builder_fee` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
  `builder_api_key` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'Builder API Key',
  `transaction_type` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '交易类型（用于按类型统计）',
  `transaction_id` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '关联 transaction.task_id',
  `gas_used` bigint NOT NULL COMMENT 'Gas 消耗量',
  `gas_price` varchar(78) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'Gas 价格（字符串，支持大整数）',
  `total_cost` varchar(78) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '总成本（MATIC，字符串格式）',
  `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_builder_api_key` (`builder_api_key`),
  KEY `idx_transaction_type` (`transaction_type`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Builder 费用统计表';

-- ----------------------------
-- Table structure for operator
-- Operator 钱包管理表：存储 Operator 钱包地址、加密私钥和状态信息
-- ----------------------------
DROP TABLE IF EXISTS `operator`;
CREATE TABLE `operator` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
  `address` varchar(42) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'Operator 钱包地址',
  `private_key_encrypted` text COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '私钥（加密存储）',
  `status` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'ACTIVE' COMMENT '状态：ACTIVE（激活）, INACTIVE（未激活）',
  `balance_threshold` varchar(78) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '1000000000000000000' COMMENT '余额告警阈值（wei，默认 1 MATIC）',
  `current_nonce` bigint NOT NULL DEFAULT '0' COMMENT '当前 nonce（用于交易排序）',
  `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_operator_address` (`address`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Operator 钱包管理表';

-- ----------------------------
-- Table structure for transaction
-- 交易记录表：存储所有通过 Relayer Service 提交的交易记录，包括交易状态、Gas 信息等
-- ----------------------------
DROP TABLE IF EXISTS `transaction`;
CREATE TABLE `transaction` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
  `task_id` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '任务 ID（UUID）',
  `tx_hash` varchar(66) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '交易哈希（0x 开头的 66 字符）',
  `builder_api_key` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'Builder API Key（用于费用追踪）',
  `from_address` varchar(42) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '发送方地址（Operator 地址）',
  `to_address` varchar(42) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '接收方地址（目标合约或转发器）',
  `target_contract` varchar(42) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '目标合约地址',
  `transaction_type` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '交易类型：WALLET_DEPLOYMENT（钱包部署）, TOKEN_APPROVAL（代币授权）, CTF_SPLIT（CTF 拆分）, CTF_MERGE（CTF 合并）, CTF_REDEEM（CTF 赎回）, CLOB_ORDER（CLOB 订单执行）, CUSTOM（自定义交易）',
  `data` text COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '交易数据（hex 编码的函数调用数据）',
  `value` varchar(78) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '0x0' COMMENT '交易金额（hex 格式，通常为 "0x0"）',
  `signature` text COLLATE utf8mb4_unicode_ci COMMENT '用户签名（消息签名，不是交易签名）',
  `forwarder` varchar(42) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '转发器合约地址（可选）',
  `nonce` bigint NOT NULL COMMENT '交易 nonce',
  `gas_limit` bigint NOT NULL COMMENT 'Gas 限制',
  `gas_price` varchar(78) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'Gas 价格（字符串，支持大整数）',
  `status` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'PENDING' COMMENT '交易状态：PENDING（待处理）, MINED（已打包）, FAILED（失败）, REPLACED（被替换）',
  `block_number` bigint DEFAULT NULL COMMENT '区块号（交易被打包后才有值）',
  `gas_used` bigint DEFAULT NULL COMMENT '实际使用的 Gas（交易被打包后才有值）',
  `error_message` text COLLATE utf8mb4_unicode_ci COMMENT '错误信息（交易失败时记录）',
  `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_transaction_task_id` (`task_id`),
  UNIQUE KEY `idx_transaction_tx_hash` (`tx_hash`),
  KEY `idx_builder_api_key` (`builder_api_key`),
  KEY `idx_from_address` (`from_address`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_status_created_at` (`updated_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='交易记录表';

SET FOREIGN_KEY_CHECKS = 1;
