import React, { useEffect, useState } from 'react';
import { Table, Button, Tag, Space, Modal, Form, Input, Select, message, Popconfirm, Badge } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, MessageOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { channelApi, Channel, CreateChannelRequest } from '../api/channel';

const { Option } = Select;
const { TextArea } = Input;

const Channels: React.FC = () => {
  const navigate = useNavigate();
  const [channels, setChannels] = useState<Channel[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [editingChannel, setEditingChannel] = useState<Channel | null>(null);
  const [form] = Form.useForm();

  useEffect(() => {
    fetchChannels();
  }, []);

  const fetchChannels = async () => {
    setLoading(true);
    try {
      const res = await channelApi.myChannels();
      if (res.code === 0) {
        setChannels(res.data || []);
      }
    } finally {
      setLoading(false);
    }
  };

  const handleCreate = () => {
    setEditingChannel(null);
    form.resetFields();
    setModalVisible(true);
  };

  const handleEdit = (record: Channel) => {
    setEditingChannel(record);
    form.setFieldsValue({
      name: record.name,
      description: record.description,
      type: record.type,
    });
    setModalVisible(true);
  };

  const handleDelete = async (id: string) => {
    try {
      const res = await channelApi.delete(id);
      if (res.code === 0) {
        message.success('删除成功');
        fetchChannels();
      }
    } catch (error) {
      message.error('删除失败');
    }
  };

  const handleModalOk = async () => {
    try {
      const values = await form.validateFields();
      
      if (editingChannel) {
        const res = await channelApi.update(editingChannel.id, {
          name: values.name,
          description: values.description,
        });
        if (res.code === 0) {
          message.success('更新成功');
        }
      } else {
        const data: CreateChannelRequest = {
          name: values.name,
          description: values.description,
          type: values.type,
        };
        const res = await channelApi.create(data);
        if (res.code === 0) {
          message.success('创建成功');
        }
      }
      setModalVisible(false);
      fetchChannels();
    } catch (error) {
      console.error('表单验证失败:', error);
    }
  };

  const getTypeColor = (type: string) => {
    switch (type) {
      case 'public': return 'green';
      case 'private': return 'orange';
      case 'direct': return 'blue';
      default: return 'default';
    }
  };

  const getTypeText = (type: string) => {
    switch (type) {
      case 'public': return '公开';
      case 'private': return '私有';
      case 'direct': return '私聊';
      default: return type;
    }
  };

  const columns = [
    {
      title: '频道名称',
      dataIndex: 'name',
      key: 'name',
      render: (name: string, record: Channel) => (
        <Space>
          <span>{name}</span>
          {record.unread_count && record.unread_count > 0 && (
            <Badge count={record.unread_count} size="small" />
          )}
        </Space>
      ),
    },
    {
      title: '描述',
      dataIndex: 'description',
      key: 'description',
      ellipsis: true,
    },
    {
      title: '类型',
      dataIndex: 'type',
      key: 'type',
      render: (type: string) => (
        <Tag color={getTypeColor(type)}>{getTypeText(type)}</Tag>
      ),
    },
    {
      title: '成员数',
      dataIndex: 'member_count',
      key: 'member_count',
      render: (count: number) => `${count} 人`,
    },
    {
      title: '创建者',
      dataIndex: 'creator_name',
      key: 'creator_name',
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <Tag color={status === 'active' ? 'success' : 'default'}>
          {status === 'active' ? '活跃' : '归档'}
        </Tag>
      ),
    },
    {
      title: '操作',
      key: 'action',
      render: (_: any, record: Channel) => (
        <Space size="small">
          <Button
            type="primary"
            size="small"
            icon={<MessageOutlined />}
            onClick={() => navigate(`/channels/${record.id}`)}
          >
            进入
          </Button>
          <Button
            type="text"
            icon={<EditOutlined />}
            onClick={() => handleEdit(record)}
          />
          <Popconfirm
            title="确认删除"
            description="确定要删除这个频道吗？"
            onConfirm={() => handleDelete(record.id)}
            okText="确定"
            cancelText="取消"
          >
            <Button type="text" danger icon={<DeleteOutlined />} />
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div>
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <h1 style={{ margin: 0 }}>频道</h1>
        <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
          新建频道
        </Button>
      </div>
      <Table
        columns={columns}
        dataSource={channels}
        rowKey="id"
        loading={loading}
        pagination={{ pageSize: 10 }}
      />

      {/* 创建/编辑弹窗 */}
      <Modal
        title={editingChannel ? '编辑频道' : '新建频道'}
        open={modalVisible}
        onOk={handleModalOk}
        onCancel={() => setModalVisible(false)}
        width={600}
      >
        <Form
          form={form}
          layout="vertical"
          initialValues={{ type: 'public' }}
        >
          <Form.Item
            name="name"
            label="频道名称"
            rules={[{ required: true, message: '请输入频道名称' }]}
          >
            <Input placeholder="请输入频道名称" />
          </Form.Item>
          <Form.Item
            name="description"
            label="描述"
          >
            <TextArea rows={3} placeholder="请输入频道描述" />
          </Form.Item>
          {!editingChannel && (
            <Form.Item
              name="type"
              label="类型"
              rules={[{ required: true }]}
            >
              <Select>
                <Option value="public">公开频道（所有人可见）</Option>
                <Option value="private">私有频道（需邀请加入）</Option>
              </Select>
            </Form.Item>
          )}
        </Form>
      </Modal>
    </div>
  );
};

export default Channels;
