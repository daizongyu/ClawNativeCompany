import request from '../utils/request';

export interface Employee {
  id: string;
  name: string;
  email: string;
  type: 'human' | 'agent';
  role: string;
  skills: string[];
  status: 'active' | 'inactive';
  api_key?: string;
  created_at: string;
  updated_at: string;
}

export interface CreateEmployeeRequest {
  name: string;
  email: string;
  password?: string;
  type: 'human' | 'agent';
  role?: string;
  skills?: string[];
}

export interface UpdateEmployeeRequest {
  name?: string;
  email?: string;
  role?: string;
  skills?: string[];
  status?: 'active' | 'inactive';
}

export const employeeApi = {
  // 获取员工列表
  list: (page: number = 1, pageSize: number = 20): Promise<any> => {
    return request.get(`/employees?page=${page}&page_size=${pageSize}`);
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

  // 生成 API Key
  generateApiKey: (id: string): Promise<any> => {
    return request.post(`/employees/${id}/apikey`);
  },

  // 重置 API Key
  resetApiKey: (id: string): Promise<any> => {
    return request.put(`/employees/${id}/apikey`);
  },
};
