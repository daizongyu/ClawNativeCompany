import React, { useState, useEffect } from 'react';
import { Form, Input, Button, Card, message } from 'antd';
import { UserOutlined, LockOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useAuthStore } from '../stores/auth';
import { authApi } from '../services/auth';
import { PageContainer } from '../components/common';

const Login: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();
  const { setToken, setUser } = useAuthStore();

  // 设置当前页面
  useEffect(() => {
    if (typeof window !== 'undefined' && window.__CLAW_TEST__) {
      window.__CLAW_TEST__.setCurrentPage('login');
    }
  }, []);

  const onFinish = async (values: { email: string; password: string }) => {
    setLoading(true);
    try {
      const res = await authApi.login(values.email, values.password);
      if (res.code === 0) {
        setToken(res.data.access_token);
        setUser(res.data.employee);
        message.success('登录成功');
        navigate('/');
      } else {
        message.error(res.message || '登录失败');
      }
    } catch (error: any) {
      // 错误信息已在 request 拦截器中显示
      // 这里只需要处理未被拦截的错误
      if (error?.response?.data?.message) {
        message.error(error.response.data.message);
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <PageContainer
      data-testid="page-login"
      data-page="login"
      loading={false}
      style={{
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        minHeight: '100vh',
        background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
      }}
    >
      <Card
        title="Claw Native - 登录"
        style={{ width: 400, boxShadow: '0 4px 20px rgba(0,0,0,0.1)' }}
        data-testid="login-card"
      >
        <Form
          name="login"
          onFinish={onFinish}
          autoComplete="off"
          layout="vertical"
          data-testid="login-form"
        >
          <Form.Item
            label="邮箱"
            name="email"
            rules={[
              { required: true, message: '请输入邮箱' },
              { type: 'email', message: '请输入有效的邮箱地址' },
            ]}
          >
            <Input
              prefix={<UserOutlined />}
              placeholder="请输入邮箱"
              size="large"
              data-testid="input-email"
              data-input-name="email"
            />
          </Form.Item>

          <Form.Item
            label="密码"
            name="password"
            rules={[{ required: true, message: '请输入密码' }]}
          >
            <Input.Password
              prefix={<LockOutlined />}
              placeholder="请输入密码"
              size="large"
              data-testid="input-password"
              data-input-name="password"
            />
          </Form.Item>

          <Form.Item>
            <Button
              type="primary"
              htmlType="submit"
              loading={loading}
              size="large"
              block
              data-testid="login-submit-btn"
              data-action="login"
              data-entity="auth"
            >
              登录
            </Button>
          </Form.Item>
        </Form>
      </Card>
    </PageContainer>
  );
};

export default Login;
