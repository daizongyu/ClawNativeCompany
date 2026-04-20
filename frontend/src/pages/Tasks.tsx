import React, { useEffect, useState } from 'react';
import {
  Table, Button, Tag, Space, Modal, Form, Input, Select, DatePicker, message,
  Row, Col, Card, Statistic, Popconfirm, Radio, Tabs, Badge
} from 'antd';
import {
  PlusOutlined, SearchOutlined, ReloadOutlined, CheckCircleOutlined,
  DeleteOutlined, InboxOutlined, TeamOutlined, StopOutlined
} from '@ant-design/icons';
import dayjs from 'dayjs';
import { taskApi } from '../services/task';
import { employeeApi } from '../services/employee';
import { PageContainer } from '../components/common';
import { useAuthStore } from '../stores/auth';

const { TextArea } = Input;
const { Option } = Select;
const { TabPane } = Tabs;

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

interface Employee {
  id: string;
  name: string;
  email: string;
}

// 指派模式类型
type AssignmentMode = 'assign' | 'claim';

const Tasks: React.FC = () => {
  const { user } = useAuthStore();
  const [tasks, setTasks] = useState<Task[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [detailModalVisible, setDetailModalVisible] = useState(false);
  const [editingTask, setEditingTask] = useState<Task | null>(null);
  const [viewingTask, setViewingTask] = useState<Task | null>(null);
  const [form] = Form.useForm();
  const [employees, setEmployees] = useState<Employee[]>([]);

  // 筛选状态
  const [activeTab, setActiveTab] = useState('all');
  const [statusFilter, setStatusFilter] = useState('');
  const [priorityFilter, setPriorityFilter] = useState('');
  const [keywordFilter, setKeywordFilter] = useState('');

  // 指派模式
  const [assignmentMode, setAssignmentMode] = useState<AssignmentMode>('assign');

  useEffect(() => {
    fetchTasks();
    fetchEmployees();
  }, [activeTab, statusFilter, priorityFilter, keywordFilter]);

  const fetchTasks = async () => {
    setLoading(true);
    try {
      const params: any = {
        page: 1,
        page_size: 100,
      };

      // 根据当前标签页设置筛选条件
      if (activeTab === 'mine') {
        params.mine = true;
      } else if (activeTab === 'unclaimed') {
        params.unclaimed = true;
      }

      // 其他筛选条件
      if (statusFilter) params.status = statusFilter;
      if (priorityFilter) params.priority = priorityFilter;
      if (keywordFilter) params.keyword = keywordFilter;

      const res = await taskApi.list(params);
      if (res.code === 0) {
        const taskList = res.data?.list || res.data?.items || [];
        setTasks(taskList);
      }
    } catch (error) {
      console.error('获取任务列表失败:', error);
      message.error('获取任务列表失败');
    } finally {
      setLoading(false);
    }
  };

  const fetchEmployees = async () => {
    try {
      const res = await employeeApi.list({ page: 1, pageSize: 100 });
      if (res.code === 0) {
        const list = res.data?.list || res.data?.items || [];
        setEmployees(list);
      }
    } catch (error) {
      console.error('获取员工列表失败:', error);
    }
  };

  const handleCreate = () => {
    setEditingTask(null);
    form.resetFields();
    setAssignmentMode('assign');
    setModalVisible(true);
  };

  const handleEdit = (record: Task) => {
    // 已完成的任务不能编辑
    if (record.status === 'completed') {
      message.warning('已完成的任务不能编辑');
      return;
    }
    setEditingTask(record);
    form.setFieldsValue({
      ...record,
      due_date: record.due_date ? dayjs(record.due_date) : null,
    });
    setAssignmentMode(record.assignee_id ? 'assign' : 'claim');
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
      message.error('删除失败');
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

  const handleCancel = async (id: string) => {
    try {
      const res = await taskApi.cancel(id);
      if (res.code === 0) {
        message.success('任务已取消');
        fetchTasks();
      } else {
        message.error(res.message || '取消失败');
      }
    } catch (error) {
      console.error('取消任务失败:', error);
      message.error('取消失败');
    }
  };

  const handleClaim = async (taskId: string) => {
    try {
      const res = await taskApi.claim(taskId);
      if (res.code === 0) {
        message.success('任务认领成功');
        fetchTasks();
      } else {
        message.error(res.message || '认领失败');
      }
    } catch (error) {
      console.error('认领任务失败:', error);
    }
  };

  const handleModalOk = async () => {
    try {
      const values = await form.validateFields();
      const taskData: any = {
        title: values.title,
        description: values.description,
        priority: values.priority,
      };

      if (values.due_date) {
        // 使用带时间的格式
        taskData.due_date = values.due_date.format('YYYY-MM-DDTHH:mm:ss');
      }

      // 根据指派模式处理 assignee_id
      if (assignmentMode === 'assign' && values.assignee_id) {
        taskData.assignee_id = values.assignee_id;
      }
      // claim 模式不传递 assignee_id，后端会处理为待认领状态

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

  const getStatusTag = (status: string) => {
    const statusMap: Record<string, { color: string; text: string }> = {
      pending: { color: 'default', text: '待处理' },
      in_progress: { color: 'processing', text: '进行中' },
      completed: { color: 'success', text: '已完成' },
      cancelled: { color: 'error', text: '已取消' },
    };
    const { color, text } = statusMap[status] || { color: 'default', text: status };
    return <Tag color={color}>{text}</Tag>;
  };

  const getPriorityTag = (priority: string) => {
    const priorityMap: Record<string, { color: string; text: string }> = {
      low: { color: 'default', text: '低' },
      medium: { color: 'blue', text: '中' },
      high: { color: 'orange', text: '高' },
      urgent: { color: 'red', text: '紧急' },
    };
    const { color, text } = priorityMap[priority] || { color: 'default', text: priority };
    return <Tag color={color}>{text}</Tag>;
  };

  const columns = [
    {
      title: '任务标题',
      dataIndex: 'title',
      key: 'title',
      render: (text: string, record: Task) => (
        <a onClick={() => handleView(record)}>{text}</a>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => getStatusTag(status),
    },
    {
      title: '优先级',
      dataIndex: 'priority',
      key: 'priority',
      render: (priority: string) => getPriorityTag(priority),
    },
    {
      title: '执行人',
      dataIndex: 'assignee_name',
      key: 'assignee_name',
      render: (name: string) => {
        if (!name || name === '') {
          return <Tag color="warning">待认领</Tag>;
        }
        return name;
      },
    },
    {
      title: '创建人',
      dataIndex: 'creator_name',
      key: 'creator_name',
    },
    {
      title: '截止时间',
      dataIndex: 'due_date',
      key: 'due_date',
      render: (date: string) => date ? dayjs(date).format('YYYY-MM-DD HH:mm') : '-',
    },
    {
      title: '操作',
      key: 'action',
      render: (_: any, taskRecord: Task) => (
        <Space size="small">
          <Button type="link" size="small" onClick={() => handleView(taskRecord)}>
            查看
          </Button>
          {taskRecord.status !== 'completed' && (
            <Button type="link" size="small" onClick={() => handleEdit(taskRecord)}>
              编辑
            </Button>
          )}
          {!taskRecord.assignee_id && taskRecord.status !== 'cancelled' && (
            <Button
              type="link"
              size="small"
              onClick={() => handleClaim(taskRecord.id)}
              data-testid={`task-claim-btn-${taskRecord.id}`}
            >
              认领
            </Button>
          )}
          {taskRecord.status !== 'completed' && taskRecord.status !== 'cancelled' && taskRecord.assignee_id === user?.id && (
            <Button
              type="link"
              size="small"
              icon={<CheckCircleOutlined />}
              onClick={() => handleComplete(taskRecord.id)}
            >
              完成
            </Button>
          )}
          {taskRecord.status !== 'completed' && taskRecord.status !== 'cancelled' && (
            <Popconfirm
              title="确认取消"
              description="确定要取消这个任务吗？"
              onConfirm={() => handleCancel(taskRecord.id)}
              okText="确认"
              cancelText="取消"
            >
              <Button type="link" size="small" icon={<StopOutlined />}>
                取消
              </Button>
            </Popconfirm>
          )}
          <Popconfirm
            title="确认删除"
            description="确定要删除这个任务吗？"
            onConfirm={() => handleDelete(taskRecord.id)}
            okText="确认"
            cancelText="取消"
          >
            <Button type="link" danger size="small" icon={<DeleteOutlined />}>
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  // 统计任务数量
  const myTasksCount = tasks.filter(t => t.assignee_id === user?.id).length;
  const unclaimedCount = tasks.filter(t => !t.assignee_id && t.status !== 'cancelled').length;
  const pendingCount = tasks.filter(t => t.status === 'pending').length;
  const inProgressCount = tasks.filter(t => t.status === 'in_progress').length;
  const completedCount = tasks.filter(t => t.status === 'completed').length;
  const cancelledCount = tasks.filter(t => t.status === 'cancelled').length;

  return (
    <PageContainer
      data-testid="page-tasks"
      data-page="tasks"
      loading={loading}
    >
      <Card>
        {/* 统计卡片 */}
        <Row gutter={16} style={{ marginBottom: 24 }}>
          <Col span={4}>
            <Statistic
              title="总任务"
              value={tasks.length}
              prefix={<TeamOutlined />}
            />
          </Col>
          <Col span={4}>
            <Statistic
              title="待处理"
              value={pendingCount}
              valueStyle={{ color: '#cf1322' }}
            />
          </Col>
          <Col span={4}>
            <Statistic
              title="进行中"
              value={inProgressCount}
              valueStyle={{ color: '#1890ff' }}
            />
          </Col>
          <Col span={4}>
            <Statistic
              title="已完成"
              value={completedCount}
              valueStyle={{ color: '#52c41a' }}
            />
          </Col>
          <Col span={4}>
            <Statistic
              title="已取消"
              value={cancelledCount}
              valueStyle={{ color: '#999' }}
            />
          </Col>
          <Col span={4}>
            <Statistic
              title="待认领"
              value={unclaimedCount}
              prefix={<InboxOutlined />}
              valueStyle={{ color: '#faad14' }}
            />
          </Col>
        </Row>

        {/* 筛选栏 */}
        <Row gutter={16} style={{ marginBottom: 16 }}>
          <Col span={6}>
            <Input
              placeholder="搜索任务标题或描述"
              prefix={<SearchOutlined />}
              value={keywordFilter}
              onChange={(e) => setKeywordFilter(e.target.value)}
              allowClear
              onPressEnter={fetchTasks}
            />
          </Col>
          <Col span={4}>
            <Select
              placeholder="筛选状态"
              allowClear
              style={{ width: '100%' }}
              value={statusFilter}
              onChange={setStatusFilter}
            >
              <Option value="pending">待处理</Option>
              <Option value="in_progress">进行中</Option>
              <Option value="completed">已完成</Option>
              <Option value="cancelled">已取消</Option>
            </Select>
          </Col>
          <Col span={4}>
            <Select
              placeholder="筛选优先级"
              allowClear
              style={{ width: '100%' }}
              value={priorityFilter}
              onChange={setPriorityFilter}
            >
              <Option value="low">低</Option>
              <Option value="medium">中</Option>
              <Option value="high">高</Option>
              <Option value="urgent">紧急</Option>
            </Select>
          </Col>
          <Col span={6}>
            <Space>
              <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
                新建任务
              </Button>
              <Button icon={<ReloadOutlined />} onClick={fetchTasks}>
                刷新
              </Button>
            </Space>
          </Col>
        </Row>

        {/* 任务标签页 */}
        <Tabs activeKey={activeTab} onChange={setActiveTab}>
          <TabPane tab="全部任务" key="all" />
          <TabPane
            tab={
              <span>
                我的任务
                {myTasksCount > 0 && <Badge count={myTasksCount} style={{ marginLeft: 8 }} />}
              </span>
            }
            key="mine"
          />
          <TabPane
            tab={
              <span>
                待认领
                {unclaimedCount > 0 && <Badge count={unclaimedCount} style={{ marginLeft: 8 }} />}
              </span>
            }
            key="unclaimed"
          />
        </Tabs>

        {/* 任务列表 */}
        <Table
          columns={columns}
          dataSource={tasks}
          rowKey="id"
          pagination={{ pageSize: 10 }}
          data-testid="task-table"
        />
      </Card>

      {/* 创建/编辑任务弹窗 */}
      <Modal
        title={editingTask ? '编辑任务' : '新建任务'}
        open={modalVisible}
        onOk={handleModalOk}
        onCancel={() => setModalVisible(false)}
        width={600}
        destroyOnClose
      >
        <Form
          form={form}
          layout="vertical"
          initialValues={{ priority: 'medium' }}
        >
          {/* 指派模式选择 - 仅新建时显示 */}
          {!editingTask && (
            <Form.Item label="指派模式" required>
              <Radio.Group
                value={assignmentMode}
                onChange={(e) => setAssignmentMode(e.target.value)}
              >
                <Radio.Button value="assign">指派给某人</Radio.Button>
                <Radio.Button value="claim">放入待认领池</Radio.Button>
              </Radio.Group>
            </Form.Item>
          )}

          <Form.Item
            name="title"
            label="任务标题"
            rules={[{ required: true, message: '请输入任务标题' }]}
          >
            <Input placeholder="请输入任务标题" />
          </Form.Item>

          <Form.Item
            name="description"
            label="任务描述"
            rules={[{ required: true, message: '请输入任务描述' }]}
          >
            <TextArea rows={4} placeholder="请输入任务描述" />
          </Form.Item>

          <Form.Item
            name="priority"
            label="优先级"
            rules={[{ required: true, message: '请选择优先级' }]}
          >
            <Select placeholder="请选择优先级">
              <Option value="low">低</Option>
              <Option value="medium">中</Option>
              <Option value="high">高</Option>
              <Option value="urgent">紧急</Option>
            </Select>
          </Form.Item>

          {/* 执行人选择 - 仅在指派模式下显示（新建时）或编辑时显示 */}
          {(!editingTask && assignmentMode === 'assign') || editingTask ? (
            <Form.Item
              name="assignee_id"
              label="执行人"
              rules={[{ required: false }]} // 编辑时可以不选
            >
              <Select placeholder="请选择执行人" showSearch optionFilterProp="children" allowClear>
                {employees.map((emp) => (
                  <Option key={emp.id} value={emp.id}>
                    {emp.name} ({emp.email})
                  </Option>
                ))}
              </Select>
            </Form.Item>
          ) : null}

          <Form.Item name="due_date" label="截止时间">
            <DatePicker
              style={{ width: '100%' }}
              placeholder="请选择截止时间"
              showTime={{ format: 'HH:mm' }}
              format="YYYY-MM-DD HH:mm"
            />
          </Form.Item>
        </Form>
      </Modal>

      {/* 任务详情弹窗 */}
      <Modal
        title="任务详情"
        open={detailModalVisible}
        onCancel={() => setDetailModalVisible(false)}
        footer={null}
        width={600}
      >
        {viewingTask && (
          <div>
            <p>
              <strong>标题：</strong>
              {viewingTask.title}
            </p>
            <p>
              <strong>描述：</strong>
              {viewingTask.description}
            </p>
            <p>
              <strong>状态：</strong>
              {getStatusTag(viewingTask.status)}
            </p>
            <p>
              <strong>优先级：</strong>
              {getPriorityTag(viewingTask.priority)}
            </p>
            <p>
              <strong>执行人：</strong>
              {viewingTask.assignee_name || <Tag color="warning">待认领</Tag>}
            </p>
            <p>
              <strong>创建人：</strong>
              {viewingTask.creator_name}
            </p>
            <p>
              <strong>截止时间：</strong>
              {viewingTask.due_date
                ? dayjs(viewingTask.due_date).format('YYYY-MM-DD HH:mm')
                : '-'}
            </p>
            <p>
              <strong>创建时间：</strong>
              {dayjs(viewingTask.created_at).format('YYYY-MM-DD HH:mm')}
            </p>
          </div>
        )}
      </Modal>
    </PageContainer>
  );
};

export default Tasks;
