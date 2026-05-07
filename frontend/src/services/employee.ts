import request from '../utils/request';

export interface NotificationChannels {
  email: boolean;
  webhook: boolean;
  internal: boolean;
}

export interface EventNotifications {
  task_assigned: boolean;
  task_completed: boolean;
  task_cancelled: boolean;
  workflow_triggered: boolean;
  workflow_completed: boolean;
  workflow_failed: boolean;
  mention_received: boolean;
  channel_message: boolean;
}

export interface NotificationPreferences {
  channels: NotificationChannels;
  events: EventNotifications;
}

export interface Employee {
  id: string;
  username: string;
  display_name: string;
  name: string;  // 兼容旧字段
  email: string;
  type: 'human' | 'agent';
  role: string;
  skills: string[];
  status: 'active' | 'inactive';
  api_key?: string;

  // 扩展资料
  avatar?: string;
  department?: string;
  position?: string;
  phone?: string;

  // 通知偏好
  notification_prefs?: NotificationPreferences;

  created_at: string;
  updated_at: string;
}

export interface CreateEmployeeRequest {
  username: string;
  display_name: string;
  name?: string;  // 兼容旧字段
  email: string;
  password?: string;
  type: 'human' | 'agent';
  role?: string;
  skills?: string[];

  // 扩展资料
  avatar?: string;
  department?: string;
  position?: string;
  phone?: string;

  // 通知偏好
  notification_prefs?: NotificationPreferences;
}

export interface UpdateEmployeeRequest {
  username?: string;
  display_name?: string;
  name?: string;
  email?: string;
  role?: string;
  skills?: string[];
  status?: 'active' | 'inactive';

  // 扩展资料
  avatar?: string;
  department?: string;
  position?: string;
  phone?: string;

  // 通知偏好
  notification_prefs?: NotificationPreferences;
}

export interface ListEmployeesParams {
  page?: number;
  pageSize?: number;
  type?: string;
  status?: string;
  role?: string;
  keyword?: string;
}

export interface UpdateNotificationPrefsRequest {
  channels: NotificationChannels;
  events: EventNotifications;
}

export const employeeApi = {
  // 获取员工列表（支持筛选）
  list: (params: ListEmployeesParams = {}): Promise<any> => {
    const { page = 1, pageSize = 20, type, status, role, keyword } = params;
    let url = `/employees?page=${page}&page_size=${pageSize}`;
    if (type) url += `&type=${encodeURIComponent(type)}`;
    if (status) url += `&status=${encodeURIComponent(status)}`;
    if (role) url += `&role=${encodeURIComponent(role)}`;
    if (keyword) url += `&keyword=${encodeURIComponent(keyword)}`;
    return request.get(url);
  },

  // 创建员工
  create: (data: CreateEmployeeRequest): Promise<any> => {
    return request.post('/employees', data);
  },

  // 获取员工详情
  getById: (id: string): Promise<any> => {
    return request.get(`/employees/${id}`);
  },

  // 更新员工
  update: (id: string, data: UpdateEmployeeRequest): Promise<any> => {
    return request.put(`/employees/${id}`, data);
  },

  // 删除员工
  delete: (id: string): Promise<any> => {
    return request.delete(`/employees/${id}`);
  },

  // 搜索员工
  search: (skills: string[]): Promise<any> => {
    const params = skills.map(s => `skills=${encodeURIComponent(s)}`).join('&');
    return request.get(`/employees/search?${params}`);
  },

  // 更新通知偏好
  updateNotificationPrefs: (id: string, data: UpdateNotificationPrefsRequest): Promise<any> => {
    return request.put(`/employees/${id}/notification-prefs`, data);
  },

  // 生成 API Key
  generateApiKey: (id: string): Promise<any> => {
    return request.post(`/employees/${id}/apikey`);
  },

  // 重置 API Key
  resetApiKey: (id: string): Promise<any> => {
    return request.put(`/employees/${id}/apikey`);
  },

  // 获取所有已有的职能值
  getDistinctRoles: (): Promise<any> => {
    return request.get('/employees/roles');
  },
};
