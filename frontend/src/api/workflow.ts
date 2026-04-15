import request from '../utils/request';

export const workflowApi = {
  list: (page?: number, pageSize?: number): Promise<any> => {
    const p = page || 1;
    const ps = pageSize || 100;
    return request.get(`/workflows?page=${p}&page_size=${ps}`);
  },

  execute: (id: string): Promise<any> => {
    return request.post(`/workflows/${id}/trigger`, {});
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
    return request.put(`/workflows/${id}/status`, { status });
  },

  trigger: (id: string, input: any): Promise<any> => {
    return request.post(`/workflows/${id}/trigger`, { input });
  },

  getExecutions: (workflowId: string, page: number, pageSize: number): Promise<any> => {
    return request.get(`/workflows/${workflowId}/executions?page=${page}&page_size=${pageSize}`);
  },

  getExecutionDetail: (executionId: string): Promise<any> => {
    return request.get(`/workflow-executions/${executionId}`);
  },

  retryExecution: (executionId: string): Promise<any> => {
    return request.post(`/workflow-executions/${executionId}/retry`, {});
  },
};
