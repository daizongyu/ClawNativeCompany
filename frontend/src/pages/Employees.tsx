import React, { useEffect, useState } from 'react';
import { Table, Button, Tag, Space, Modal, Form, Input, Select, message, Popconfirm, Alert } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, KeyOutlined, SearchOutlined, ReloadOutlined } from '@ant-design/icons';
import { employeeApi, Employee } from '../services/employee';
import { PageContainer } from '../components/common';
import EmployeeForm from '../components/EmployeeForm';

const { Option } = Select;

const Employees: React.FC = () => {
  const [employees, setEmployees] = useState<Employee[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [editingEmployee, setEditingEmployee] = useState<Employee | null>(null);
  const [apiKeyModalVisible, setApiKeyModalVisible] = useState(false);
  const [apiKey, setApiKey] = useState('');
  const [form] = Form.useForm();
  
  // 筛选状态
  const [filterType, setFilterType] = useState<string>('');
  const [filterRole, setFilterRole] = useState<string>('');
  const [filterKeyword, setFilterKeyword] = useState<string>('');
  const [filterStatus, setFilterStatus] = useState<string>('');
  const [roleOptions, setRoleOptions] = useState<string[]>([]);

  // 设置当前页面
  useEffect(() => {
    if (typeof window !== 'undefined' && window.__CLAW_TEST__) {
      window.__CLAW_TEST__.setCurrentPage('employees');
    }
  }, []);

  // 加载职能选项
  useEffect(() => {
    const loadRoleOptions = async () => {
      try {
        const res = await employeeApi.getDistinctRoles();
        if (res.code === 0) {
          setRoleOptions(res.data.roles || []);
        }
      } catch (error) {
        console.error('加载职能选项失败:', error);
      }
    };
    loadRoleOptions();
  }, []);

  // 暴露测试函数
  useEffect(() => {
    if (typeof window !== 'undefined') {
      (window as any).__TEST_EMPLOYEES__ = {
        openModal: () => setModalVisible(true),
        closeModal: () => setModalVisible(false),
        getEmployees: () => employees,
        setEditingEmployee: (emp: Employee | null) => setEditingEmployee(emp),
      };
    }
  }, [employees]);

  const fetchEmployees = async () => {
    setLoading(true);
    try {
      const res = await employeeApi.list({
        type: filterType,
        role: filterRole,
        keyword: filterKeyword,
        status: filterStatus,
      });
      if (res.code === 0) {
        // 后端返回的数据格式是 { list: [...], total: n, page: 1, page_size: 20, total_page: 1 }
        const employeeList = res.data.list || res.data.items || [];
        setEmployees(employeeList);
      }
    } catch (error) {
      console.error('获取员工列表失败:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchEmployees();
  }, [filterType, filterRole, filterStatus]);

  // 搜索处理（防抖）
  useEffect(() => {
    const timer = setTimeout(() => {
      fetchEmployees();
    }, 500);
    return () => clearTimeout(timer);
  }, [filterKeyword]);

  const handleResetFilters = () => {
    setFilterType('');
    setFilterRole('');
    setFilterKeyword('');
    setFilterStatus('');
  };

  const handleCreate = () => {
    setEditingEmployee(null);
    form.resetFields();
    setModalVisible(true);
  };

  const handleEdit = (record: Employee) => {
    setEditingEmployee(record);
    form.setFieldsValue(record);
    setModalVisible(true);
  };

  const handleDelete = async (id: string) => {
    try {
      const res = await employeeApi.delete(id);
      if (res.code === 0) {
        message.success('删除成功');
        fetchEmployees();
      } else {
        message.error(res.message || '删除失败');
      }
    } catch (error) {
      console.error('删除员工失败:', error);
    }
  };

  const handleGenerateApiKey = async (id: string) => {
    try {
      const res = await employeeApi.generateApiKey(id);
      if (res.code === 0) {
        setApiKey(res.data.api_key);
        setApiKeyModalVisible(true);
        message.success('API Key 生成成功');
      } else {
        message.error(res.message || '生成失败');
      }
    } catch (error) {
      console.error('生成 API Key 失败:', error);
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
        <Tag color={type === 'human' ? 'blue' : 'purple'}>
          {type === 'human' ? '人类员工' : 'AI Agent'}
        </Tag>
      ),
    },
    {
      title: '职能',
      dataIndex: 'role',
      key: 'role',
      render: (role: string) => role || '-',
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <Tag color={status === 'active' ? 'green' : 'red'}>
          {status === 'active' ? '在职' : '离职'}
        </Tag>
      ),
    },
    {
      title: '操作',
      key: 'action',
      render: (_: any, record: Employee) => (
        <Space size="middle">
          <Button
            type="primary"
            icon={<EditOutlined />}
            onClick={() => handleEdit(record)}
            data-testid={`employee-edit-btn-${record.id}`}
            data-action="edit"
            data-entity="employee"
          >
            编辑
          </Button>
          <Button
            icon={<KeyOutlined />}
            onClick={() => handleGenerateApiKey(record.id)}
            data-testid={`employee-apikey-btn-${record.id}`}
            data-action="generate-apikey"
            data-entity="employee"
          >
            API Key
          </Button>
          <Popconfirm
            title="确定删除该员工吗？"
            onConfirm={() => handleDelete(record.id)}
            okText="确定"
            cancelText="取消"
          >
            <Button
              danger
              icon={<DeleteOutlined />}
              data-testid={`employee-delete-btn-${record.id}`}
              data-action="delete"
              data-entity="employee"
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
      data-testid="page-employees"
      data-page="employees"
      loading={loading}
    >
      <div style={{ padding: '24px' }}>
        {/* 筛选栏 */}
        <div style={{ marginBottom: '24px', padding: '16px', background: '#f5f5f5', borderRadius: '8px' }}>
          <Space size="middle" wrap>
            <Select
              placeholder="选择类型"
              value={filterType || undefined}
              onChange={setFilterType}
              allowClear
              style={{ width: 120 }}
              data-testid="filter-employee-type"
            >
              <Option value="human">人类员工</Option>
              <Option value="agent">AI Agent</Option>
            </Select>
            <Select
              placeholder="选择状态"
              value={filterStatus || undefined}
              onChange={setFilterStatus}
              allowClear
              style={{ width: 120 }}
              data-testid="filter-employee-status"
            >
              <Option value="active">在职</Option>
              <Option value="inactive">离职</Option>
            </Select>
            <Select
              placeholder="选择职能"
              value={filterRole || undefined}
              onChange={setFilterRole}
              allowClear
              showSearch
              optionFilterProp="children"
              style={{ width: 150 }}
              data-testid="filter-employee-role"
            >
              {roleOptions.map(role => (
                <Option key={role} value={role}>{role}</Option>
              ))}
            </Select>
            <Input
              placeholder="搜索姓名或邮箱"
              value={filterKeyword}
              onChange={(e) => setFilterKeyword(e.target.value)}
              prefix={<SearchOutlined />}
              style={{ width: 200 }}
              data-testid="filter-employee-keyword"
            />
            <Button
              icon={<ReloadOutlined />}
              onClick={handleResetFilters}
              data-testid="filter-reset-btn"
            >
              重置
            </Button>
          </Space>
        </div>

        <div style={{ marginBottom: '16px', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <h1>员工管理</h1>
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={handleCreate}
            data-testid="employee-create-btn"
            data-action="create"
            data-entity="employee"
          >
            新建员工
          </Button>
        </div>

        <Table
          columns={columns}
          dataSource={employees}
          rowKey="id"
          data-testid="employee-table"
          data-entity="employee"
          rowClassName={(record) => `employee-row-${record.id}`}
          onRow={(record) => ({
            'data-testid': `employee-row-${record.id}`,
            'data-employee-id': record.id,
          } as any)}
        />

        {/* 编辑/创建模态框 */}
        <EmployeeForm
          visible={modalVisible}
          onCancel={() => setModalVisible(false)}
          onSuccess={() => {
            setModalVisible(false);
            fetchEmployees();
          }}
          initialValues={editingEmployee || undefined}
          mode={editingEmployee ? 'edit' : 'create'}
        />

        {/* API Key 展示模态框 */}
        <Modal
          title="API Key"
          open={apiKeyModalVisible}
          onOk={() => setApiKeyModalVisible(false)}
          onCancel={() => setApiKeyModalVisible(false)}
          data-testid="apikey-modal"
        >
          <Alert
            message="安全提示"
            description="API Key 是访问系统的重要凭证，请妥善保管。请勿将 API Key 泄露给他人，避免造成数据安全风险。"
            type="warning"
            showIcon
            style={{ marginBottom: '16px' }}
          />
          <p>请复制并保存您的 API Key：</p>
          <Input.TextArea
            value={apiKey}
            readOnly
            rows={3}
            data-testid="apikey-display"
          />
          <p style={{ marginTop: '16px', color: '#666', fontSize: '12px' }}>
            提示：此 API Key 仅显示一次，关闭后将无法再次查看。
          </p>
        </Modal>
      </div>
    </PageContainer>
  );
};

export default Employees;
