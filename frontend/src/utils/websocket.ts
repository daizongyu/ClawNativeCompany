import { useAuthStore } from '../stores/auth';

export type WebSocketStatus = 'connecting' | 'connected' | 'disconnected' | 'error';

export interface WebSocketMessage {
  type: string;
  data: any;
  timestamp: number;
}

export class WebSocketManager {
  private ws: WebSocket | null = null;
  private url: string;
  private reconnectAttempts: number = 0;
  private maxReconnectAttempts: number = 5;
  private reconnectDelay: number = 3000;
  private reconnectTimer: NodeJS.Timeout | null = null;
  private messageHandlers: Map<string, Set<(data: any) => void>> = new Map();
  private statusHandlers: Set<(status: WebSocketStatus) => void> = new Set();
  private currentStatus: WebSocketStatus = 'disconnected';

  constructor(url: string) {
    this.url = url;
  }

  // 连接 WebSocket
  connect(): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      return;
    }

    const token = useAuthStore.getState().token;
    if (!token) {
      console.error('WebSocket: No token available');
      this.setStatus('error');
      return;
    }

    this.setStatus('connecting');

    try {
      // 在 URL 中添加 token 作为查询参数
      const wsUrl = `${this.url}?token=${encodeURIComponent(token)}`;
      this.ws = new WebSocket(wsUrl);

      this.ws.onopen = () => {
        console.log('WebSocket: Connected');
        this.reconnectAttempts = 0;
        this.setStatus('connected');
      };

      this.ws.onmessage = (event) => {
        try {
          const message: WebSocketMessage = JSON.parse(event.data);
          this.handleMessage(message);
        } catch (error) {
          console.error('WebSocket: Failed to parse message', error);
        }
      };

      this.ws.onclose = () => {
        console.log('WebSocket: Disconnected');
        this.setStatus('disconnected');
        this.attemptReconnect();
      };

      this.ws.onerror = (error) => {
        console.error('WebSocket: Error', error);
        this.setStatus('error');
      };
    } catch (error) {
      console.error('WebSocket: Failed to connect', error);
      this.setStatus('error');
      this.attemptReconnect();
    }
  }

  // 断开连接
  disconnect(): void {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }

    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }

    this.setStatus('disconnected');
  }

  // 发送消息
  send(type: string, data: any): boolean {
    if (this.ws?.readyState !== WebSocket.OPEN) {
      console.error('WebSocket: Not connected');
      return false;
    }

    try {
      const message: WebSocketMessage = {
        type,
        data,
        timestamp: Date.now(),
      };
      this.ws.send(JSON.stringify(message));
      return true;
    } catch (error) {
      console.error('WebSocket: Failed to send message', error);
      return false;
    }
  }

  // 订阅消息类型
  subscribe(type: string, handler: (data: any) => void): () => void {
    if (!this.messageHandlers.has(type)) {
      this.messageHandlers.set(type, new Set());
    }
    this.messageHandlers.get(type)!.add(handler);

    // 返回取消订阅函数
    return () => {
      this.messageHandlers.get(type)?.delete(handler);
    };
  }

  // 订阅状态变化
  onStatusChange(handler: (status: WebSocketStatus) => void): () => void {
    this.statusHandlers.add(handler);
    // 立即通知当前状态
    handler(this.currentStatus);

    // 返回取消订阅函数
    return () => {
      this.statusHandlers.delete(handler);
    };
  }

  // 获取当前状态
  getStatus(): WebSocketStatus {
    return this.currentStatus;
  }

  // 是否已连接
  isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }

  private setStatus(status: WebSocketStatus): void {
    this.currentStatus = status;
    this.statusHandlers.forEach(handler => handler(status));
  }

  private handleMessage(message: WebSocketMessage): void {
    const handlers = this.messageHandlers.get(message.type);
    if (handlers) {
      handlers.forEach(handler => {
        try {
          handler(message.data);
        } catch (error) {
          console.error(`WebSocket: Handler error for type ${message.type}`, error);
        }
      });
    }
  }

  private attemptReconnect(): void {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error('WebSocket: Max reconnection attempts reached');
      return;
    }

    this.reconnectAttempts++;
    console.log(`WebSocket: Reconnecting in ${this.reconnectDelay}ms (attempt ${this.reconnectAttempts})`);

    this.reconnectTimer = setTimeout(() => {
      this.connect();
    }, this.reconnectDelay);
  }
}

// 创建单例实例
let wsManager: WebSocketManager | null = null;

export const getWebSocketManager = (): WebSocketManager => {
  if (!wsManager) {
    const wsUrl = import.meta.env.VITE_WS_URL || 'ws://localhost:8080/ws';
    wsManager = new WebSocketManager(wsUrl);
  }
  return wsManager;
};

// React Hook for WebSocket
export const useWebSocket = () => {
  const manager = getWebSocketManager();
  return {
    connect: () => manager.connect(),
    disconnect: () => manager.disconnect(),
    send: (type: string, data: any) => manager.send(type, data),
    subscribe: (type: string, handler: (data: any) => void) => manager.subscribe(type, handler),
    onStatusChange: (handler: (status: WebSocketStatus) => void) => manager.onStatusChange(handler),
    isConnected: () => manager.isConnected(),
    getStatus: () => manager.getStatus(),
  };
};

export default WebSocketManager;
