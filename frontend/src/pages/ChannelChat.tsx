import React, { useEffect, useState, useRef } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Input, Button, Avatar, Space, Tag, message, Spin, Empty, Popconfirm } from 'antd';
import { SendOutlined, ArrowLeftOutlined, UserOutlined, DeleteOutlined, InfoCircleOutlined } from '@ant-design/icons';
import { channelApi, Channel, ChannelMember } from '../api/channel';
import { messageApi, Message } from '../api/message';
import { useAuthStore } from '../stores/auth';
import dayjs from 'dayjs';

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

  useEffect(() => {
    if (id) {
      fetchChannelInfo();
      fetchMessages();
      fetchMembers();
    }
  }, [id]);

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const fetchChannelInfo = async () => {
    if (!id) return;
    try {
      const res = await channelApi.getById(id);
      if (res.code === 0) {
        setChannel(res.data);
      }
    } catch (error) {
      message.error('获取频道信息失败');
    }
  };

  const fetchMessages = async () => {
    if (!id) return;
    setLoading(true);
    try {
      const res = await messageApi.listByChannel(id, 1, 100);
      if (res.code === 0) {
        setMessages(res.data.list || []);
      }
    } finally {
      setLoading(false);
    }
  };

  const fetchMembers = async () => {
    if (!id) return;
    try {
      const res = await channelApi.getMembers(id);
      if (res.code === 0) {
        setMembers(res.data || []);
      }
    } catch (error) {
      console.error('获取成员列表失败:', error);
    }
  };

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  const handleSend = async () => {
    if (!id || !inputValue.trim()) return;
    
    setSending(true);
    try {
      const res = await messageApi.send({
        channel_id: id,
        content: inputValue.trim(),
        content_type: 'text',
      });
      if (res.code === 0) {
        setInputValue('');
        fetchMessages();
      }
    } catch (error) {
      message.error('发送失败');
    } finally {
      setSending(false);
    }
  };

  const handleRecall = async (messageId: string) => {
    try {
      const res = await messageApi.recall(messageId);
      if (res.code === 0) {
        message.success('撤回成功');
        fetchMessages();
      }
    } catch (error) {
      message.error('撤回失败');
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  const isMyMessage = (msg: Message) => msg.sender_id === user?.id;

  const formatTime = (time: string) => {
    return dayjs(time).format('HH:mm');
  };

  const formatDate = (time: string) => {
    return dayjs(time).format('YYYY-MM-DD');
  };

  // 按日期分组消息
  const groupedMessages = messages.reduce((groups: { [key: string]: Message[] }, msg) => {
    const date = formatDate(msg.created_at);
    if (!groups[date]) {
      groups[date] = [];
    }
    groups[date].push(msg);
    return groups;
  }, {});

  if (loading && messages.length === 0) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100%' }}>
        <Spin size="large" />
      </div>
    );
  }

  return (
    <div style={{ height: 'calc(100vh - 112px)', display: 'flex', flexDirection: 'column' }}>
      {/* 头部 */}
      <div style={{ padding: '16px 0', borderBottom: '1px solid #f0f0f0', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <Space>
          <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/channels')}>
            返回
          </Button>
          <h2 style={{ margin: 0 }}>{channel?.name || '频道聊天'}</h2>
          {channel?.type && (
            <Tag color={channel.type === 'public' ? 'green' : channel.type === 'private' ? 'orange' : 'blue'}>
              {channel.type === 'public' ? '公开' : channel.type === 'private' ? '私有' : '私聊'}
            </Tag>
          )}
        </Space>
        <Space>
          <span style={{ color: '#666' }}>{members.length} 人在线</span>
          <Button icon={<InfoCircleOutlined />} type="text" />
        </Space>
      </div>

      {/* 消息区域 */}
      <div style={{ flex: 1, overflow: 'auto', padding: '16px 0' }}>
        {messages.length === 0 ? (
          <Empty description="暂无消息，发送第一条消息吧" />
        ) : (
          <div>
            {Object.entries(groupedMessages).map(([date, msgs]) => (
              <div key={date}>
                <div style={{ textAlign: 'center', margin: '16px 0' }}>
                  <Tag color="default">{date}</Tag>
                </div>
                {msgs.map((msg) => (
                  <div
                    key={msg.id}
                    style={{
                      display: 'flex',
                      justifyContent: isMyMessage(msg) ? 'flex-end' : 'flex-start',
                      marginBottom: '16px',
                      padding: '0 8px',
                    }}
                  >
                    <div
                      style={{
                        display: 'flex',
                        flexDirection: isMyMessage(msg) ? 'row-reverse' : 'row',
                        alignItems: 'flex-start',
                        maxWidth: '70%',
                      }}
                    >
                      <Avatar
                        icon={<UserOutlined />}
                        style={{
                          backgroundColor: isMyMessage(msg) ? '#1890ff' : '#52c41a',
                          margin: isMyMessage(msg) ? '0 0 0 8px' : '0 8px 0 0',
                        }}
                      />
                      <div>
                        <div
                          style={{
                            display: 'flex',
                            alignItems: 'center',
                            marginBottom: '4px',
                            justifyContent: isMyMessage(msg) ? 'flex-end' : 'flex-start',
                          }}
                        >
                          <span style={{ fontWeight: 'bold', marginRight: '8px' }}>
                            {msg.sender_name}
                          </span>
                          <span style={{ fontSize: '12px', color: '#999' }}>
                            {formatTime(msg.created_at)}
                          </span>
                          {msg.sender_type === 'agent' && (
                            <Tag color="blue" style={{ marginLeft: '4px' }}>
                              Agent
                            </Tag>
                          )}
                        </div>
                        <div
                          style={{
                            backgroundColor: isMyMessage(msg) ? '#1890ff' : '#f0f0f0',
                            color: isMyMessage(msg) ? '#fff' : '#000',
                            padding: '8px 12px',
                            borderRadius: '8px',
                            wordBreak: 'break-word',
                            position: 'relative',
                          }}
                        >
                          {msg.content}
                          {isMyMessage(msg) && (
                            <Popconfirm
                              title="撤回消息"
                              description="确定要撤回这条消息吗？"
                              onConfirm={() => handleRecall(msg.id)}
                              okText="确定"
                              cancelText="取消"
                            >
                              <Button
                                type="text"
                                size="small"
                                icon={<DeleteOutlined />}
                                style={{
                                  position: 'absolute',
                                  top: '-20px',
                                  right: '0',
                                  color: '#999',
                                  opacity: 0,
                                  transition: 'opacity 0.2s',
                                }}
                                className="message-recall-btn"
                              />
                            </Popconfirm>
                          )}
                        </div>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            ))}
            <div ref={messagesEndRef} />
          </div>
        )}
      </div>

      {/* 输入区域 */}
      <div style={{ padding: '16px 0', borderTop: '1px solid #f0f0f0' }}>
        <Space.Compact style={{ width: '100%' }}>
          <TextArea
            value={inputValue}
            onChange={(e) => setInputValue(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="输入消息... (Enter 发送, Shift+Enter 换行)"
            autoSize={{ minRows: 1, maxRows: 4 }}
            style={{ flex: 1 }}
          />
          <Button
            type="primary"
            icon={<SendOutlined />}
            onClick={handleSend}
            loading={sending}
            disabled={!inputValue.trim()}
          >
            发送
          </Button>
        </Space.Compact>
      </div>

      <style>{`
        .message-recall-btn:hover {
          opacity: 1 !important;
        }
      `}</style>
    </div>
  );
};

export default ChannelChat;
