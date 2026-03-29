import request from '../utils/request';

interface LoginResponse {
  code: number;
  message: string;
  data: {
    token: string;
    user: {
      id: string;
      name: string;
      email: string;
      type: string;
      role: string;
    };
  };
}

export const authApi = {
  login: (email: string, password: string): Promise<LoginResponse> => {
    return request.post('/auth/login', { email, password });
  },

  logout: (): Promise<{ code: number; message: string }> => {
    return request.post('/auth/logout');
  },

  getProfile: (): Promise<any> => {
    return request.get('/auth/profile');
  },
};
