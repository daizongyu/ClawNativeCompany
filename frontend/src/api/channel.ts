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
}

export interface UpdateChannelRequest {
  name?: string;
  description?: string;
  status?: 'active' | 'archived';
}

export const channelApi = {
  // 获取频道列表
  list: (page: number = 1, pageSize: number = 20): Promise<any> => {
    return request.get(`/channels?page=${page}&page_size=${pageSize}`);
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
