import request from '../utils/request';

export const taskApi = {
  list: (params: any): Promise<any> => {
    const queryParams = new URLSearchParams();
    if (params.page) queryParams.append('page', params.page.toString());
    if (params.page_size) queryParams.append('page_size', params.page_size.toString());
    if (params.status) queryParams.append('status', params.status);
    if (params.priority) queryParams.append('priority', params.priority);
    if (params.keyword) queryParams.append('keyword', params.keyword);
    if (params.mine) queryParams.append('mine', 'true');
    if (params.unclaimed) queryParams.append('unclaimed', 'true');
    return request.get(`/tasks?${queryParams.toString()}`);
  },

  create: (data: any): Promise<any> => {
    return request.post('/tasks', data);
  },

  getById: (id: string): Promise<any> => {
    return request.get(`/tasks/${id}`);
  },

  update: (id: string, data: any): Promise<any> => {
    return request.put(`/tasks/${id}`, data);
  },

  delete: (id: string): Promise<any> => {
    return request.delete(`/tasks/${id}`);
  },

  claim: (id: string): Promise<any> => {
    return request.post(`/tasks/${id}/claim`, {});
  },

  assign: (id: string, assigneeId: string): Promise<any> => {
    return request.post(`/tasks/${id}/assign`, { assignee_id: assigneeId });
  },

  complete: (id: string, result: any): Promise<any> => {
    return request.post(`/tasks/${id}/complete`, { result });
  },

  cancel: (id: string): Promise<any> => {
    return request.post(`/tasks/${id}/cancel`, {});
  },

  getMyTasks: (): Promise<any> => {
    return request.get('/tasks/my');
  },

  getUnclaimed: (): Promise<any> => {
    return request.get('/tasks/unclaimed');
  },
};
