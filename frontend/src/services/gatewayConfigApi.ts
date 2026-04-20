import request from '../utils/request';

export type GatewayType = 'dingtalk' | 'slack' | 'custom';
export type GatewayStatus = 'active' | 'inactive' | 'error';

export interface GatewayConfig {
  id: string;
  employee_id: string;
  type: GatewayType;
  name: string;
  status: GatewayStatus;
  is_default: boolean;

  // 钉钉配置
  app_key?: string;
  app_secret?: string;
  agent_id?: string;

  // Slack 配置
  webhook_url?: string;
  bot_token?: string;
  default_channel?: string;

  // 自定义配置
  auth_type?: 'none' | 'bearer' | 'basic';
  auth_token?: string;

  created_at: string;
  updated_at: string;
}

export interface GatewayConfigListResponse {
  list: GatewayConfig[];
  pagination: {
    page: number;
    page_size: number;
    total: number;
    total_page: number;
  };
}

export const gatewayConfigApi = {
  // 创建 Gateway 配置
  create: (data: Partial<GatewayConfig>): Promise<any> => {
    return request.post('/gateway-configs', data);
  },

  // 获取 Gateway 配置列表
  list: (params?: { type?: string; status?: string; page?: number; page_size?: number }): Promise<any> => {
    const { type, status, page = 1, page_size = 20 } = params || {};
    let url = `/gateway-configs?page=${page}&page_size=${page_size}`;
    if (type) url += `&type=${encodeURIComponent(type)}`;
    if (status) url += `&status=${encodeURIComponent(status)}`;
    return request.get(url);
  },

  // 获取单个 Gateway 配置
  getById: (id: string): Promise<any> => {
    return request.get(`/gateway-configs/${id}`);
  },

  // 更新 Gateway 配置
  update: (id: string, data: Partial<GatewayConfig>): Promise<any> => {
    return request.put(`/gateway-configs/${id}`, data);
  },

  // 删除 Gateway 配置
  delete: (id: string): Promise<any> => {
    return request.delete(`/gateway-configs/${id}`);
  },

  // 验证 Gateway 配置
  verify: (id: string): Promise<any> => {
    return request.post(`/gateway-configs/${id}/verify`);
  },

  // 发送测试消息
  test: (id: string): Promise<any> => {
    return request.post(`/gateway-configs/${id}/test`);
  },

  // 设为默认配置
  setDefault: (id: string): Promise<any> => {
    return request.post(`/gateway-configs/${id}/default`);
  },
};
