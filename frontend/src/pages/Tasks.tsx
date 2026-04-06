import React, { useEffect, useState } from 'react';
import { Table, Button, Tag, Space, Modal, Form, Input, Select, DatePicker, message, Row, Col, Card, Statistic, Popconfirm } from 'antd';
import { PlusOutlined, SearchOutlined, ReloadOutlined, CheckCircleOutlined, DeleteOutlined } from '@ant-design/icons';
import dayjs from 'dayjs';
import { taskApi } from '../api/task';
import { PageContainer } from '../components/common';

const { TextArea } = Input;
const { Option } = Select;

interface Task {
  id: string;
  title: string;
  description: string;
  status: string;
  priority: string;
  assignee_id?: string;
  assignee_name?: string;
  creator_id: string;
  creator_name: string;
  due_date?: string;
  created_at: string;
  updated_at: string;
}

const Tasks: React.FC = () => {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [filteredTasks, setFilteredTasks] = useState<Task[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [detailModalVisible, setDetailModalVisible] = useState(false);
  const [editingTask, setEditingTask] = useState<Task | null>(null);
  const [viewingTask, setViewingTask] = useState<Task | null>(null);
  const [searchKeyword, setSearchKeyword] = useState('');
  const [statusFilter, setStatusFilter] = useState<string>('');
  const [priorityFilter, setPriorityFilter] = useState<string>('');
  const [form] = Form.useForm();

  // 设置当前页面
  useEffect(() => {
    if (typeof window !== 'undefined' && window.__CLAW_TEST__) {
      window.__CLAW_TEST__.setCurrentPage('tasks');
    }
  }, []);

  // 暴露测试函数
  useEffect(() => {
    if (typeof window !== 'undefined') {
      (window as any).__TEST_TASKS__ = {
        openModal: () => setModalVisible(true),
        closeModal: () => setModalVisible(false),
        getTasks: () => tasks,
        setEditingTask: (task: Task | null) => setEditingTask(task),
      };
    }
  }, [tasks]);

  const fetchTasks = async () => {
    setLoading(true);
    try {
      const res = await taskApi.list(1, 100);
      if (res.code === 0) {
        // 后端返回的数据格式是 { list: [...], total: n, page: 1, page_size: 20, total_page: 1 }
        const taskList = res.data.list || res.data.items || [];
        setTasks(taskList);
        setFilteredTasks(taskList);
      }
    } catch (error) {
      console.error('获取任务列表失败:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchTasks();
  }, []);

  // 过滤任务
  useEffect(() => {
    let result = tasks;
    if (searchKeyword) {
      result = result.filter(
        (task) =>
          task.title.toLowerCase().includes(searchKeyword.toLowerCase()) ||
          task.description.toLowerCase().includes(searchKeyword.toLowerCase())
      );
    }
    if (statusFilter) {
      result = result.filter((task) => task.status === statusFilter);
    }
    if (priorityFilter) {
      result = result.filter((task) => task.priority === priorityFilter);
    }
    setFilteredTasks(result);
  }, [tasks, searchKeyword, statusFilter, priorityFilter]);

  const handleCreate = () => {
    setEditingTask(null);
    form.resetFields();
    setModalVisible(true);
  };

  const handleEdit = (record: Task) => {
    setEditingTask(record);
    form.setFieldsValue({
      ...record,
      due_date: record.due_date ? dayjs(record.due_date) : null,
    });
    setModalVisible(true);
  };

  const handleView = (record: Task) => {
    setViewingTask(record);
    setDetailModalVisible(true);
  };

  const handleDelete = async (id: string) => {
    try {
      const res = await taskApi.delete(id);
      if (res.code === 0) {
        message.success('删除成功');
        fetchTasks();
      } else {
        message.error(res.message || '删除失败');
      }
    } catch (error) {
      console.error('删除任务失败:', error);
    }
  };

  const handleComplete = async (id: string) => {
    try {
      const res = await taskApi.complete(id, {});
      if (res.code === 0) {
        message.success('任务已完成');
        fetchTasks();
      } else {
        message.error(res.message || '操作失败');
      }
    } catch (error) {
      console.error('完成任务失败:', error);
    }
  };

  const handleModalOk = async () => {
    try {
      const values = await form.validateFields();
      const taskData = {
        ...values,
        due_date: values.due_date ? values.due_date.format('YYYY-MM-DD') : null,
      };

      if (editingTask) {
        const res = await taskApi.update(editingTask.id, taskData);
        if (res.code === 0) {
          message.success('更新成功');
          setModalVisible(false);
          fetchTasks();
        } else {
          message.error(res.message || '更新失败');
        }
      } else {
        const res = await taskApi.create(taskData);
        if (res.code === 0) {
          message.success('创建成功');
          setModalVisible(false);
          fetchTasks();
        } else {
          message.error(res.message || '创建失败');
        }
      }
    } catch (error) {
      console.error('保存任务失败:', error);
    }
  };

  const handleResetFilters = () => {
    setSearchKeyword('');
    setStatusFilter('');
    setPriorityFilter('');
  };

  const getStatusColor = (status: string) => {
    const colors: Record<string, string> = {
      pending: 'default',
      in_progress: 'processing',
      completed: 'success',
      cancelled: 'error',
    };
    return colors[status] || 'default';
  };

  const getStatusText = (status: string) => {
    const texts: Record<string, string> = {
      pending: '待处理',
      in_progress: '进行中',
      completed: '已完成',
      cancelled: '已取消',
    };
    return texts[status] || status;
  };

  const getPriorityColor = (priority: string) => {
    const colors: Record<string, string> = {
      low: 'success',
      medium: 'warning',
      high: 'error',
    };
    return colors[priority] || 'default';
  };

  const getPriorityText = (priority: string) => {
    const texts: Record<string, string> = {
      low: '低',
      medium: '中',
      high: '高',
    };
    return texts[priority] || priority;
  };

  const columns = [
    {
      title: '标题',
      dataIndex: 'title',
      key: 'title',
      render: (text: string, record: Task) => (
        <Button
          type="link"
          onClick={() => handleView(record)}
          style={{ padding: 0 }}
          data-testid={`task-title-${record.id}`}
        >
          {text}
        </Button>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <Tag color={getStatusColor(status)}>{getStatusText(status)}</Tag>
      ),
    },
    {
      title: '优先级',
      dataIndex: 'priority',
      key: 'priority',
      render: (priority: string) => (
        <Tag color={getPriorityColor(priority)}>{getPriorityText(priority)}</Tag>
      ),
    },
    {
      title: '负责人',
      key: 'assignee',
      render: (_: any, record: Task) => {
        // 优先显示指派人，如果没有则显示创建人
        return record.assignee_name || record.creator_name || '未分配';
      },
    },
    {
      title: '截止日期',
      dataIndex: 'due_date',
      key: 'due_date',
      render: (date: string) => (date ? dayjs(date).format('YYYY-MM-DD') : '无'),
    },
    {
      title: '操作',
      key: 'action',
      render: (_: any, record: Task) => (
        <Space size="middle">
          <Button
            type="primary"
            icon={<CheckCircleOutlined />}
            onClick={() => handleComplete(record.id)}
            disabled={record.status === 'completed'}
            data-testid={`task-complete-btn-${record.id}`}
            data-action="complete"
            data-entity="task"
          >
            完成
          </Button>
          <Button
            icon={<ReloadOutlined />}
            onClick={() => handleEdit(record)}
            data-testid={`task-edit-btn-${record.id}`}
            data-action="edit"
            data-entity="task"
          >
            编辑
          </Button>
          <Popconfirm
            title="确定删除该任务吗？"
            onConfirm={() => handleDelete(record.id)}
            okText="确定"
            cancelText="取消"
          >
            <Button
              danger
              icon={<DeleteOutlined />}
              data-testid={`task-delete-btn-${record.id}`}
              data-action="delete"
              data-entity="task"
            >
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  // 统计任务数量
  const pendingCount = tasks.filter((t) => t.status === 'pending').length;
  const inProgressCount = tasks.filter((t) => t.status === 'in_progress').length;
  const completedCount = tasks.filter((t) => t.status === 'completed').length;

  return (
    <PageContainer
      data-testid="page-tasks"
      data-page="tasks"
      loading={loading}
    >
      <div style={{ padding: '24px' }}>
        <Row gutter={[16, 16]} style={{ marginBottom: '24px' }}>
          <Col xs={24} sm={8}>
            <Card data-testid="task-stat-pending">
              <Statistic title="待处理" value={pendingCount} valueStyle={{ color: '#cf1322' }} />
            </Card>
          </Col>
          <Col xs={24} sm={8}>
            <Card data-testid="task-stat-inprogress">
              <Statistic title="进行中" value={inProgressCount} valueStyle={{ color: '#1890ff' }} />
            </Card>
          </Col>
          <Col xs={24} sm={8}>
            <Card data-testid="task-stat-completed">
              <Statistic title="已完成" value={completedCount} valueStyle={{ color: '#3f8600' }} />
            </Card>
          </Col>
        </Row>

        <div style={{ marginBottom: '16px', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <h1>任务管理</h1>
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={handleCreate}
            data-testid="task-create-btn"
            data-action="create"
            data-entity="task"
          >
            新建任务
          </Button>
        </div>

        <Card style={{ marginBottom: '16px' }} data-testid="task-filter-card">
          <Space wrap>
            <Input
              placeholder="搜索任务"
              value={searchKeyword}
              onChange={(e) => setSearchKeyword(e.target.value)}
              prefix={<SearchOutlined />}
              style={{ width: 200 }}
              data-testid="input-task-search"
              data-input-name="task-search"
            />
            <Select
              placeholder="状态筛选"
              value={statusFilter || undefined}
              onChange={setStatusFilter}
              style={{ width: 120 }}
              allowClear
              data-testid="input-task-status-filter"
              data-input-name="task-status-filter"
            >
              <Option value="pending">待处理</Option>
              <Option value="in_progress">进行中</Option>
              <Option value="completed">已完成</Option>
              <Option value="cancelled">已取消</Option>
            </Select>
            <Select
              placeholder="优先级筛选"
              value={priorityFilter || undefined}
              onChange={setPriorityFilter}
              style={{ width: 120 }}
              allowClear
              data-testid="input-task-priority-filter"
              data-input-name="task-priority-filter"
            >
              <Option value="high">高</Option>
              <Option value="medium">中</Option>
              <Option value="low">低</Option>
            </Select>
            <Button
              icon={<ReloadOutlined />}
              onClick={handleResetFilters}
              data-testid="task-reset-filter-btn"
              data-action="reset-filter"
              data-entity="task"
            >
              重置
            </Button>
          </Space>
        </Card>

        <Table
          columns={columns}
          dataSource={filteredTasks}
          rowKey="id"
          data-testid="task-table"
          data-entity="task"
          onRow={(record) => ({
            'data-testid': `task-row-${record.id}`,
            'data-task-id': record.id,
          } as any)}
        />

        {/* 编辑/创建模态框 */}
        <Modal
          title={editingTask ? '编辑任务' : '新建任务'}
          open={modalVisible}
          onOk={handleModalOk}
          onCancel={() => setModalVisible(false)}
          destroyOnClose
          width={700}
          data-testid="task-modal"
        >
          <Form form={form} layout="vertical">
            <Form.Item
              label="标题"
              name="title"
              rules={[{ required: true, message: '请输入任务标题' }]}
            >
              <Input
                placeholder="请输入任务标题"
                data-testid="input-task-title"
                data-input-name="task-title"
              />
            </Form.Item>
            <Form.Item
              label="描述"
              name="description"
            >
              <TextArea
                rows={4}
                placeholder="请输入任务描述"
                data-testid="input-task-description"
                data-input-name="task-description"
              />
            </Form.Item>
            <Form.Item
              label="状态"
              name="status"
              rules={[{ required: true, message: '请选择状态' }]}
            >
              <Select
                placeholder="请选择状态"
                data-testid="input-task-status"
                data-input-name="task-status"
              >
                <Option value="pending">待处理</Option>
                <Option value="in_progress">进行中</Option>
                <Option value="completed">已完成</Option>
                <Option value="cancelled">已取消</Option>
              </Select>
            </Form.Item>
            <Form.Item
              label="优先级"
              name="priority"
              rules={[{ required: true, message: '请选择优先级' }]}
            >
              <Select
                placeholder="请选择优先级"
                data-testid="input-task-priority"
                data-input-name="task-priority"
              >
                <Option value="high">高</Option>
                <Option value="medium">中</Option>
                <Option value="low">低</Option>
              </Select>
            </Form.Item>
            <Form.Item
              label="截止日期"
              name="due_date"
            >
              <DatePicker
                style={{ width: '100%' }}
                data-testid="input-task-due-date"
                data-input-name="task-due-date"
              />
            </Form.Item>
          </Form>
        </Modal>

        {/* 详情模态框 */}
        <Modal
          title="任务详情"
          open={detailModalVisible}
          onOk={() => setDetailModalVisible(false)}
          onCancel={() => setDetailModalVisible(false)}
          footer={[
            <Button
              key="close"
              onClick={() => setDetailModalVisible(false)}
              data-testid="task-detail-close-btn"
            >
              关闭
            </Button>,
          ]}
          data-testid="task-detail-modal"
        >
          {viewingTask && (
            <div>
              <p><strong>标题：</strong>{viewingTask.title}</p>
              <p><strong>描述：</strong>{viewingTask.description || '无'}</p>
              <p>
                <strong>状态：</strong>
                <Tag color={getStatusColor(viewingTask.status)}>
                  {getStatusText(viewingTask.status)}
                </Tag>
              </p>
              <p>
                <strong>优先级：</strong>
                <Tag color={getPriorityColor(viewingTask.priority)}>
                  {getPriorityText(viewingTask.priority)}
                </Tag>
              </p>
              <p><strong>负责人：</strong>{viewingTask.assignee_name || '未分配'}</p>
              <p><strong>创建人：</strong>{viewingTask.creator_name}</p>
              <p><strong>截止日期：</strong>{viewingTask.due_date || '无'}</p>
              <p><strong>创建时间：</strong>{dayjs(viewingTask.created_at).format('YYYY-MM-DD HH:mm')}</p>
            </div>
          )}
        </Modal>
      </div>
    </PageContainer>
  );
};

export default Tasks;
