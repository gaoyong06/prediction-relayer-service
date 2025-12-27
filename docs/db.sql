-- Relayer Service 数据库建表 SQL

-- 1. Transactions 表（交易记录）
CREATE TABLE IF NOT EXISTS transactions (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    task_id VARCHAR(36) UNIQUE NOT NULL COMMENT '任务 ID（UUID）',
    tx_hash VARCHAR(66) UNIQUE COMMENT '交易哈希',
    builder_api_key VARCHAR(255) NOT NULL COMMENT 'Builder API Key',
    from_address VARCHAR(42) NOT NULL COMMENT 'Operator 地址',
    to_address VARCHAR(42) NOT NULL COMMENT '目标合约地址',
    target_contract VARCHAR(42) NOT NULL COMMENT '目标合约地址（与 to_address 相同）',
    transaction_type VARCHAR(50) NOT NULL COMMENT '交易类型：WALLET_DEPLOYMENT, TOKEN_APPROVAL, CTF_SPLIT, CTF_MERGE, CTF_REDEEM, CLOB_ORDER, CUSTOM',
    data TEXT NOT NULL COMMENT '交易数据（hex）',
    value VARCHAR(78) NOT NULL DEFAULT '0x0' COMMENT '交易金额（hex）',
    signature TEXT COMMENT '用户签名',
    forwarder VARCHAR(42) COMMENT '转发器合约地址',
    nonce BIGINT NOT NULL COMMENT 'Nonce',
    gas_limit BIGINT NOT NULL COMMENT 'Gas Limit',
    gas_price VARCHAR(78) NOT NULL COMMENT 'Gas Price（字符串，支持大整数）',
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING' COMMENT '状态：PENDING, MINED, FAILED, REPLACED',
    block_number BIGINT COMMENT '区块号',
    gas_used BIGINT COMMENT '实际使用的 Gas',
    error_message TEXT COMMENT '错误信息',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_tx_hash (tx_hash),
    INDEX idx_task_id (task_id),
    INDEX idx_builder_api_key (builder_api_key),
    INDEX idx_from_address (from_address),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at),
    INDEX idx_status_created_at (status, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='交易记录表';

-- 2. Builders 表（Builder 认证信息）
CREATE TABLE IF NOT EXISTS builders (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    api_key VARCHAR(255) UNIQUE NOT NULL COMMENT 'Builder API Key',
    secret_hash VARCHAR(255) NOT NULL COMMENT 'Secret 哈希（加密存储）',
    passphrase_hash VARCHAR(255) NOT NULL COMMENT 'Passphrase 哈希（加密存储）',
    name VARCHAR(255) COMMENT 'Builder 名称',
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE' COMMENT '状态：ACTIVE, SUSPENDED, REVOKED',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_api_key (api_key),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Builder 认证信息表';

-- 3. Builder Fees 表（Builder 费用统计）
CREATE TABLE IF NOT EXISTS builder_fees (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    builder_api_key VARCHAR(255) NOT NULL COMMENT 'Builder API Key',
    transaction_type VARCHAR(50) NOT NULL COMMENT '交易类型',
    transaction_id VARCHAR(36) NOT NULL COMMENT '交易 ID（关联 transactions.task_id）',
    gas_used BIGINT NOT NULL COMMENT 'Gas 消耗',
    gas_price VARCHAR(78) NOT NULL COMMENT 'Gas 价格（字符串）',
    total_cost VARCHAR(78) NOT NULL COMMENT '总成本（MATIC，字符串）',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX idx_builder_api_key (builder_api_key),
    INDEX idx_transaction_type (transaction_type),
    INDEX idx_created_at (created_at),
    INDEX idx_builder_type_created (builder_api_key, transaction_type, created_at),
    FOREIGN KEY (transaction_id) REFERENCES transactions(task_id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Builder 费用统计表';

-- 4. Operators 表（Operator 钱包管理）
CREATE TABLE IF NOT EXISTS operators (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    address VARCHAR(42) UNIQUE NOT NULL COMMENT 'Operator 地址',
    private_key_encrypted TEXT NOT NULL COMMENT '私钥（加密存储）',
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE' COMMENT '状态：ACTIVE, INACTIVE',
    balance_threshold VARCHAR(78) NOT NULL DEFAULT '1000000000000000000' COMMENT '余额阈值（wei，字符串）',
    current_nonce BIGINT NOT NULL DEFAULT 0 COMMENT '当前 Nonce',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_address (address),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Operator 钱包管理表';


-- 1. Transactions 表（交易记录）
CREATE TABLE IF NOT EXISTS transactions (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    task_id VARCHAR(36) UNIQUE NOT NULL COMMENT '任务 ID（UUID）',
    tx_hash VARCHAR(66) UNIQUE COMMENT '交易哈希',
    builder_api_key VARCHAR(255) NOT NULL COMMENT 'Builder API Key',
    from_address VARCHAR(42) NOT NULL COMMENT 'Operator 地址',
    to_address VARCHAR(42) NOT NULL COMMENT '目标合约地址',
    target_contract VARCHAR(42) NOT NULL COMMENT '目标合约地址（与 to_address 相同）',
    transaction_type VARCHAR(50) NOT NULL COMMENT '交易类型：WALLET_DEPLOYMENT, TOKEN_APPROVAL, CTF_SPLIT, CTF_MERGE, CTF_REDEEM, CLOB_ORDER, CUSTOM',
    data TEXT NOT NULL COMMENT '交易数据（hex）',
    value VARCHAR(78) NOT NULL DEFAULT '0x0' COMMENT '交易金额（hex）',
    signature TEXT COMMENT '用户签名',
    forwarder VARCHAR(42) COMMENT '转发器合约地址',
    nonce BIGINT NOT NULL COMMENT 'Nonce',
    gas_limit BIGINT NOT NULL COMMENT 'Gas Limit',
    gas_price VARCHAR(78) NOT NULL COMMENT 'Gas Price（字符串，支持大整数）',
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING' COMMENT '状态：PENDING, MINED, FAILED, REPLACED',
    block_number BIGINT COMMENT '区块号',
    gas_used BIGINT COMMENT '实际使用的 Gas',
    error_message TEXT COMMENT '错误信息',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_tx_hash (tx_hash),
    INDEX idx_task_id (task_id),
    INDEX idx_builder_api_key (builder_api_key),
    INDEX idx_from_address (from_address),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at),
    INDEX idx_status_created_at (status, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='交易记录表';

-- 2. Builders 表（Builder 认证信息）
CREATE TABLE IF NOT EXISTS builders (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    api_key VARCHAR(255) UNIQUE NOT NULL COMMENT 'Builder API Key',
    secret_hash VARCHAR(255) NOT NULL COMMENT 'Secret 哈希（加密存储）',
    passphrase_hash VARCHAR(255) NOT NULL COMMENT 'Passphrase 哈希（加密存储）',
    name VARCHAR(255) COMMENT 'Builder 名称',
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE' COMMENT '状态：ACTIVE, SUSPENDED, REVOKED',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_api_key (api_key),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Builder 认证信息表';

-- 3. Builder Fees 表（Builder 费用统计）
CREATE TABLE IF NOT EXISTS builder_fees (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    builder_api_key VARCHAR(255) NOT NULL COMMENT 'Builder API Key',
    transaction_type VARCHAR(50) NOT NULL COMMENT '交易类型',
    transaction_id VARCHAR(36) NOT NULL COMMENT '交易 ID（关联 transactions.task_id）',
    gas_used BIGINT NOT NULL COMMENT 'Gas 消耗',
    gas_price VARCHAR(78) NOT NULL COMMENT 'Gas 价格（字符串）',
    total_cost VARCHAR(78) NOT NULL COMMENT '总成本（MATIC，字符串）',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX idx_builder_api_key (builder_api_key),
    INDEX idx_transaction_type (transaction_type),
    INDEX idx_created_at (created_at),
    INDEX idx_builder_type_created (builder_api_key, transaction_type, created_at),
    FOREIGN KEY (transaction_id) REFERENCES transactions(task_id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Builder 费用统计表';

-- 4. Operators 表（Operator 钱包管理）
CREATE TABLE IF NOT EXISTS operators (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    address VARCHAR(42) UNIQUE NOT NULL COMMENT 'Operator 地址',
    private_key_encrypted TEXT NOT NULL COMMENT '私钥（加密存储）',
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE' COMMENT '状态：ACTIVE, INACTIVE',
    balance_threshold VARCHAR(78) NOT NULL DEFAULT '1000000000000000000' COMMENT '余额阈值（wei，字符串）',
    current_nonce BIGINT NOT NULL DEFAULT 0 COMMENT '当前 Nonce',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_address (address),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Operator 钱包管理表';


-- 1. Transactions 表（交易记录）
CREATE TABLE IF NOT EXISTS transactions (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    task_id VARCHAR(36) UNIQUE NOT NULL COMMENT '任务 ID（UUID）',
    tx_hash VARCHAR(66) UNIQUE COMMENT '交易哈希',
    builder_api_key VARCHAR(255) NOT NULL COMMENT 'Builder API Key',
    from_address VARCHAR(42) NOT NULL COMMENT 'Operator 地址',
    to_address VARCHAR(42) NOT NULL COMMENT '目标合约地址',
    target_contract VARCHAR(42) NOT NULL COMMENT '目标合约地址（与 to_address 相同）',
    transaction_type VARCHAR(50) NOT NULL COMMENT '交易类型：WALLET_DEPLOYMENT, TOKEN_APPROVAL, CTF_SPLIT, CTF_MERGE, CTF_REDEEM, CLOB_ORDER, CUSTOM',
    data TEXT NOT NULL COMMENT '交易数据（hex）',
    value VARCHAR(78) NOT NULL DEFAULT '0x0' COMMENT '交易金额（hex）',
    signature TEXT COMMENT '用户签名',
    forwarder VARCHAR(42) COMMENT '转发器合约地址',
    nonce BIGINT NOT NULL COMMENT 'Nonce',
    gas_limit BIGINT NOT NULL COMMENT 'Gas Limit',
    gas_price VARCHAR(78) NOT NULL COMMENT 'Gas Price（字符串，支持大整数）',
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING' COMMENT '状态：PENDING, MINED, FAILED, REPLACED',
    block_number BIGINT COMMENT '区块号',
    gas_used BIGINT COMMENT '实际使用的 Gas',
    error_message TEXT COMMENT '错误信息',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_tx_hash (tx_hash),
    INDEX idx_task_id (task_id),
    INDEX idx_builder_api_key (builder_api_key),
    INDEX idx_from_address (from_address),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at),
    INDEX idx_status_created_at (status, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='交易记录表';

-- 2. Builders 表（Builder 认证信息）
CREATE TABLE IF NOT EXISTS builders (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    api_key VARCHAR(255) UNIQUE NOT NULL COMMENT 'Builder API Key',
    secret_hash VARCHAR(255) NOT NULL COMMENT 'Secret 哈希（加密存储）',
    passphrase_hash VARCHAR(255) NOT NULL COMMENT 'Passphrase 哈希（加密存储）',
    name VARCHAR(255) COMMENT 'Builder 名称',
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE' COMMENT '状态：ACTIVE, SUSPENDED, REVOKED',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_api_key (api_key),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Builder 认证信息表';

-- 3. Builder Fees 表（Builder 费用统计）
CREATE TABLE IF NOT EXISTS builder_fees (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    builder_api_key VARCHAR(255) NOT NULL COMMENT 'Builder API Key',
    transaction_type VARCHAR(50) NOT NULL COMMENT '交易类型',
    transaction_id VARCHAR(36) NOT NULL COMMENT '交易 ID（关联 transactions.task_id）',
    gas_used BIGINT NOT NULL COMMENT 'Gas 消耗',
    gas_price VARCHAR(78) NOT NULL COMMENT 'Gas 价格（字符串）',
    total_cost VARCHAR(78) NOT NULL COMMENT '总成本（MATIC，字符串）',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX idx_builder_api_key (builder_api_key),
    INDEX idx_transaction_type (transaction_type),
    INDEX idx_created_at (created_at),
    INDEX idx_builder_type_created (builder_api_key, transaction_type, created_at),
    FOREIGN KEY (transaction_id) REFERENCES transactions(task_id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Builder 费用统计表';

-- 4. Operators 表（Operator 钱包管理）
CREATE TABLE IF NOT EXISTS operators (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    address VARCHAR(42) UNIQUE NOT NULL COMMENT 'Operator 地址',
    private_key_encrypted TEXT NOT NULL COMMENT '私钥（加密存储）',
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE' COMMENT '状态：ACTIVE, INACTIVE',
    balance_threshold VARCHAR(78) NOT NULL DEFAULT '1000000000000000000' COMMENT '余额阈值（wei，字符串）',
    current_nonce BIGINT NOT NULL DEFAULT 0 COMMENT '当前 Nonce',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_address (address),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Operator 钱包管理表';



