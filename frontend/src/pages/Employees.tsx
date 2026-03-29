import React, { useEffect, useState } from 'react';
import { Table, Button, Tag, Space, Modal, Form, Input, Select, message, Popconfirm } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, KeyOutlined } from '@ant-design/icons';
import { employeeApi, Employee, CreateEmployeeRequest, UpdateEmployeeRequest } from '../api/employee';

const { Option } = Select;

const Employees: React.FC = () => {
  const [employees, setEmployees] = useState<Employee[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [editingEmployee, setEditingEmployee] = useState<Employee | null>(null);
  const [apiKeyModalVisible, setApiKeyModalVisible] = useState(false);
  const [apiKey, setApiKey] = useState('');
  const [form] = Form.useForm();

  useEffect(() => {
    fetchEmployees();
  }, []);

  const fetchEmployees = async () => {
    setLoading(true);
    try {
      const res = await employeeApi.list(1, 100);
      if (res.code === 0) {
        setEmployees(res.data.list || []);
      }
    } finally {
      setLoading(false);
    }
  };

  const handleCreate = () => {
    setEditingEmployee(null);
    form.resetFields();
    setModalVisible(true);
  };

  const handleEdit = (record: Employee) => {
    setEditingEmployee(record);
    form.setFieldsValue({
      name: record.name,
      email: record.email,
      type: record.type,
      role: record.role,
      skills: record.skills?.join(', '),
    });
    setModalVisible(true);
  };

  const handleDelete = async (id: string) => {
    try {
      const res = await employeeApi.delete(id);
      if (res.code === 0) {
        message.success('删除成功');
        fetchEmployees();
      }
    } catch (error) {
      message.error('删除失败');
    }
  };

  const handleGenerateApiKey = async (id: string) => {
    try {
      const res = await employeeApi.generateApiKey(id);
      if (res.code === 0) {
        setApiKey(res.data.api_key);
        setApiKeyModalVisible(true);
      }
    } catch (error) {
      message.error('生成 API Key 失败');
    }
  };

  const handleModalOk = async () => {
    try {
      const values = await form.validateFields();
      const skills = values.skills ? values.skills.split(',').map((s: string) => s.trim()).filter(Boolean) : [];
      
      if (editingEmployee) {
        // 更新
        const data: UpdateEmployeeRequest = {
          name: values.name,
          email: values.email,
          role: values.role,
          skills,
        };
        const res = await employeeApi.update(editingEmployee.id, data);
        if (res.code === 0) {
          message.success('更新成功');
        }
      } else {
        // 创建
        const data: CreateEmployeeRequest = {
          name: values.name,
          email: values.email,
          password: values.password,
          type: values.type,
          role: values.role,
          skills,
        };
        const res = await employeeApi.create(data);
        if (res.code === 0) {
          message.success('创建成功');
        }
      }
      setModalVisible(false);
      fetchEmployees();
    } catch (error) {
      console.error('表单验证失败:', error);
    }
  };

  const columns = [
    {
      title: '姓名',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: '邮箱',
      dataIndex: 'email',
      key: 'email',
    },
    {
      title: '类型',
      dataIndex: 'type',
      key: 'type',
      render: (type: string) => (
        <Tag color={type === 'human' ? 'blue' : 'green'}>
          {type === 'human' ? '人类' : 'Agent'}
        </Tag>
      ),
    },
    {
      title: '角色',
      dataIndex: 'role',
      key: 'role',
    },
    {
      title: '技能',
      dataIndex: 'skills',
      key: 'skills',
      render: (skills: string[]) => (
        <Space size="small">
          {skills?.map(skill => (
            <Tag key={skill}>{skill}</Tag>
          ))}
        </Space>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <Tag color={status === 'active' ? 'success' : 'default'}>
          {status === 'active' ? '活跃' : '停用'}
        </Tag>
      ),
    },
    {
      title: '操作',
      key: 'action',
      render: (_: any, record: Employee) => (
        <Space size="small">
          <Button
            type="text"
            icon={<EditOutlined />}
            onClick={() => handleEdit(record)}
          />
          <Button
            type="text"
            icon={<KeyOutlined />}
            onClick={() => handleGenerateApiKey(record.id)}
          />
          <Popconfirm
            title="确认删除"
            description="确定要删除这个员工吗？"
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
        <h1 style={{ margin: 0 }}>员工管理</h1>
        <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
          新建员工
        </Button>
      </div>
      <Table
        columns={columns}
        dataSource={employees}
        rowKey="id"
        loading={loading}
        pagination={{ pageSize: 10 }}
      />

      {/* 创建/编辑弹窗 */}
      <Modal
        title={editingEmployee ? '编辑员工' : '新建员工'}
        open={modalVisible}
        onOk={handleModalOk}
        onCancel={() => setModalVisible(false)}
        width={600}
      >
        <Form
          form={form}
          layout="vertical"
          initialValues={{ type: 'human' }}
        >
          <Form.Item
            name="name"
            label="姓名"
            rules={[{ required: true, message: '请输入姓名' }]}
          >
            <Input placeholder="请输入姓名" />
          </Form.Item>
          <Form.Item
            name="email"
            label="邮箱"
            rules={[
              { required: true, message: '请输入邮箱' },
              { type: 'email', message: '请输入有效的邮箱' },
            ]}
          >
            <Input placeholder="请输入邮箱" />
          </Form.Item>
          {!editingEmployee && (
            <Form.Item
              name="password"
              label="密码"
              rules={[{ required: true, message: '请输入密码' }]}
            >
              <Input.Password placeholder="请输入密码" />
            </Form.Item>
          )}
          <Form.Item
            name="type"
            label="类型"
            rules={[{ required: true }]}
          >
            <Select disabled={!!editingEmployee}>
              <Option value="human">人类</Option>
              <Option value="agent">Agent</Option>
            </Select>
          </Form.Item>
          <Form.Item
            name="role"
            label="角色"
          >
            <Input placeholder="请输入角色，如：开发工程师" />
          </Form.Item>
          <Form.Item
            name="skills"
            label="技能"
            extra="多个技能用逗号分隔"
          >
            <Input placeholder="如：Python, React, 数据分析" />
          </Form.Item>
        </Form>
      </Modal>

      {/* API Key 弹窗 */}
      <Modal
        title="API Key"
        open={apiKeyModalVisible}
        onOk={() => setApiKeyModalVisible(false)}
        onCancel={() => setApiKeyModalVisible(false)}
        footer={[
          <Button key="ok" type="primary" onClick={() => setApiKeyModalVisible(false)}>
            确定
          </Button>,
        ]}
      >
        <p>请妥善保存以下 API Key，它只会显示一次：</p>
        <Input.TextArea
          value={apiKey}
          readOnly
          rows={4}
          style={{ fontFamily: 'monospace' }}
        />
      </Modal>
    </div>
  );
};

export default Employees;
