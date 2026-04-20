import React, { useState, useEffect } from 'react';
import {
  Card,
  Typography,
  Switch,
  Divider,
  Row,
  Col,
  Tag,
  message,
} from 'antd';
import {
  MailOutlined,
  ApiOutlined,
  BellOutlined,
} from '@ant-design/icons';
import { employeeApi, NotificationPreferences } from '../../services/employee';

interface NotificationPrefsFormProps {
  employeeId: string;
}

const NotificationPrefsForm: React.FC<NotificationPrefsFormProps> = ({ employeeId }) => {
  const [prefs, setPrefs] = useState<NotificationPreferences>({
    channels: {
      email: false,
      webhook: false,
      internal: true,
    },
    events: {
      task_assigned: true,
      task_completed: true,
      task_cancelled: true,
      workflow_triggered: true,
      workflow_completed: true,
      workflow_failed: true,
      mention_received: true,
      channel_message: false,
    },
  });
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);

  // 加载当前偏好设置
  useEffect(() => {
    const loadPrefs = async () => {
      setLoading(true);
      try {
        const res = await employeeApi.getById(employeeId);
        if (res.data?.notification_prefs) {
          setPrefs(res.data.notification_prefs);
        }
      } catch (err) {
        console.error('加载通知偏好失败:', err);
      } finally {
        setLoading(false);
      }
    };
    loadPrefs();
  }, [employeeId]);

  // 保存偏好设置
  const savePrefs = async (newPrefs: NotificationPreferences) => {
    setSaving(true);
    try {
      await employeeApi.updateNotificationPrefs(employeeId, {
        channels: newPrefs.channels,
        events: newPrefs.events,
      });
      message.success('通知偏好已保存');
    } catch (err) {
      message.error('保存通知偏好失败');
      console.error('保存通知偏好失败:', err);
    } finally {
      setSaving(false);
    }
  };

  // 更新渠道设置
  const handleChannelChange = (channel: keyof typeof prefs.channels, checked: boolean) => {
    const newPrefs = {
      ...prefs,
      channels: {
        ...prefs.channels,
        [channel]: checked,
      },
    };
    setPrefs(newPrefs);
    savePrefs(newPrefs);
  };

  // 更新事件设置
  const handleEventChange = (event: keyof typeof prefs.events, checked: boolean) => {
    const newPrefs = {
      ...prefs,
      events: {
        ...prefs.events,
        [event]: checked,
      },
    };
    setPrefs(newPrefs);
    savePrefs(newPrefs);
  };

  // 事件标签映射
  const eventLabels: Record<keyof typeof prefs.events, string> = {
    task_assigned: '任务指派',
    task_completed: '任务完成',
    task_cancelled: '任务取消',
    workflow_triggered: '工作流触发',
    workflow_completed: '工作流完成',
    workflow_failed: '工作流失败',
    mention_received: '收到@提及',
    channel_message: '频道消息',
  };

  return (
    <div data-testid="notification-prefs-form">
      {/* 通知渠道 */}
      <Card title="通知渠道" style={{ marginBottom: 16 }} loading={loading}>
        <Typography.Text type="secondary" style={{ display: 'block', marginBottom: 16 }}>
          选择您希望接收通知的方式
        </Typography.Text>

        <Row gutter={[16, 16]}>
          <Col xs={24} sm={8}>
            <Card
              bordered
              style={{
                textAlign: 'center',
                opacity: prefs.channels.email ? 1 : 0.6,
                borderColor: prefs.channels.email ? '#1890ff' : undefined,
              }}
            >
              <MailOutlined style={{ fontSize: 24, color: prefs.channels.email ? '#1890ff' : '#999', marginBottom: 8 }} />
              <div>邮件通知</div>
              <Switch
                checked={prefs.channels.email}
                onChange={(checked) => handleChannelChange('email', checked)}
                size="small"
                style={{ marginTop: 8 }}
                data-testid="notification-channel-email"
              />
            </Card>
          </Col>

          <Col xs={24} sm={8}>
            <Card
              bordered
              style={{
                textAlign: 'center',
                opacity: prefs.channels.webhook ? 1 : 0.6,
                borderColor: prefs.channels.webhook ? '#1890ff' : undefined,
              }}
            >
              <ApiOutlined style={{ fontSize: 24, color: prefs.channels.webhook ? '#1890ff' : '#999', marginBottom: 8 }} />
              <div>Webhook</div>
              <Switch
                checked={prefs.channels.webhook}
                onChange={(checked) => handleChannelChange('webhook', checked)}
                size="small"
                style={{ marginTop: 8 }}
                data-testid="notification-channel-webhook"
              />
            </Card>
          </Col>

          <Col xs={24} sm={8}>
            <Card
              bordered
              style={{
                textAlign: 'center',
                opacity: prefs.channels.internal ? 1 : 0.6,
                borderColor: prefs.channels.internal ? '#1890ff' : undefined,
              }}
            >
              <BellOutlined style={{ fontSize: 24, color: prefs.channels.internal ? '#1890ff' : '#999', marginBottom: 8 }} />
              <div>站内通知</div>
              <Switch
                checked={prefs.channels.internal}
                onChange={(checked) => handleChannelChange('internal', checked)}
                size="small"
                style={{ marginTop: 8 }}
                data-testid="notification-channel-internal"
              />
            </Card>
          </Col>
        </Row>
      </Card>

      {/* 通知事件 */}
      <Card title="通知事件" loading={loading}>
        <Typography.Text type="secondary" style={{ display: 'block', marginBottom: 16 }}>
          选择您希望接收哪些类型的事件通知
        </Typography.Text>

        <div style={{ marginTop: 16 }}>
          {/* 任务相关 */}
          <Typography.Text type="secondary" style={{ display: 'block', marginBottom: 8 }}>
            任务通知
          </Typography.Text>
          <Row gutter={[8, 8]}>
            {(['task_assigned', 'task_completed', 'task_cancelled'] as const).map((event) => (
              <Col key={event}>
                <Tag
                  icon={<BellOutlined />}
                  color={prefs.events[event] ? 'blue' : 'default'}
                  onClick={() => handleEventChange(event, !prefs.events[event])}
                  style={{ cursor: 'pointer' }}
                  data-testid={`notification-event-${event}`}
                >
                  {eventLabels[event]}
                </Tag>
              </Col>
            ))}
          </Row>

          <Divider />

          {/* 工作流相关 */}
          <Typography.Text type="secondary" style={{ display: 'block', marginBottom: 8 }}>
            工作流通知
          </Typography.Text>
          <Row gutter={[8, 8]}>
            {(['workflow_triggered', 'workflow_completed', 'workflow_failed'] as const).map((event) => (
              <Col key={event}>
                <Tag
                  icon={<BellOutlined />}
                  color={prefs.events[event] ? 'blue' : 'default'}
                  onClick={() => handleEventChange(event, !prefs.events[event])}
                  style={{ cursor: 'pointer' }}
                  data-testid={`notification-event-${event}`}
                >
                  {eventLabels[event]}
                </Tag>
              </Col>
            ))}
          </Row>

          <Divider />

          {/* 消息相关 */}
          <Typography.Text type="secondary" style={{ display: 'block', marginBottom: 8 }}>
            消息通知
          </Typography.Text>
          <Row gutter={[8, 8]}>
            {(['mention_received', 'channel_message'] as const).map((event) => (
              <Col key={event}>
                <Tag
                  icon={<BellOutlined />}
                  color={prefs.events[event] ? 'blue' : 'default'}
                  onClick={() => handleEventChange(event, !prefs.events[event])}
                  style={{ cursor: 'pointer' }}
                  data-testid={`notification-event-${event}`}
                >
                  {eventLabels[event]}
                </Tag>
              </Col>
            ))}
          </Row>
        </div>
      </Card>

      {saving && (
        <Typography.Text type="secondary" style={{ marginTop: 8, display: 'block' }}>
          保存中...
        </Typography.Text>
      )}
    </div>
  );
};

export default NotificationPrefsForm;
