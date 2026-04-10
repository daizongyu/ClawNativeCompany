import React, { useEffect, useState, useRef } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Input, Button, Space, Tag, message, Empty, Modal, List, Avatar } from 'antd';
import { SendOutlined, ArrowLeftOutlined, UserOutlined, InfoCircleOutlined, PlusOutlined } from '@ant-design/icons';
import { channelApi, Channel, ChannelMember } from '../api/channel';
import { messageApi, Message } from '../api/message';
import { useAuthStore } from '../stores/auth';
import dayjs from 'dayjs';
import { PageContainer } from '../components/common';

const { TextArea } = Input;

const ChannelChat: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { user } = useAuthStore();
  const messagesEndRef = useRef<HTMLDivElement>(null);

  const [channel, setChannel] = useState<Channel | null>(null);
  const [messages, setMessages] = useState<Message[]>([]);
  const [members, setMembers] = useState<ChannelMember[]>([]);
  const [loading, setLoading] = useState(false);
  const [sending, setSending] = useState(false);
  const [inputValue, setInputValue] = useState('');
  const [infoModalVisible, setInfoModalVisible] = useState(false);
  const [addMemberModalVisible, setAddMemberModalVisible] = useState(false);
  const [addingMember, setAddingMember] = useState(false);

  // 设置当前页面
  useEffect(() => {
    if (typeof window !== 'undefined' && window.__CLAW_TEST__) {
      window.__CLAW_TEST__.setCurrentPage('channel-chat');
    }
  }, []);

  // 暴露测试函数
  useEffect(() => {
    if (typeof window !== 'undefined') {
      (window as any).__TEST_CHANNEL_CHAT__ = {
        getMessages: () => messages,
        getChannel: () => channel,
        sendMessage: (content: string) => handleSend(content),
      };
    }
  }, [messages, channel]);

  const fetchChannel = async () => {
    if (!id) return;
    try {
      const res = await channelApi.getById(id);
      if (res.code === 0) {
        setChannel(res.data);
      }
    } catch (error) {
      console.error('获取频道信息失败:', error);
    }
  };

  const fetchMessages = async () => {
    if (!id) return;
    setLoading(true);
    try {
      const res = await messageApi.listByChannel(id);
      if (res.code === 0) {
        // 支持两种返回格式: list 或 items
        const messageList = res.data?.list || res.data?.items || [];
        setMessages(messageList);
      }
    } catch (error) {
      console.error('获取消息失败:', error);
    } finally {
      setLoading(false);
    }
  };

  const fetchMembers = async () => {
    if (!id) return;
    try {
      const res = await channelApi.getMembers(id);
      if (res.code === 0) {
        setMembers(res.data);
      }
    } catch (error) {
      console.error('获取成员列表失败:', error);
    }
  };

  useEffect(() => {
    if (id) {
      fetchChannel();
      fetchMessages();
      fetchMembers();
    }
  }, [id]);

  // 自动滚动到底部
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  const handleSend = async (content?: string) => {
    const text = content || inputValue;
    if (!text.trim() || !id) return;

    setSending(true);
    try {
      const res = await messageApi.send({
        channel_id: id,
        content: text.trim(),
        content_type: 'text',
      });
      if (res.code === 0) {
        setInputValue('');
        fetchMessages();
      } else {
        message.error(res.message || '发送失败');
      }
    } catch (error) {
      console.error('发送消息失败:', error);
    } finally {
      setSending(false);
    }
  };

  const handleGoBack = () => {
    navigate('/channels');
  };

  const handleShowInfo = () => {
    setInfoModalVisible(true);
  };

  const handleCloseInfo = () => {
    setInfoModalVisible(false);
  };

  const handleShowAddMember = () => {
    setAddMemberModalVisible(true);
  };

  const handleCloseAddMember = () => {
    setAddMemberModalVisible(false);
  };

  const handleAddMember = async (employeeId: string, role: string = 'member') => {
    if (!id) return;
    setAddingMember(true);
    try {
      const res = await channelApi.addMember(id, employeeId, role);
      if (res.code === 0) {
        message.success('添加成员成功');
        fetchMembers();
        handleCloseAddMember();
      } else {
        message.error(res.message || '添加成员失败');
      }
    } catch (error) {
      console.error('添加成员失败:', error);
      message.error('添加成员失败');
    } finally {
      setAddingMember(false);
    }
  };

  const isOwnMessage = (msg: Message) => {
    return msg.sender_id === user?.id;
  };

  // 获取发送者名称（优先使用消息中的 sender 对象）
  const getSenderName = (msg: Message) => {
    // 优先使用消息中的 sender 对象
    if (msg.sender?.name) {
      return msg.sender.name;
    }
    // 兼容旧格式
    if (msg.sender_name) {
      return msg.sender_name;
    }
    // 从成员列表查找
    const member = members.find((m) => m.employee_id === msg.sender_id);
    return member?.employee_name || '未知用户';
  };

  // 解析 @提及并高亮显示
  const renderMessageContent = (content: string, mentions?: string[]) => {
    if (!mentions || mentions.length === 0) {
      return content;
    }

    // 将 @提及替换为高亮样式
    let result = content;
    mentions.forEach((mention) => {
      // 支持 @用户名 或 @员工ID 格式
      const regex = new RegExp(`@(${mention}|[^\\s]+)`, 'g');
      result = result.replace(regex, (match) => {
        return `<span style="color: #1890ff; background: #e6f7ff; padding: 0 4px; border-radius: 4px;">${match}</span>`;
      });
    });

    return <span dangerouslySetInnerHTML={{ __html: result }} />;
  };

  return (
    <PageContainer
      data-testid="page-channel-chat"
      data-page="channel-chat"
      loading={loading}
      style={{ height: '100vh', display: 'flex', flexDirection: 'column' }}
    >
      {/* 头部 */}
      <div
        style={{
          padding: '16px 24px',
          borderBottom: '1px solid #e8e8e8',
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          background: '#fff',
        }}
        data-testid="channel-chat-header"
      >
        <Space>
          <Button
            icon={<ArrowLeftOutlined />}
            onClick={handleGoBack}
            data-testid="channel-chat-back-btn"
            data-action="back"
            data-entity="channel"
          >
            返回
          </Button>
          <div>
            <h2 style={{ margin: 0 }}>{channel?.name || '频道聊天'}</h2>
            <Tag color={channel?.type === 'public' ? 'green' : 'orange'}>
              {channel?.type === 'public' ? '公开' : '私有'}
            </Tag>
          </div>
        </Space>
        <Space>
          <Tag icon={<UserOutlined />}>
            {members.length} 成员
          </Tag>
          <Button
            icon={<InfoCircleOutlined />}
            onClick={handleShowInfo}
            data-testid="channel-chat-info-btn"
            data-action="info"
            data-entity="channel"
          >
            详情
          </Button>
        </Space>
      </div>

      {/* 消息列表 */}
      <div
        style={{
          flex: 1,
          overflow: 'auto',
          padding: '24px',
          background: '#f5f5f5',
        }}
        data-testid="channel-chat-messages"
      >
        {messages.length === 0 ? (
          <Empty description="暂无消息" />
        ) : (
          <Space direction="vertical" style={{ width: '100%' }} size="large">
            {messages.map((msg) => (
              <div
                key={msg.id}
                style={{
                  display: 'flex',
                  justifyContent: isOwnMessage(msg) ? 'flex-end' : 'flex-start',
                  width: '100%',
                }}
                data-testid={`message-${msg.id}`}
                data-message-id={msg.id}
                data-sender-id={msg.sender_id}
              >
                <div
                  style={{
                    maxWidth: '70%',
                    padding: '12px 16px',
                    borderRadius: '12px',
                    background: isOwnMessage(msg) ? '#1890ff' : '#fff',
                    color: isOwnMessage(msg) ? '#fff' : '#333',
                    boxShadow: '0 2px 8px rgba(0,0,0,0.1)',
                  }}
                >
                  <div style={{ marginBottom: '4px', fontSize: '12px', opacity: 0.8 }}>
                    {getSenderName(msg)} · {dayjs(msg.created_at).format('HH:mm')}
                  </div>
                  <div style={{ wordBreak: 'break-word' }}>
                    {renderMessageContent(msg.content, msg.mentions)}
                  </div>
                </div>
              </div>
            ))}
            <div ref={messagesEndRef} />
          </Space>
        )}
      </div>

      {/* 输入框 */}
      <div
        style={{
          padding: '16px 24px',
          borderTop: '1px solid #e8e8e8',
          background: '#fff',
        }}
        data-testid="channel-chat-input-area"
      >
        <Space style={{ width: '100%' }} size="middle">
          <TextArea
            value={inputValue}
            onChange={(e) => setInputValue(e.target.value)}
            placeholder="输入消息..."
            autoSize={{ minRows: 1, maxRows: 4 }}
            style={{ flex: 1 }}
            onPressEnter={(e) => {
              if (!e.shiftKey) {
                e.preventDefault();
                handleSend();
              }
            }}
            data-testid="input-message-content"
            data-input-name="message-content"
          />
          <Button
            type="primary"
            icon={<SendOutlined />}
            loading={sending}
            onClick={() => handleSend()}
            data-testid="message-send-btn"
            data-action="send"
            data-entity="message"
          >
            发送
          </Button>
        </Space>
      </div>

      {/* 频道详情弹窗 */}
      <Modal
        title="频道详情"
        open={infoModalVisible}
        onCancel={handleCloseInfo}
        footer={null}
        width={600}
        data-testid="channel-info-modal"
      >
        {channel && (
          <div style={{ padding: '16px 0' }}>
            <div style={{ marginBottom: '24px' }}>
              <h3 style={{ marginBottom: '8px' }}>{channel.name}</h3>
              <Tag color={channel.type === 'public' ? 'green' : 'orange'}>
                {channel.type === 'public' ? '公开频道' : '私有频道'}
              </Tag>
              <p style={{ marginTop: '12px', color: '#666' }}>
                {channel.description || '暂无描述'}
              </p>
              <div style={{ marginTop: '12px', fontSize: '12px', color: '#999' }}>
                创建者: {channel.creator_name || channel.created_by} | 
                创建时间: {dayjs(channel.created_at).format('YYYY-MM-DD HH:mm')}
              </div>
            </div>

            <div>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px' }}>
                <h4 style={{ margin: 0 }}>成员列表 ({members.length})</h4>
                <Button
                  type="primary"
                  size="small"
                  icon={<PlusOutlined />}
                  onClick={handleShowAddMember}
                  data-testid="channel-add-member-btn"
                  data-action="add-member"
                  data-entity="channel"
                >
                  添加成员
                </Button>
              </div>
              <List
                dataSource={members}
                renderItem={(member) => (
                  <List.Item key={member.employee_id}>
                    <List.Item.Meta
                      avatar={<Avatar>{member.employee?.name?.charAt(0) || '?'}</Avatar>}
                      title={member.employee?.name || member.employee_id}
                      description={
                        <Space>
                          <Tag color={member.role === 'admin' ? 'red' : 'blue'}>
                            {member.role === 'admin' ? '管理员' : member.role === 'readonly' ? '只读' : '成员'}
                          </Tag>
                          {member.employee?.email}
                        </Space>
                      }
                    />
                  </List.Item>
                )}
              />
            </div>
          </div>
        )}
      </Modal>

      {/* 添加成员弹窗 */}
      <Modal
        title="添加成员"
        open={addMemberModalVisible}
        onCancel={handleCloseAddMember}
        footer={null}
        width={500}
        data-testid="add-member-modal"
      >
        <AddMemberForm
          onSubmit={handleAddMember}
          onCancel={handleCloseAddMember}
          loading={addingMember}
        />
      </Modal>
    </PageContainer>
  );
};

// 添加成员表单组件
interface AddMemberFormProps {
  onSubmit: (employeeId: string, role: string) => void;
  onCancel: () => void;
  loading: boolean;
}

const AddMemberForm: React.FC<AddMemberFormProps> = ({ onSubmit, onCancel, loading }) => {
  const [employeeId, setEmployeeId] = useState('');
  const [role, setRole] = useState('member');

  const handleSubmit = () => {
    if (!employeeId.trim()) {
      message.error('请输入员工ID');
      return;
    }
    onSubmit(employeeId.trim(), role);
  };

  return (
    <div style={{ padding: '16px 0' }}>
      <div style={{ marginBottom: '16px' }}>
        <label style={{ display: 'block', marginBottom: '8px' }}>员工ID：</label>
        <Input
          value={employeeId}
          onChange={(e) => setEmployeeId(e.target.value)}
          placeholder="请输入员工ID"
          data-testid="add-member-employee-id"
        />
      </div>
      <div style={{ marginBottom: '24px' }}>
        <label style={{ display: 'block', marginBottom: '8px' }}>角色：</label>
        <select
          value={role}
          onChange={(e) => setRole(e.target.value)}
          style={{ width: '100%', padding: '8px', borderRadius: '4px', border: '1px solid #d9d9d9' }}
          data-testid="add-member-role"
        >
          <option value="member">成员</option>
          <option value="admin">管理员</option>
          <option value="readonly">只读</option>
        </select>
      </div>
      <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '8px' }}>
        <Button onClick={onCancel} data-testid="add-member-cancel">取消</Button>
        <Button type="primary" loading={loading} onClick={handleSubmit} data-testid="add-member-submit">
          确定
        </Button>
      </div>
    </div>
  );
};

export default ChannelChat;
