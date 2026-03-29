-- 创建外部系统映射表
-- 用于存储外部系统（钉钉、飞书等）ID 与内部 ID 的映射关系

CREATE TABLE IF NOT EXISTS external_mappings (
    id VARCHAR(36) PRIMARY KEY,
    source_type VARCHAR(50) NOT NULL,
    external_id VARCHAR(255) NOT NULL,
    mapping_type VARCHAR(20) NOT NULL,
    internal_id VARCHAR(36) NOT NULL,
    name VARCHAR(200),
    extra_data TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    -- 索引
    INDEX idx_source (source_type),
    INDEX idx_external (external_id),
    INDEX idx_internal (internal_id),
    UNIQUE INDEX idx_source_external_mapping (source_type, external_id, mapping_type)
);

-- 添加注释
COMMENT ON TABLE external_mappings IS '外部系统映射表，用于存储外部系统ID与内部ID的映射关系';
