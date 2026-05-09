import request from '../utils/request';

export interface Channel {
  id: string;
  name: string;
  description: string;
  type: 'public' | 'private' | 'direct';
  status: 'active' | 'archived';
  created_by: string;
  creator_name: string;
  member_count: number;
  unread_count?: number;
  created_at: string;
  updated_at: string;
}

export interface ChannelMember {
  channel_id: string;
  employee_id: string;
  employee_name?: string;
  role: 'owner' | 'admin' | 'member' | 'readonly';
  joined_at: string;
  employee?: {
    id: string;
    name: string;
    type: string;
    email: string;
    skills: string[];
  };
}

export interface CreateChannelRequest {
  name: string;
  description?: string;
  type: 'public' | 'private' | 'direct';
  member_ids?: string[];
  parent_id?: string;  // 父频道ID，用于创建子频道
}

export interface ChannelNode {
  id: string;
  name: string;
  parent_id: string | null;
  path: string;
  depth: number;
  type: 'public' | 'private' | 'direct';
  description: string;
  doc_count: number;
  child_count: number;
  children: ChannelNode[];
  created_at: string;
  updated_at: string;
}

export interface ChannelTreeResponse {
  channels: ChannelNode[];
}

export interface UpdateChannelRequest {
  name?: string;
  description?: string;
  status?: 'active' | 'archived';
}

export interface ListChannelsParams {
  page?: number;
  pageSize?: number;
  type?: string;
  keyword?: string;
}

export const channelApi = {
  // 获取频道树形结构
  getTree: (): Promise<any> => {
    return request.get('/channels/tree');
  },

  // 获取频道列表（支持筛选）
  list: (params: ListChannelsParams = {}): Promise<any> => {
    const { page = 1, pageSize = 20, type, keyword } = params;
    let url = `/channels?page=${page}&page_size=${pageSize}`;
    if (type) url += `&type=${encodeURIComponent(type)}`;
    if (keyword) url += `&keyword=${encodeURIComponent(keyword)}`;
    return request.get(url);
  },

  // 获取我的频道
  myChannels: (): Promise<any> => {
    return request.get('/channels/my');
  },

  // 创建频道
  create: (data: CreateChannelRequest): Promise<any> => {
    return request.post('/channels', data);
  },

  // 获取频道详情
  getById: (id: string): Promise<any> => {
    return request.get(`/channels/${id}`);
  },

  // 更新频道
  update: (id: string, data: UpdateChannelRequest): Promise<any> => {
    return request.put(`/channels/${id}`, data);
  },

  // 删除频道
  delete: (id: string): Promise<any> => {
    return request.delete(`/channels/${id}`);
  },

  // 获取频道成员
  getMembers: (id: string): Promise<any> => {
    return request.get(`/channels/${id}/members`);
  },

  // 添加成员
  addMember: (id: string, employeeId: string, role: string = 'member'): Promise<any> => {
    return request.post(`/channels/${id}/members`, { employee_id: employeeId, role });
  },

  // 移除成员
  removeMember: (id: string, employeeId: string): Promise<any> => {
    return request.delete(`/channels/${id}/members/${employeeId}`);
  },

  // 更新成员角色
  updateMemberRole: (id: string, employeeId: string, role: string): Promise<any> => {
    return request.put(`/channels/${id}/members/${employeeId}`, { role });
  },

  // 加入频道
  join: (id: string): Promise<any> => {
    return request.post(`/channels/${id}/join`);
  },

  // 离开频道
  leave: (id: string): Promise<any> => {
    return request.post(`/channels/${id}/leave`);
  },
};
