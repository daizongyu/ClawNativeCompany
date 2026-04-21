import React, { useState, useEffect } from 'react';
import { Card, Tabs, Typography, message, Spin } from 'antd';
import { UserOutlined, BellOutlined, SettingOutlined } from '@ant-design/icons';
import UserCard from '../../components/UserCard';
import NotificationPrefsForm from '../../components/notification/NotificationPrefsForm';
import GatewayConfigPanel from '../../components/gateway/GatewayConfigPanel';
import EmployeeForm from '../../components/EmployeeForm';
import { employeeApi, Employee } from '../../services/employee';
import { useAuthStore } from '../../stores/auth';

const { Title } = Typography;

const Profile: React.FC = () => {
  const { user } = useAuthStore();
  const [employee, setEmployee] = useState<Employee | null>(null);
  const [loading, setLoading] = useState(true);
  const [editModalVisible, setEditModalVisible] = useState(false);
  const [activeTab, setActiveTab] = useState('profile');

  const fetchEmployee = async () => {
    if (!user?.id) return;
    setLoading(true);
    try {
      const res = await employeeApi.getById(user.id);
      setEmployee(res.data);
    } catch (err) {
      message.error('加载用户信息失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchEmployee();
  }, [user?.id]);

  const handleEditSuccess = () => {
    setEditModalVisible(false);
    fetchEmployee();
    message.success('资料已更新');
  };

  if (loading || !employee) {
    return (
      <div style={{ textAlign: 'center', padding: 50 }}>
        <Spin size="large" />
      </div>
    );
  }

  const tabItems = [
    {
      key: 'profile',
      label: (
        <span>
          <UserOutlined />
          个人资料
        </span>
      ),
      children: (
        <UserCard
          employee={employee}
          onEdit={() => setEditModalVisible(true)}
          onNotificationSettings={() => setActiveTab('notifications')}
          onGatewaySettings={() => setActiveTab('gateway')}
        />
      ),
    },
    {
      key: 'notifications',
      label: (
        <span>
          <BellOutlined />
          通知偏好
        </span>
      ),
      children: <NotificationPrefsForm employeeId={employee.id} />,
    },
    {
      key: 'gateway',
      label: (
        <span>
          <SettingOutlined />
          推送配置
        </span>
      ),
      children: <GatewayConfigPanel employeeId={employee.id} />,
    },
  ];

  return (
    <div className="profile-page" data-testid="profile-page">
      <Card>
        <Title level={4} style={{ marginBottom: 24 }}>
          个人中心
        </Title>

        <Tabs
          activeKey={activeTab}
          onChange={setActiveTab}
          items={tabItems}
          data-testid="profile-tabs"
        />
      </Card>

      {/* 编辑资料弹窗 */}
      <EmployeeForm
        visible={editModalVisible}
        onCancel={() => setEditModalVisible(false)}
        onSuccess={handleEditSuccess}
        initialValues={employee}
        mode="edit"
      />
    </div>
  );
};

export default Profile;
