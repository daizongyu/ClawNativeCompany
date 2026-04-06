import React from 'react';
import { Layout, Dropdown, Avatar, Space, Button } from 'antd';
import { LogoutOutlined, UserOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useAuthStore } from '../../stores/auth';

const { Header: AntHeader } = Layout;

const Header: React.FC = () => {
  const navigate = useNavigate();
  const { user, logout } = useAuthStore();

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  const items = [
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: '退出登录',
      onClick: handleLogout,
    },
  ];

  return (
    <AntHeader
      style={{
        background: '#fff',
        padding: '0 24px',
        borderBottom: '1px solid #f0f0f0',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'flex-end',
      }}
      data-testid="header"
    >
      <Dropdown menu={{ items }} placement="bottomRight">
        <Space style={{ cursor: 'pointer' }} data-testid="user-menu">
          <Avatar icon={<UserOutlined />} data-testid="user-avatar" />
          <span data-testid="user-name">{user?.name || '用户'}</span>
        </Space>
      </Dropdown>
      <Button
        type="text"
        icon={<LogoutOutlined />}
        onClick={handleLogout}
        style={{ marginLeft: '16px' }}
        data-testid="logout-btn"
        data-action="logout"
        data-entity="auth"
      >
        退出
      </Button>
    </AntHeader>
  );
};

export default Header;
