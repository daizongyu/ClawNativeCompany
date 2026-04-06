import React from 'react';
import { Layout, Menu } from 'antd';
import {
  DashboardOutlined,
  TeamOutlined,
  MessageOutlined,
  CheckSquareOutlined,
  NodeIndexOutlined,
} from '@ant-design/icons';
import { useNavigate, useLocation } from 'react-router-dom';

const { Sider } = Layout;

const Sidebar: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();

  const menuItems = [
    {
      key: '/',
      icon: <DashboardOutlined />,
      label: '仪表盘',
      'data-testid': 'nav-dashboard',
    },
    {
      key: '/employees',
      icon: <TeamOutlined />,
      label: '员工管理',
      'data-testid': 'nav-employees',
    },
    {
      key: '/channels',
      icon: <MessageOutlined />,
      label: '频道',
      'data-testid': 'nav-channels',
    },
    {
      key: '/tasks',
      icon: <CheckSquareOutlined />,
      label: '任务',
      'data-testid': 'nav-tasks',
    },
    {
      key: '/workflows',
      icon: <NodeIndexOutlined />,
      label: '工作流',
      'data-testid': 'nav-workflows',
    },
  ];

  const handleMenuClick = (key: string) => {
    navigate(key);
  };

  return (
    <Sider
      width={200}
      theme="light"
      style={{ borderRight: '1px solid #f0f0f0' }}
      data-testid="sidebar"
    >
      <div
        style={{
          height: '64px',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          borderBottom: '1px solid #f0f0f0',
        }}
        data-testid="sidebar-logo"
      >
        <h2 style={{ margin: 0, fontSize: '18px', color: '#1890ff' }}>Claw</h2>
      </div>
      <Menu
        mode="inline"
        selectedKeys={[location.pathname]}
        items={menuItems}
        onClick={({ key }) => handleMenuClick(key)}
        style={{ borderRight: 0 }}
        data-testid="sidebar-menu"
      />
    </Sider>
  );
};

export default Sidebar;
