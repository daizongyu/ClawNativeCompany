import request from '../utils/request';

export interface Message {
  id: string;
  channel_id: string;
  sender_id: string;
  sender_name: string;
  sender_type: 'human' | 'agent';
  content: string;
  content_type: 'text' | 'image' | 'file' | 'system';
  parent_id?: string;
  created_at: string;
  updated_at: string;
}

export interface SendMessageRequest {
  channel_id: string;
  content: string;
  content_type?: 'text' | 'image' | 'file';
  parent_id?: string;
}

export const messageApi = {
  // 发送消息
  send: (data: SendMessageRequest): Promise<any> => {
    return request.post('/messages', data);
  },

  // 获取频道消息列表
  listByChannel: (channelId: string, page: number = 1, pageSize: number = 50): Promise<any> => {
    return request.get(`/messages?channel_id=${channelId}&page=${page}&page_size=${pageSize}`);
  },

  // 获取消息历史（分页）
  getHistory: (channelId: string, before?: string, limit: number = 50): Promise<any> => {
    let url = `/channels/${channelId}/messages?limit=${limit}`;
    if (before) {
      url += `&before=${before}`;
    }
    return request.get(url);
  },

  // 撤回消息
  recall: (messageId: string): Promise<any> => {
    return request.post(`/messages/${messageId}/recall`);
  },

  // 获取消息详情
  getById: (messageId: string): Promise<any> => {
    return request.get(`/messages/${messageId}`);
  },

  // 回复消息
  reply: (parentId: string, content: string, channelId: string): Promise<any> => {
    return request.post('/messages', { channel_id: channelId, content, parent_id: parentId });
  },
};
