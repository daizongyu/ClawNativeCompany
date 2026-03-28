// Package config 提供应用配置管理
// 支持 YAML 配置文件 + 环境变量覆盖
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

// Config 应用配置根结构
type Config struct {
	Server     ServerConfig     `yaml:"server"`
	Database   DatabaseConfig   `yaml:"database"`
	JWT        JWTConfig        `yaml:"jwt"`
	WebSocket  WebSocketConfig  `yaml:"websocket"`
	Log        LogConfig        `yaml:"log"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port int    `yaml:"port"`
	Mode string `yaml:"mode"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Path         string `yaml:"path"`
	MaxOpenConns int    `yaml:"max_open_conns"`
	MaxIdleConns int    `yaml:"max_idle_conns"`
}

// JWTConfig JWT 认证配置
type JWTConfig struct {
	Secret      string `yaml:"secret"`
	ExpireHours int    `yaml:"expire_hours"`
}

// WebSocketConfig WebSocket 配置
type WebSocketConfig struct {
	HeartbeatInterval int `yaml:"heartbeat_interval"`
	WriteTimeout      int `yaml:"write_timeout"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

var (
	instance *Config
	once     sync.Once
)

// Load 加载配置文件
// 优先级：环境变量 > 配置文件 > 默认值
func Load(path string) (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port: 8080,
			Mode: "release",
		},
		Database: DatabaseConfig{
			Path:         "./data/claw.db",
			MaxOpenConns: 10,
			MaxIdleConns: 5,
		},
		JWT: JWTConfig{
			Secret:      "change-me-in-production",
			ExpireHours: 24,
		},
		WebSocket: WebSocketConfig{
			HeartbeatInterval: 30,
			WriteTimeout:      10,
		},
		Log: LogConfig{
			Level:  "info",
			Format: "json",
		},
	}

	// 如果指定了配置文件路径，尝试加载
	if path != "" {
		if err := cfg.loadFromFile(path); err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("加载配置文件失败: %w", err)
		}
	}

	// 环境变量覆盖
	cfg.loadFromEnv()

	// 验证配置
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	return cfg, nil
}

// Get 获取单例配置实例
func Get() *Config {
	if instance == nil {
		// 如果未初始化，使用默认配置加载
		cfg, err := Load("config.yaml")
		if err != nil {
			// 加载失败时使用默认配置
			cfg = &Config{
				Server: ServerConfig{Port: 8080, Mode: "release"},
				Database: DatabaseConfig{
					Path:         "./data/claw.db",
					MaxOpenConns: 10,
					MaxIdleConns: 5,
				},
				JWT: JWTConfig{Secret: "change-me", ExpireHours: 24},
				WebSocket: WebSocketConfig{HeartbeatInterval: 30, WriteTimeout: 10},
				Log: LogConfig{Level: "info", Format: "json"},
			}
		}
		instance = cfg
	}
	return instance
}

// Init 初始化配置（带错误处理）
func Init(path string) error {
	var err error
	once.Do(func() {
		instance, err = Load(path)
	})
	return err
}

// loadFromFile 从 YAML 文件加载配置
func (c *Config) loadFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, c)
}

// loadFromEnv 从环境变量加载配置
// 环境变量命名规则：CLAW_<SECTION>_<KEY>，大写，下划线分隔
func (c *Config) loadFromEnv() {
	// Server 配置
	if v := os.Getenv("CLAW_SERVER_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			c.Server.Port = port
		}
	}
	if v := os.Getenv("CLAW_SERVER_MODE"); v != "" {
		c.Server.Mode = v
	}

	// Database 配置
	if v := os.Getenv("CLAW_DATABASE_PATH"); v != "" {
		c.Database.Path = v
	}
	if v := os.Getenv("CLAW_DATABASE_MAX_OPEN_CONNS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.Database.MaxOpenConns = n
		}
	}
	if v := os.Getenv("CLAW_DATABASE_MAX_IDLE_CONNS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.Database.MaxIdleConns = n
		}
	}

	// JWT 配置
	if v := os.Getenv("CLAW_JWT_SECRET"); v != "" {
		c.JWT.Secret = v
	}
	if v := os.Getenv("CLAW_JWT_EXPIRE_HOURS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.JWT.ExpireHours = n
		}
	}

	// WebSocket 配置
	if v := os.Getenv("CLAW_WEBSOCKET_HEARTBEAT_INTERVAL"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.WebSocket.HeartbeatInterval = n
		}
	}
	if v := os.Getenv("CLAW_WEBSOCKET_WRITE_TIMEOUT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.WebSocket.WriteTimeout = n
		}
	}

	// Log 配置
	if v := os.Getenv("CLAW_LOG_LEVEL"); v != "" {
		c.Log.Level = v
	}
	if v := os.Getenv("CLAW_LOG_FORMAT"); v != "" {
		c.Log.Format = v
	}
}

// Validate 验证配置有效性
func (c *Config) Validate() error {
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("服务器端口必须在 1-65535 之间: %d", c.Server.Port)
	}
	if c.Server.Mode != "debug" && c.Server.Mode != "release" {
		return fmt.Errorf("服务器模式必须是 debug 或 release: %s", c.Server.Mode)
	}
	if c.JWT.Secret == "" || c.JWT.Secret == "change-me-in-production" || c.JWT.Secret == "change-me" {
		// 生产环境警告，但不阻止启动
		fmt.Fprintf(os.Stderr, "警告: JWT Secret 使用默认值，请在生产环境修改\n")
	}
	if c.JWT.ExpireHours < 1 {
		return fmt.Errorf("JWT 过期时间必须大于 0 小时")
	}
	if c.Log.Level != "debug" && c.Log.Level != "info" && c.Log.Level != "warn" && c.Log.Level != "error" {
		return fmt.Errorf("日志级别必须是 debug/info/warn/error: %s", c.Log.Level)
	}
	return nil
}

// EnsureDataDir 确保数据目录存在
func (c *Config) EnsureDataDir() error {
	dir := filepath.Dir(c.Database.Path)
	if dir == "" || dir == "." {
		return nil
	}
	return os.MkdirAll(dir, 0755)
}

// IsDebug 是否为调试模式
func (c *Config) IsDebug() bool {
	return strings.ToLower(c.Server.Mode) == "debug"
}

// DSN 返回数据库连接字符串
func (c *Config) DSN() string {
	return c.Database.Path
}
