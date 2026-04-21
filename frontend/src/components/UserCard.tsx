import React from 'react';
import { Card, Avatar, Typography, Tag, Descriptions, Button, Space } from 'antd';
import {
  UserOutlined,
  MailOutlined,
  PhoneOutlined,
  TeamOutlined,
  IdcardOutlined,
  EditOutlined,
  SettingOutlined,
  BellOutlined,
} from '@ant-design/icons';
import { Employee } from '../services/employee';

interface UserCardProps {
  employee: Employee;
  onEdit?: () => void;
  onNotificationSettings?: () => void;
  onGatewaySettings?: () => void;
  showActions?: boolean;
}

const UserCard: React.FC<UserCardProps> = ({
  employee,
  onEdit,
  onNotificationSettings,
  onGatewaySettings,
  showActions = true,
}) => {
  const getTypeTag = (type: string) => {
    switch (type) {
      case 'human':
        return <Tag color="blue">人类</Tag>;
      case 'agent':
        return <Tag color="purple">Agent</Tag>;
      default:
        return <Tag>{type}</Tag>;
    }
  };

  const getStatusTag = (status: string) => {
    switch (status) {
      case 'active':
        return <Tag color="success">活跃</Tag>;
      case 'inactive':
        return <Tag>停用</Tag>;
      default:
        return <Tag>{status}</Tag>;
    }
  };

  return (
    <Card
      className="user-card"
      data-testid={`user-card-${employee.id}`}
      actions={
        showActions
          ? [
              <Button
                key="edit"
                type="text"
                icon={<EditOutlined />}
                onClick={onEdit}
                data-testid="user-card-edit-btn"
              >
                编辑资料
              </Button>,
              <Button
                key="notification"
                type="text"
                icon={<BellOutlined />}
                onClick={onNotificationSettings}
                data-testid="user-card-notification-btn"
              >
                通知设置
              </Button>,
              <Button
                key="gateway"
                type="text"
                icon={<SettingOutlined />}
                onClick={onGatewaySettings}
                data-testid="user-card-gateway-btn"
              >
                推送配置
              </Button>,
            ]
          : undefined
      }
    >
      <Card.Meta
        avatar={
          <Avatar
            size={64}
            src={employee.avatar}
            icon={<UserOutlined />}
            data-testid="user-card-avatar"
          />
        }
        title={
          <Space>
            <Typography.Text strong style={{ fontSize: 18 }}>
              {employee.display_name || employee.name}
            </Typography.Text>
            {getTypeTag(employee.type)}
            {getStatusTag(employee.status)}
          </Space>
        }
        description={
          <Space direction="vertical" size={0}>
            <Typography.Text type="secondary">
              @{employee.username}
            </Typography.Text>
            {employee.role && (
              <Typography.Text type="secondary">
                {employee.role}
              </Typography.Text>
            )}
          </Space>
        }
      />

      <Descriptions
        column={1}
        size="small"
        style={{ marginTop: 24 }}
        items={[
          {
            key: 'email',
            label: <><MailOutlined /> 邮箱</>,
            children: employee.email,
          },
          ...(employee.phone
            ? [
                {
                  key: 'phone',
                  label: <><PhoneOutlined /> 电话</>,
                  children: employee.phone,
                },
              ]
            : []),
          ...(employee.department
            ? [
                {
                  key: 'department',
                  label: <><TeamOutlined /> 部门</>,
                  children: employee.department,
                },
              ]
            : []),
          ...(employee.position
            ? [
                {
                  key: 'position',
                  label: <><IdcardOutlined /> 职位</>,
                  children: employee.position,
                },
              ]
            : []),
          ...(employee.skills && employee.skills.length > 0
            ? [
                {
                  key: 'skills',
                  label: '技能',
                  children: (
                    <Space size={[0, 4]} wrap>
                      {employee.skills.map((skill) => (
                        <Tag key={skill}>
                          {skill}
                        </Tag>
                      ))}
                    </Space>
                  ),
                },
              ]
            : []),
        ]}
      />
    </Card>
  );
};

export default UserCard;
