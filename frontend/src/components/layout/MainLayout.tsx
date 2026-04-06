import React from 'react';
import { Layout } from 'antd';
import Sidebar from './Sidebar';
import Header from './Header';

const { Content } = Layout;

interface MainLayoutProps {
  children: React.ReactNode;
}

const MainLayout: React.FC<MainLayoutProps> = ({ children }) => {
  return (
    <Layout style={{ minHeight: '100vh' }} data-testid="main-layout">
      <Sidebar />
      <Layout data-testid="main-layout-content">
        <Header />
        <Content
          style={{ margin: '24px', padding: '24px', background: '#fff', borderRadius: '8px' }}
          data-testid="page-content"
        >
          {children}
        </Content>
      </Layout>
    </Layout>
  );
};

export default MainLayout;
