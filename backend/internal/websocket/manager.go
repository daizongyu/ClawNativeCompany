// Package websocket 提供 WebSocket 连接管理
package websocket

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"claw/internal/jwt"
	"claw/internal/logger"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var (
	// Upgrader WebSocket 升级器
	Upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			// 允许所有来源，生产环境应该限制
			return true
		},
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

// Client WebSocket 客户端
type Client struct {
	ID         string          // 客户端 ID（员工 ID）
	Conn       *websocket.Conn // WebSocket 连接
	Send       chan []byte     // 发送通道
	Channels   map[string]bool // 订阅的频道
	mu         sync.RWMutex    // 保护 Channels
}

// NewClient 创建新客户端
func NewClient(id string, conn *websocket.Conn) *Client {
	return &Client{
		ID:       id,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		Channels: make(map[string]bool),
	}
}

// Subscribe 订阅频道
func (c *Client) Subscribe(channelID string) {
	c.mu.Lock()
	c.Channels[channelID] = true
	c.mu.Unlock()
}

// Unsubscribe 取消订阅频道
func (c *Client) Unsubscribe(channelID string) {
	c.mu.Lock()
	delete(c.Channels, channelID)
	c.mu.Unlock()
}

// IsSubscribed 检查是否订阅了频道
func (c *Client) IsSubscribed(channelID string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Channels[channelID]
}

// Manager WebSocket 管理器
type Manager struct {
	clients    map[string]*Client // 所有连接的客户端（key: employeeID）
	broadcast  chan []byte        // 广播通道
	register   chan *Client       // 注册通道
	unregister chan *Client       // 注销通道
	mu         sync.RWMutex       // 保护 clients
}

// NewManager 创建 WebSocket 管理器
func NewManager() *Manager {
	return &Manager{
		clients:    make(map[string]*Client),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run 启动管理器
func (m *Manager) Run() {
	for {
		select {
		case client := <-m.register:
			m.mu.Lock()
			m.clients[client.ID] = client
			m.mu.Unlock()
			logger.Info("WebSocket 客户端连接", "client_id", client.ID, "total", len(m.clients))

		case client := <-m.unregister:
			m.mu.Lock()
			if _, ok := m.clients[client.ID]; ok {
				delete(m.clients, client.ID)
				close(client.Send)
			}
			m.mu.Unlock()
			logger.Info("WebSocket 客户端断开", "client_id", client.ID, "total", len(m.clients))

		case message := <-m.broadcast:
			m.broadcastToAll(message)
		}
	}
}

// broadcastToAll 广播给所有客户端
func (m *Manager) broadcastToAll(message []byte) {
	m.mu.RLock()
	clients := make([]*Client, 0, len(m.clients))
	for _, client := range m.clients {
		clients = append(clients, client)
	}
	m.mu.RUnlock()

	for _, client := range clients {
		select {
		case client.Send <- message:
		default:
			// 发送通道已满，关闭连接
			m.unregisterClient(client)
		}
	}
}

// unregisterClient 注销客户端
func (m *Manager) unregisterClient(client *Client) {
	m.mu.Lock()
	if _, ok := m.clients[client.ID]; ok {
		delete(m.clients, client.ID)
		close(client.Send)
		client.Conn.Close()
	}
	m.mu.Unlock()
}

// BroadcastToChannel 广播给频道内的所有客户端
func (m *Manager) BroadcastToChannel(channelID string, message []byte) {
	m.mu.RLock()
	clients := make([]*Client, 0)
	for _, client := range m.clients {
		if client.IsSubscribed(channelID) {
			clients = append(clients, client)
		}
	}
	m.mu.RUnlock()

	for _, client := range clients {
		select {
		case client.Send <- message:
		default:
			m.unregisterClient(client)
		}
	}
}

// BroadcastToUser 发送给指定用户
func (m *Manager) BroadcastToUser(userID string, message []byte) {
	m.mu.RLock()
	client, ok := m.clients[userID]
	m.mu.RUnlock()

	if !ok {
		return
	}

	select {
	case client.Send <- message:
	default:
		m.unregisterClient(client)
	}
}

// GetOnlineCount 获取在线客户端数量
func (m *Manager) GetOnlineCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.clients)
}

// IsOnline 检查用户是否在线
func (m *Manager) IsOnline(userID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.clients[userID]
	return ok
}

// HandleWebSocket 处理 WebSocket 连接
func (m *Manager) HandleWebSocket(c *gin.Context) {
	// 从 URL 参数获取 token
	token := c.Query("token")
	if token == "" {
		// 尝试从 Authorization header 获取
		authHeader := c.GetHeader("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}

	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供 token"})
		return
	}

	// 验证 token
	claims, err := jwt.ParseToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的 token"})
		return
	}

	userIDStr := claims.EmployeeID

	// 升级连接
	conn, err := Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("WebSocket 升级失败", "error", err)
		return
	}

	// 创建客户端
	client := NewClient(userIDStr, conn)

	// 注册客户端
	m.register <- client

	// 启动读写 goroutine
	go client.writePump()
	go client.readPump(m)
}

// readPump 读取客户端消息
func (c *Client) readPump(m *Manager) {
	defer func() {
		m.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(512 * 1024) // 512KB 限制
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Error("WebSocket 读取错误", "error", err, "client_id", c.ID)
			}
			break
		}

		// 处理客户端消息（订阅/取消订阅等）
		m.handleClientMessage(c, message)
	}
}

// writePump 向客户端写入消息
func (c *Client) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			c.Conn.WriteMessage(websocket.TextMessage, message)

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ClientMessage 客户端消息
type ClientMessage struct {
	Type      string `json:"type"`       // subscribe, unsubscribe, ping
	ChannelID string `json:"channel_id"` // 频道 ID
}

// handleClientMessage 处理客户端消息
func (m *Manager) handleClientMessage(client *Client, data []byte) {
	var msg ClientMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		logger.Warn("解析客户端消息失败", "error", err, "client_id", client.ID)
		return
	}

	switch msg.Type {
	case "subscribe":
		if msg.ChannelID != "" {
			client.Subscribe(msg.ChannelID)
			logger.Info("客户端订阅频道", "client_id", client.ID, "channel_id", msg.ChannelID)
		}
	case "unsubscribe":
		if msg.ChannelID != "" {
			client.Unsubscribe(msg.ChannelID)
			logger.Info("客户端取消订阅频道", "client_id", client.ID, "channel_id", msg.ChannelID)
		}
	case "ping":
		// 发送 pong 响应
		pong := map[string]string{"type": "pong", "time": time.Now().Format(time.RFC3339)}
		data, _ := json.Marshal(pong)
		client.Send <- data
	}
}

// BroadcastMessage 广播消息结构
type BroadcastMessage struct {
	Type      string          `json:"type"`       // message, system, notification
	ChannelID string          `json:"channel_id"` // 频道 ID
	Data      json.RawMessage `json:"data"`       // 消息数据
	Timestamp time.Time       `json:"timestamp"`
}

// NewBroadcastMessage 创建广播消息
func NewBroadcastMessage(msgType, channelID string, data interface{}) ([]byte, error) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	msg := BroadcastMessage{
		Type:      msgType,
		ChannelID: channelID,
		Data:      dataBytes,
		Timestamp: time.Now(),
	}

	return json.Marshal(msg)
}

// 全局 WebSocket 管理器实例
var globalManager *Manager

// Init 初始化 WebSocket 管理器
func Init() {
	globalManager = NewManager()
	go globalManager.Run()
	logger.Info("WebSocket 管理器已启动")
}

// GetManager 获取全局管理器
func GetManager() *Manager {
	return globalManager
}

// BroadcastToChannel 便捷函数：广播给频道
func BroadcastToChannel(channelID string, message []byte) {
	if globalManager != nil {
		globalManager.BroadcastToChannel(channelID, message)
	}
}

// BroadcastToUser 便捷函数：发送给用户
func BroadcastToUser(userID string, message []byte) {
	if globalManager != nil {
		globalManager.BroadcastToUser(userID, message)
	}
}
