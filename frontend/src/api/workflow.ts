import request from '../utils/request';

export const workflowApi = {
  list: (page: number, pageSize: number): Promise<any> => {
    return request.get(`/workflows?page=${page}&page_size=${pageSize}`);
  },

  create: (data: any): Promise<any> => {
    return request.post('/workflows', data);
  },

  getById: (id: string): Promise<any> => {
    return request.get(`/workflows/${id}`);
  },

  update: (id: string, data: any): Promise<any> => {
    return request.put(`/workflows/${id}`, data);
  },

  delete: (id: string): Promise<any> => {
    return request.delete(`/workflows/${id}`);
  },

  updateStatus: (id: string, status: string): Promise<any> => {
    return request.patch(`/workflows/${id}/status`, { status });
  },

  trigger: (id: string, input: any): Promise<any> => {
    return request.post(`/workflows/${id}/trigger`, { input });
  },

  getExecutions: (workflowId: string, page: number, pageSize: number): Promise<any> => {
    return request.get(`/workflows/${workflowId}/executions?page=${page}&page_size=${pageSize}`);
  },
};
