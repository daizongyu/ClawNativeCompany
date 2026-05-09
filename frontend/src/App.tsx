import React, { useEffect } from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { ConfigProvider, App as AntdApp } from 'antd';
import zhCN from 'antd/locale/zh_CN';
import { useAuthStore } from './stores/auth';
import MainLayout from './components/layout/MainLayout';
import Login from './pages/Login';
import Dashboard from './pages/Dashboard';
import Employees from './pages/Employees';
import Documents from './pages/Documents';
import Tasks from './pages/Tasks';
import Workflows from './pages/Workflows/index';
import WorkflowEditor from './pages/WorkflowEditor';
import ExecutionHistory from './pages/ExecutionHistory';
import Profile from './pages/Profile';

// 受保护的路由组件
const ProtectedRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { isAuthenticated } = useAuthStore();
  return isAuthenticated ? <>{children}</> : <Navigate to="/login" replace />;
};

// 带布局的路由包装器
const LayoutRoute: React.FC<{ element: React.ReactElement }> = ({ element }) => {
  return (
    <MainLayout>
      {element}
    </MainLayout>
  );
};

// 应用主组件
const App: React.FC = () => {
  // 初始化测试工具
  useEffect(() => {
    // 动态导入测试工具，确保在客户端环境
    if (typeof window !== 'undefined') {
      // 初始化消息拦截器
      import('./utils/messageInterceptor').then((module) => {
        if (module.messageInterceptor) {
          module.messageInterceptor.init();
        }
      }).catch(() => {
        // 忽略错误，测试工具是可选的
      });

      // 初始化测试 API 暴露
      import('./utils/testExposer').then((module) => {
        if (module.testExposer) {
          module.testExposer.init();
        }
      }).catch(() => {
        // 忽略错误，测试工具是可选的
      });
    }
  }, []);

  return (
    <ConfigProvider locale={zhCN}>
      <AntdApp>
        <BrowserRouter>
          <Routes>
            <Route path="/login" element={<Login />} />
            <Route
              path="/"
              element={
                <ProtectedRoute>
                  <LayoutRoute element={<Dashboard />} />
                </ProtectedRoute>
              }
            />
            <Route
              path="/employees"
              element={
                <ProtectedRoute>
                  <LayoutRoute element={<Employees />} />
                </ProtectedRoute>
              }
            />
            <Route
              path="/documents"
              element={
                <ProtectedRoute>
                  <LayoutRoute element={<Documents />} />
                </ProtectedRoute>
              }
            />
            <Route
              path="/tasks"
              element={
                <ProtectedRoute>
                  <LayoutRoute element={<Tasks />} />
                </ProtectedRoute>
              }
            />
            <Route
              path="/workflows"
              element={
                <ProtectedRoute>
                  <LayoutRoute element={<Workflows />} />
                </ProtectedRoute>
              }
            />
            <Route
              path="/workflows/editor/:id?"
              element={
                <ProtectedRoute>
                  <LayoutRoute element={<WorkflowEditor />} />
                </ProtectedRoute>
              }
            />
            <Route
              path="/workflows/executions/:workflowId?"
              element={
                <ProtectedRoute>
                  <LayoutRoute element={<ExecutionHistory />} />
                </ProtectedRoute>
              }
            />
            <Route
              path="/profile"
              element={
                <ProtectedRoute>
                  <LayoutRoute element={<Profile />} />
                </ProtectedRoute>
              }
            />
          </Routes>
        </BrowserRouter>
      </AntdApp>
    </ConfigProvider>
  );
};

export default App;
