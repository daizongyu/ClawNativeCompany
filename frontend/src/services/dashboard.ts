import request from '../utils/request';

export const dashboardApi = {
  getStats: (): Promise<any> => {
    return request.get('/dashboard/stats');
  },
};
