import request from '../utils/request';

export const taskApi = {
  list: (page: number, pageSize: number): Promise<any> => {
    return request.get(`/tasks?page=${page}&page_size=${pageSize}`);
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
