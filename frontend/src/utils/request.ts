import axios, { AxiosInstance, AxiosResponse } from 'axios';
import { message } from 'antd';
import { useAuthStore } from '../stores/auth';

const request: AxiosInstance = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// 请求拦截器
request.interceptors.request.use(
  (config) => {
    const token = useAuthStore.getState().token;
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// 响应拦截器
request.interceptors.response.use(
  (response: AxiosResponse) => {
    const { data } = response;
    if (data.code !== 0 && data.code !== undefined) {
      message.error(data.message || '请求失败');
      return Promise.reject(new Error(data.message));
    }
    return data;
  },
  (error) => {
    if (error.response) {
      const { status, data } = error.response;
      // 登录接口的错误特殊处理：只显示消息，不重定向
      const isLoginRequest = error.config?.url?.includes('/auth/login');
      
      if (status === 401) {
        if (isLoginRequest) {
          message.error(data?.message || '邮箱或密码错误');
        } else {
          message.error('登录已过期，请重新登录');
          useAuthStore.getState().logout();
          window.location.href = '/login';
        }
      } else {
        message.error(data?.message || '请求失败');
      }
    } else {
      message.error('网络错误，请检查网络连接');
    }
    return Promise.reject(error);
  }
);

export default request;
