import React, { useEffect, useState } from 'react';
import { Table, Button, Tag, Space, Modal, Form, Input, Select, Switch, message, Card, Steps, Tabs, List, Row, Col, Statistic, Popconfirm, Empty } from 'antd';
import { PlusOutlined, PlayCircleOutlined, EditOutlined, DeleteOutlined, ReloadOutlined, CheckCircleOutlined, CloseCircleOutlined, ClockCircleOutlined } from '@ant-design/icons';
import { workflowApi } from '../api/workflow';
import dayjs from 'dayjs';

const { TextArea } = Input;
const { Option } = Select;
const { Step } = Steps;
const { TabPane } = Tabs;

interface Workflow {
  id: string;
  name: string;
  description: string;
  status: 'active' | 'inactive';
  trigger_type: 'manual' | 'keyword' | 'webhook' | 'schedule';
  created_at: string;
  updated_at: string;
  steps: WorkflowStep[];
}

interface WorkflowStep {
  id: string;
  name: string;
  type: string;
  config: any;
  order: number;
}

interface WorkflowExecution {
  id: string;
  workflow_id: string;
  status: 'running' | 'completed' | 'failed' | 'cancelled';
  input: any;
  output: any;
  started_at: string;
  completed_at?: string;
  error?: string;
}

const Workflows: React.FC = () => {
  const [workflows, setWorkflows] = useState<Workflow[]>([]);
  const [executions, setExecutions] = useState<WorkflowExecution[]>([]);
  const [loading, setLoading] = useState(false);
  const [executionsLoading, setExecutionsLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [detailVisible, setDetailVisible] = useState(false);
  const [executionModalVisible, setExecutionModalVisible] = useState(false);
  const [selectedWorkflow, setSelectedWorkflow] = useState<Workflow | null>(null);
  const [selectedExecution, setSelectedExecution] = useState<WorkflowExecution | null>(null);
  const [form] = Form.useForm();

  useEffect(() => {
    fetchWorkflows();
  }, []);

  const fetchWorkflows = async () => {
    setLoading(true);
    try {
      const res = await workflowApi.list(1, 50);
      if (res.code === 0) {
        setWorkflows(res.data.list || []);
      }
    } finally {
      setLoading(false);
    }
  };

  const fetchExecutions = async (workflowId: string) => {
    setExecutionsLoading(true);
    try {
      const res = await workflowApi.getExecutions(workflowId, 1, 20);
      if (res.code === 0) {
        setExecutions(res.data.list || []);
      }
    } finally {
      setExecutionsLoading(false);
    }
  };

  const handleCreate = async (values: any) => {
    try {
      const res = await workflowApi.create(values);
      if (res.code === 0) {
        message.success('工作流创建成功');
        setModalVisible(false);
        form.resetFields();
        fetchWorkflows();
      }
    } catch (error) {
      message.error('创建失败');
    }
  };

  const handleToggleStatus = async (id: string, status: string) => {
    try {
      const newStatus = status === 'active' ? 'inactive' : 'active';
      const res = await workflowApi.updateStatus(id, newStatus);
      if (res.code === 0) {
        message.success('状态更新成功');
        fetchWorkflows();
      }
    } catch (error) {
      message.error('更新失败');
    }
  };

  const handleDelete = async (id: string) => {
    try {
      const res = await workflowApi.delete(id);
      if (res.code === 0) {
        message.success('删除成功');
        fetchWorkflows();
      }
    } catch (error) {
      message.error('删除失败');
    }
  };

  const handleTrigger = async (id: string) => {
    try {
      const res = await workflowApi.trigger(id, {});
      if (res.code === 0) {
        message.success('工作流已触发');
        // 刷新执行历史
        if (selectedWorkflow?.id === id) {
          fetchExecutions(id);
        }
      }
    } catch (error) {
      message.error('触发失败');
    }
  };

  const showDetail = (workflow: Workflow) => {
    setSelectedWorkflow(workflow);
    setDetailVisible(true);
    fetchExecutions(workflow.id);
  };

  const showExecutionDetail = (execution: WorkflowExecution) => {
    setSelectedExecution(execution);
    setExecutionModalVisible(true);
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active': return 'success';
      case 'inactive': return 'default';
      default: return 'default';
    }
  };

  const getStatusText = (status: string) => {
    switch (status) {
      case 'active': return '激活';
      case 'inactive': return '停用';
      default: return status;
    }
  };

  const getTriggerTypeText = (type: string) => {
    const map: Record<string, string> = {
      manual: '手动',
      keyword: '关键词',
      webhook: 'Webhook',
      schedule: '定时',
    };
    return map[type] || type;
  };

  const getExecutionStatusColor = (status: string) => {
    switch (status) {
      case 'completed': return 'success';
      case 'running': return 'processing';
      case 'failed': return 'error';
      case 'cancelled': return 'default';
      default: return 'default';
    }
  };

  const getExecutionStatusText = (status: string) => {
    switch (status) {
      case 'completed': return '已完成';
      case 'running': return '运行中';
      case 'failed': return '失败';
      case 'cancelled': return '已取消';
      default: return status;
    }
  };

  const getExecutionStatusIcon = (status: string) => {
    switch (status) {
      case 'completed': return <CheckCircleOutlined style={{ color: '#52c41a' }} />;
      case 'running': return <ClockCircleOutlined style={{ color: '#1890ff' }} />;
      case 'failed': return <CloseCircleOutlined style={{ color: '#f5222d' }} />;
      case 'cancelled': return <CloseCircleOutlined style={{ color: '#999' }} />;
      default: return null;
    }
  };

  // 统计数据
  const stats = {
    total: workflows.length,
    active: workflows.filter(w => w.status === 'active').length,
    inactive: workflows.filter(w => w.status === 'inactive').length,
  };

  const columns = [
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
      render: (text: string, record: Workflow) => (
        <a onClick={() => showDetail(record)}>{text}</a>
      ),
    },
    {
      title: '描述',
      dataIndex: 'description',
      key: 'description',
      ellipsis: true,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: string) => (
        <Tag color={getStatusColor(status)}>{getStatusText(status)}</Tag>
      ),
    },
    {
      title: '触发方式',
      dataIndex: 'trigger_type',
      key: 'trigger_type',
      width: 100,
      render: (type: string) => getTriggerTypeText(type),
    },
    {
      title: '步骤数',
      dataIndex: 'steps',
      key: 'steps',
      width: 80,
      render: (steps: any[]) => steps?.length || 0,
    },
    {
      title: '操作',
      key: 'action',
      width: 200,
      render: (_: any, record: Workflow) => (
        <Space size="small">
          <Button
            type="primary"
            size="small"
            icon={<PlayCircleOutlined />}
            onClick={() => handleTrigger(record.id)}
          >
            运行
          </Button>
          <Switch
            checked={record.status === 'active'}
            onChange={() => handleToggleStatus(record.id, record.status)}
            size="small"
          />
          <Button
            size="small"
            icon={<EditOutlined />}
            onClick={() => showDetail(record)}
          />
          <Popconfirm
            title="确认删除"
            description="确定要删除这个工作流吗？"
            onConfirm={() => handleDelete(record.id)}
            okText="确定"
            cancelText="取消"
          >
            <Button size="small" danger icon={<DeleteOutlined />} />
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div>
      {/* 统计卡片 */}
      <Row gutter={16} style={{ marginBottom: 24 }}>
        <Col span={8}>
          <Card size="small">
            <Statistic title="总工作流" value={stats.total} />
          </Card>
        </Col>
        <Col span={8}>
          <Card size="small">
            <Statistic title="激活" value={stats.active} valueStyle={{ color: '#52c41a' }} />
          </Card>
        </Col>
        <Col span={8}>
          <Card size="small">
            <Statistic title="停用" value={stats.inactive} />
          </Card>
        </Col>
      </Row>

      {/* 操作栏 */}
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <h1 style={{ margin: 0 }}>工作流管理</h1>
        <Space>
          <Button icon={<ReloadOutlined />} onClick={fetchWorkflows}>
            刷新
          </Button>
          <Button type="primary" icon={<PlusOutlined />} onClick={() => setModalVisible(true)}>
            新建工作流
          </Button>
        </Space>
      </div>

      {/* 工作流表格 */}
      <Table
        columns={columns}
        dataSource={workflows}
        rowKey="id"
        loading={loading}
        pagination={{ pageSize: 10, showSizeChanger: true, showTotal: (total) => `共 ${total} 条` }}
      />

      {/* 新建工作流弹窗 */}
      <Modal
        title="新建工作流"
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        footer={null}
        width={600}
      >
        <Form form={form} onFinish={handleCreate} layout="vertical">
          <Form.Item
            name="name"
            label="名称"
            rules={[{ required: true, message: '请输入工作流名称' }]}
          >
            <Input placeholder="工作流名称" />
          </Form.Item>
          <Form.Item
            name="description"
            label="描述"
          >
            <TextArea rows={2} placeholder="工作流描述" />
          </Form.Item>
          <Form.Item
            name="trigger_type"
            label="触发方式"
            rules={[{ required: true, message: '请选择触发方式' }]}
          >
            <Select placeholder="选择触发方式">
              <Option value="manual">手动</Option>
              <Option value="keyword">关键词</Option>
              <Option value="webhook">Webhook</Option>
              <Option value="schedule">定时</Option>
            </Select>
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" block>
              创建
            </Button>
          </Form.Item>
        </Form>
      </Modal>

      {/* 工作流详情弹窗 */}
      <Modal
        title={selectedWorkflow?.name || '工作流详情'}
        open={detailVisible}
        onCancel={() => setDetailVisible(false)}
        footer={null}
        width={900}
      >
        {selectedWorkflow && (
          <Tabs defaultActiveKey="info">
            <TabPane tab="基本信息" key="info">
              <Card size="small" style={{ marginBottom: 16 }}>
                <Row gutter={16}>
                  <Col span={12}>
                    <p><strong>名称：</strong>{selectedWorkflow.name}</p>
                    <p><strong>描述：</strong>{selectedWorkflow.description || '-'}</p>
                  </Col>
                  <Col span={12}>
                    <p>
                      <strong>状态：</strong>
                      <Tag color={getStatusColor(selectedWorkflow.status)}>
                        {getStatusText(selectedWorkflow.status)}
                      </Tag>
                    </p>
                    <p><strong>触发方式：</strong>{getTriggerTypeText(selectedWorkflow.trigger_type)}</p>
                  </Col>
                </Row>
                <p><strong>创建时间：</strong>{dayjs(selectedWorkflow.created_at).format('YYYY-MM-DD HH:mm:ss')}</p>
              </Card>
              <Card title="执行步骤" size="small">
                {selectedWorkflow.steps && selectedWorkflow.steps.length > 0 ? (
                  <Steps direction="vertical" size="small">
                    {selectedWorkflow.steps.map((step, index) => (
                      <Step
                        key={step.id || index}
                        title={step.name}
                        description={`类型: ${step.type} | 顺序: ${step.order}`}
                      />
                    ))}
                  </Steps>
                ) : (
                  <Empty description="暂无步骤" />
                )}
              </Card>
            </TabPane>
            <TabPane tab="执行历史" key="executions">
              <List
                loading={executionsLoading}
                dataSource={executions}
                renderItem={(execution) => (
                  <List.Item
                    key={execution.id}
                    actions={[
                      <Button
                        type="link"
                        size="small"
                        onClick={() => showExecutionDetail(execution)}
                      >
                        查看详情
                      </Button>,
                    ]}
                  >
                    <List.Item.Meta
                      avatar={getExecutionStatusIcon(execution.status)}
                      title={
                        <Space>
                          <span>执行 #{execution.id.slice(-6)}</span>
                          <Tag color={getExecutionStatusColor(execution.status)}>
                            {getExecutionStatusText(execution.status)}
                          </Tag>
                        </Space>
                      }
                      description={
                        <Space direction="vertical" size={0}>
                          <span>开始: {dayjs(execution.started_at).format('MM-DD HH:mm:ss')}</span>
                          {execution.completed_at && (
                            <span>完成: {dayjs(execution.completed_at).format('MM-DD HH:mm:ss')}</span>
                          )}
                        </Space>
                      }
                    />
                  </List.Item>
                )}
              />
            </TabPane>
          </Tabs>
        )}
      </Modal>

      {/* 执行详情弹窗 */}
      <Modal
        title="执行详情"
        open={executionModalVisible}
        onCancel={() => setExecutionModalVisible(false)}
        footer={null}
        width={700}
      >
        {selectedExecution && (
          <div>
            <Card size="small" style={{ marginBottom: 16 }}>
              <Row gutter={16}>
                <Col span={12}>
                  <p><strong>执行ID：</strong>#{selectedExecution.id.slice(-8)}</p>
                  <p>
                    <strong>状态：</strong>
                    <Tag color={getExecutionStatusColor(selectedExecution.status)}>
                      {getExecutionStatusText(selectedExecution.status)}
                    </Tag>
                  </p>
                </Col>
                <Col span={12}>
                  <p><strong>开始时间：</strong>{dayjs(selectedExecution.started_at).format('YYYY-MM-DD HH:mm:ss')}</p>
                  {selectedExecution.completed_at && (
                    <p><strong>完成时间：</strong>{dayjs(selectedExecution.completed_at).format('YYYY-MM-DD HH:mm:ss')}</p>
                  )}
                </Col>
              </Row>
            </Card>
            
            <Card title="输入参数" size="small" style={{ marginBottom: 16 }}>
              <pre style={{ background: '#f5f5f5', padding: 12, borderRadius: 4, overflow: 'auto' }}>
                {JSON.stringify(selectedExecution.input, null, 2)}
              </pre>
            </Card>
            
            {selectedExecution.output && (
              <Card title="输出结果" size="small" style={{ marginBottom: 16 }}>
                <pre style={{ background: '#f5f5f5', padding: 12, borderRadius: 4, overflow: 'auto' }}>
                  {JSON.stringify(selectedExecution.output, null, 2)}
                </pre>
              </Card>
            )}
            
            {selectedExecution.error && (
              <Card title="错误信息" size="small">
                <pre style={{ background: '#fff2f0', padding: 12, borderRadius: 4, color: '#f5222d', overflow: 'auto' }}>
                  {selectedExecution.error}
                </pre>
              </Card>
            )}
          </div>
        )}
      </Modal>
    </div>
  );
};

export default Workflows;
