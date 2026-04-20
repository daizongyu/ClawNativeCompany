import React, { useEffect, useState } from 'react';
import { Table, Button, Tag, Space, Modal, Form, Input, Select, message, Popconfirm, Badge } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, MessageOutlined, UserAddOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import dayjs from 'dayjs';
import { channelApi, Channel, CreateChannelRequest } from '../services/channel';
import { PageContainer } from '../components/common';

const { Option } = Select;
const { TextArea } = Input;

const Channels: React.FC = () => {
  const navigate = useNavigate();
  const [channels, setChannels] = useState<Channel[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [editingChannel, setEditingChannel] = useState<Channel | null>(null);
  const [addMemberModalVisible, setAddMemberModalVisible] = useState(false);
  const [addingMemberChannel, setAddingMemberChannel] = useState<Channel | null>(null);
  const [addMemberLoading, setAddMemberLoading] = useState(false);
  const [form] = Form.useForm();
  const [addMemberForm] = Form.useForm();

  // 筛选状态
  const [filterType, setFilterType] = useState<string>('');
  const [filterKeyword, setFilterKeyword] = useState<string>('');

  // 设置当前页面
  useEffect(() => {
    if (typeof window !== 'undefined' && window.__CLAW_TEST__) {
      window.__CLAW_TEST__.setCurrentPage('channels');
    }
  }, []);

  // 暴露测试函数
  useEffect(() => {
    if (typeof window !== 'undefined') {
      (window as any).__TEST_CHANNELS__ = {
        openModal: () => setModalVisible(true),
        closeModal: () => setModalVisible(false),
        getChannels: () => channels,
        setEditingChannel: (ch: Channel | null) => setEditingChannel(ch),
      };
    }
  }, [channels]);

  // 筛选变化时重新加载
  useEffect(() => {
    fetchChannels();
  }, [filterType]);

  // 关键词搜索防抖
  useEffect(() => {
    const timer = setTimeout(() => {
      fetchChannels();
    }, 500);
    return () => clearTimeout(timer);
  }, [filterKeyword]);

  const handleResetFilters = () => {
    setFilterType('');
    setFilterKeyword('');
  };

  const fetchChannels = async () => {
    setLoading(true);
    try {
      const params: any = {};
      if (filterType) params.type = filterType;
      if (filterKeyword) params.keyword = filterKeyword;
      
      const res = await channelApi.list(params);
      if (res.code === 0) {
        // 后端返回的数据格式是 { list: [...], total: n, page: 1, page_size: 20, total_page: 1 }
        const channelList = res.data.list || res.data.items || [];
        setChannels(channelList);
      }
    } catch (error) {
      console.error('获取频道列表失败:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchChannels();
  }, []);

  const handleCreate = () => {
    setEditingChannel(null);
    form.resetFields();
    setModalVisible(true);
  };

  const handleEdit = (record: Channel) => {
    setEditingChannel(record);
    form.setFieldsValue({
      name: record.name,
      type: record.type,
      description: record.description,
    });
    setModalVisible(true);
  };

  const handleDelete = async (id: string) => {
    try {
      const res = await channelApi.delete(id);
      if (res.code === 0) {
        message.success('删除成功');
        fetchChannels();
      } else {
        message.error(res.message || '删除失败');
      }
    } catch (error) {
      console.error('删除频道失败:', error);
    }
  };

  const handleEnterChat = (id: string) => {
    navigate(`/channels/${id}`);
  };

  const handleAddMember = (channel: Channel) => {
    setAddingMemberChannel(channel);
    addMemberForm.resetFields();
    setAddMemberModalVisible(true);
  };

  const handleAddMemberCancel = () => {
    setAddMemberModalVisible(false);
    setAddingMemberChannel(null);
  };

  const handleAddMemberSubmit = async () => {
    try {
      const values = await addMemberForm.validateFields();
      if (!addingMemberChannel) return;
      
      setAddMemberLoading(true);
      const res = await channelApi.addMember(addingMemberChannel.id, values.employee_id, values.role);
      if (res.code === 0) {
        message.success('添加成员成功');
        setAddMemberModalVisible(false);
        fetchChannels();
      } else {
        message.error(res.message || '添加成员失败');
      }
    } catch (error) {
      console.error('添加成员失败:', error);
    } finally {
      setAddMemberLoading(false);
    }
  };

  const handleModalOk = async () => {
    try {
      const values = await form.validateFields();
      if (editingChannel) {
        const res = await channelApi.update(editingChannel.id, values);
        if (res.code === 0) {
          message.success('更新成功');
          setModalVisible(false);
          fetchChannels();
        } else {
          message.error(res.message || '更新失败');
        }
      } else {
        const res = await channelApi.create(values as CreateChannelRequest);
        if (res.code === 0) {
          message.success('创建成功');
          setModalVisible(false);
          fetchChannels();
        } else {
          message.error(res.message || '创建失败');
        }
      }
    } catch (error) {
      console.error('保存频道失败:', error);
    }
  };

  const columns = [
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: '类型',
      dataIndex: 'type',
      key: 'type',
      render: (type: string) => {
        const typeMap: Record<string, { label: string; color: string }> = {
          public: { label: '公开频道', color: 'green' },
          private: { label: '私有频道', color: 'orange' },
          dm: { label: '私信', color: 'blue' },
        };
        const config = typeMap[type] || { label: type, color: 'default' };
        return <Tag color={config.color}>{config.label}</Tag>;
      },
    },
    {
      title: '成员数',
      dataIndex: 'member_count',
      key: 'member_count',
      render: (count: number) => <Badge count={count} showZero color="#108ee9" />,
    },
    {
      title: '描述',
      dataIndex: 'description',
      key: 'description',
      ellipsis: true,
    },
    {
      title: '创建者',
      dataIndex: 'creator_name',
      key: 'creator_name',
      render: (name: string) => name || '-',
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (date: string) => date ? dayjs(date).format('YYYY-MM-DD HH:mm') : '-',
    },
    {
      title: '操作',
      key: 'action',
      render: (_: any, record: Channel) => (
        <Space size="middle">
          <Button
            type="primary"
            icon={<MessageOutlined />}
            onClick={() => handleEnterChat(record.id)}
            data-testid={`channel-chat-btn-${record.id}`}
            data-action="chat"
            data-entity="channel"
          >
            进入
          </Button>
          <Button
            icon={<UserAddOutlined />}
            onClick={() => handleAddMember(record)}
            data-testid={`channel-add-member-btn-${record.id}`}
            data-action="add-member"
            data-entity="channel"
          >
            添加成员
          </Button>
          <Button
            icon={<EditOutlined />}
            onClick={() => handleEdit(record)}
            data-testid={`channel-edit-btn-${record.id}`}
            data-action="edit"
            data-entity="channel"
          >
            编辑
          </Button>
          <Popconfirm
            title="确定删除该频道吗？"
            onConfirm={() => handleDelete(record.id)}
            okText="确定"
            cancelText="取消"
          >
            <Button
              danger
              icon={<DeleteOutlined />}
              data-testid={`channel-delete-btn-${record.id}`}
              data-action="delete"
              data-entity="channel"
            >
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <PageContainer
      data-testid="page-channels"
      data-page="channels"
      loading={loading}
    >
      <div style={{ padding: '24px' }}>
        <div style={{ marginBottom: '16px', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <h1>频道管理</h1>
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={handleCreate}
            data-testid="channel-create-btn"
            data-action="create"
            data-entity="channel"
          >
            新建频道
          </Button>
        </div>

        {/* 筛选栏 */}
        <div style={{ marginBottom: '16px', display: 'flex', gap: '16px', alignItems: 'center' }}>
          <Select
            placeholder="筛选类型"
            allowClear
            value={filterType || undefined}
            onChange={(value) => setFilterType(value || '')}
            style={{ width: 150 }}
            data-testid="channel-filter-type"
          >
            <Option value="public">公开频道</Option>
            <Option value="private">私有频道</Option>
            <Option value="dm">私信</Option>
          </Select>
          <Input
            placeholder="搜索频道名称或描述"
            value={filterKeyword}
            onChange={(e) => setFilterKeyword(e.target.value)}
            style={{ width: 250 }}
            data-testid="channel-filter-keyword"
          />
          <Button onClick={handleResetFilters} data-testid="channel-filter-reset">
            重置
          </Button>
        </div>

        <Table
          columns={columns}
          dataSource={channels}
          rowKey="id"
          data-testid="channel-table"
          data-entity="channel"
          onRow={(record) => ({
            'data-testid': `channel-row-${record.id}`,
            'data-channel-id': record.id,
          } as any)}
        />

        <Modal
          title={editingChannel ? '编辑频道' : '新建频道'}
          open={modalVisible}
          onOk={handleModalOk}
          onCancel={() => setModalVisible(false)}
          destroyOnClose
          width={600}
          data-testid="channel-modal"
        >
          <Form form={form} layout="vertical">
            <Form.Item
              label="名称"
              name="name"
              rules={[{ required: true, message: '请输入频道名称' }]}
            >
              <Input
                placeholder="请输入频道名称"
                data-testid="input-channel-name"
                data-input-name="channel-name"
              />
            </Form.Item>
            <Form.Item
              label="类型"
              name="type"
              rules={[{ required: true, message: '请选择频道类型' }]}
            >
              <Select
                placeholder="请选择频道类型"
                data-testid="input-channel-type"
                data-input-name="channel-type"
              >
                <Option value="public">公开</Option>
                <Option value="private">私有</Option>
              </Select>
            </Form.Item>
            <Form.Item
              label="描述"
              name="description"
            >
              <TextArea
                rows={4}
                placeholder="请输入频道描述"
                data-testid="input-channel-description"
                data-input-name="channel-description"
              />
            </Form.Item>
          </Form>
        </Modal>

        {/* 添加成员弹窗 */}
        <Modal
          title={`添加成员 - ${addingMemberChannel?.name || ''}`}
          open={addMemberModalVisible}
          onOk={handleAddMemberSubmit}
          onCancel={handleAddMemberCancel}
          confirmLoading={addMemberLoading}
          destroyOnClose
          width={500}
          data-testid="add-member-modal"
        >
          <Form form={addMemberForm} layout="vertical">
            <Form.Item
              label="员工ID"
              name="employee_id"
              rules={[{ required: true, message: '请输入员工ID' }]}
            >
              <Input
                placeholder="请输入员工ID"
                data-testid="input-add-member-employee-id"
                data-input-name="employee-id"
              />
            </Form.Item>
            <Form.Item
              label="角色"
              name="role"
              initialValue="member"
              rules={[{ required: true, message: '请选择角色' }]}
            >
              <Select
                placeholder="请选择角色"
                data-testid="select-add-member-role"
                data-input-name="member-role"
              >
                <Option value="member">成员</Option>
                <Option value="admin">管理员</Option>
                <Option value="readonly">只读</Option>
              </Select>
            </Form.Item>
          </Form>
        </Modal>
      </div>
    </PageContainer>
  );
};

export default Channels;
