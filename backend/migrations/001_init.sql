-- 初始数据库迁移脚本
-- 创建所有表结构和索引

-- ============================================
-- 员工表 (employees)
-- ============================================
CREATE TABLE IF NOT EXISTS employees (
    id TEXT PRIMARY KEY,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    deleted_at DATETIME,
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'member',
    status TEXT NOT NULL DEFAULT 'active',
    last_seen_at DATETIME,
    api_key TEXT
);

-- 员工表索引
CREATE INDEX IF NOT EXISTS idx_employees_email ON employees(email);
CREATE INDEX IF NOT EXISTS idx_employees_status ON employees(status);
CREATE INDEX IF NOT EXISTS idx_employees_deleted_at ON employees(deleted_at);
CREATE INDEX IF NOT EXISTS idx_employees_created_at ON employees(created_at);

-- ============================================
-- 频道表 (channels)
-- ============================================
CREATE TABLE IF NOT EXISTS channels (
    id TEXT PRIMARY KEY,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    deleted_at DATETIME,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    description TEXT,
    created_by TEXT NOT NULL
);

-- 频道表索引
CREATE INDEX IF NOT EXISTS idx_channels_created_by ON channels(created_by);
CREATE INDEX IF NOT EXISTS idx_channels_type ON channels(type);
CREATE INDEX IF NOT EXISTS idx_channels_deleted_at ON channels(deleted_at);

-- ============================================
-- 频道成员关联表 (channel_members)
-- ============================================
CREATE TABLE IF NOT EXISTS channel_members (
    id TEXT PRIMARY KEY,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    deleted_at DATETIME,
    channel_id TEXT NOT NULL,
    employee_id TEXT NOT NULL
);

-- 频道成员表索引（联合唯一索引）
CREATE UNIQUE INDEX IF NOT EXISTS idx_channel_member ON channel_members(channel_id, employee_id);
CREATE INDEX IF NOT EXISTS idx_channel_members_employee ON channel_members(employee_id);
CREATE INDEX IF NOT EXISTS idx_channel_members_deleted_at ON channel_members(deleted_at);

-- ============================================
-- 消息表 (messages)
-- ============================================
CREATE TABLE IF NOT EXISTS messages (
    id TEXT PRIMARY KEY,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    deleted_at DATETIME,
    channel_id TEXT NOT NULL,
    sender_id TEXT NOT NULL,
    type TEXT NOT NULL,
    content TEXT,
    mentions TEXT,        -- JSON 数组
    skills TEXT,          -- JSON 数组
    workflow_id TEXT,
    parent_id TEXT,
    is_deleted BOOLEAN DEFAULT FALSE
);

-- 消息表索引
CREATE INDEX IF NOT EXISTS idx_messages_channel_id ON messages(channel_id);
CREATE INDEX IF NOT EXISTS idx_messages_sender_id ON messages(sender_id);
CREATE INDEX IF NOT EXISTS idx_messages_workflow_id ON messages(workflow_id);
CREATE INDEX IF NOT EXISTS idx_messages_parent_id ON messages(parent_id);
CREATE INDEX IF NOT EXISTS idx_messages_type ON messages(type);
CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at);
CREATE INDEX IF NOT EXISTS idx_messages_deleted_at ON messages(deleted_at);

-- ============================================
-- 工作流表 (workflows)
-- ============================================
CREATE TABLE IF NOT EXISTS workflows (
    id TEXT PRIMARY KEY,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    deleted_at DATETIME,
    name TEXT NOT NULL,
    description TEXT,
    status TEXT NOT NULL DEFAULT 'active',
    created_by TEXT NOT NULL,
    trigger_type TEXT NOT NULL,
    trigger_config TEXT,  -- JSON 对象
    steps TEXT            -- JSON 数组
);

-- 工作流表索引
CREATE INDEX IF NOT EXISTS idx_workflows_created_by ON workflows(created_by);
CREATE INDEX IF NOT EXISTS idx_workflows_status ON workflows(status);
CREATE INDEX IF NOT EXISTS idx_workflows_trigger_type ON workflows(trigger_type);
CREATE INDEX IF NOT EXISTS idx_workflows_deleted_at ON workflows(deleted_at);

-- ============================================
-- 工作流执行记录表 (workflow_executions)
-- ============================================
CREATE TABLE IF NOT EXISTS workflow_executions (
    id TEXT PRIMARY KEY,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    deleted_at DATETIME,
    workflow_id TEXT NOT NULL,
    triggered_by TEXT NOT NULL,
    trigger_type TEXT,
    input TEXT,           -- JSON 对象
    status TEXT NOT NULL,
    output TEXT,          -- JSON 对象
    error_message TEXT,
    started_at INTEGER NOT NULL,
    completed_at INTEGER
);

-- 执行记录表索引
CREATE INDEX IF NOT EXISTS idx_executions_workflow_id ON workflow_executions(workflow_id);
CREATE INDEX IF NOT EXISTS idx_executions_triggered_by ON workflow_executions(triggered_by);
CREATE INDEX IF NOT EXISTS idx_executions_status ON workflow_executions(status);
CREATE INDEX IF NOT EXISTS idx_executions_started_at ON workflow_executions(started_at);
CREATE INDEX IF NOT EXISTS idx_executions_deleted_at ON workflow_executions(deleted_at);

-- ============================================
-- 任务表 (tasks)
-- ============================================
CREATE TABLE IF NOT EXISTS tasks (
    id TEXT PRIMARY KEY,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    deleted_at DATETIME,
    title TEXT NOT NULL,
    description TEXT,
    status TEXT NOT NULL DEFAULT 'pending',
    priority TEXT NOT NULL DEFAULT 'medium',
    source TEXT NOT NULL,
    channel_id TEXT,
    message_id TEXT,
    workflow_id TEXT,
    assignee_id TEXT,
    creator_id TEXT NOT NULL,
    due_date DATETIME,
    completed_at DATETIME
);

-- 任务表索引
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
CREATE INDEX IF NOT EXISTS idx_tasks_assignee_id ON tasks(assignee_id);
CREATE INDEX IF NOT EXISTS idx_tasks_creator_id ON tasks(creator_id);
CREATE INDEX IF NOT EXISTS idx_tasks_channel_id ON tasks(channel_id);
CREATE INDEX IF NOT EXISTS idx_tasks_message_id ON tasks(message_id);
CREATE INDEX IF NOT EXISTS idx_tasks_workflow_id ON tasks(workflow_id);
CREATE INDEX IF NOT EXISTS idx_tasks_due_date ON tasks(due_date);
CREATE INDEX IF NOT EXISTS idx_tasks_deleted_at ON tasks(deleted_at);

-- ============================================
-- 插入默认管理员账号
-- 密码: admin123 (bcrypt hash)
-- ============================================
INSERT OR IGNORE INTO employees (id, created_at, updated_at, name, email, password, role, status)
VALUES (
    'admin-001',
    datetime('now'),
    datetime('now'),
    '管理员',
    'admin@claw.local',
    'JDJhJDEwJGdIVGRXUDhoQmZLYjVGWE9PUVB2YmVjcy41ekxqcWVuQTU5OVpycjhidTJzT3VydDYzbVoy',
    'admin',
    'active'
);
