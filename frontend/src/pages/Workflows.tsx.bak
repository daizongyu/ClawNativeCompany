import React, { useEffect, useState } from 'react';
import { Table, Button, Tag, Space, Modal, Form, Input, Select, Switch, message, Card, Steps, Tabs, Row, Col, Statistic, Popconfirm, Empty } from 'antd';
import { PlusOutlined, PlayCircleOutlined, EditOutlined, DeleteOutlined, ReloadOutlined } from '@ant-design/icons';
import { workflowApi } from '../api/workflow';
import dayjs from 'dayjs';
import { PageContainer } from '../components/common';

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
  started_at: string;
  completed_at?: string;
  result?: any;
}

const Workflows: React.FC = () => {
  const [workflows, setWorkflows] = useState<Workflow[]>([]);
  const [executions, setExecutions] = useState<WorkflowExecution[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [detailModalVisible, setDetailModalVisible] = useState(false);
  const [executionModalVisible, setExecutionModalVisible] = useState(false);
  const [editingWorkflow, setEditingWorkflow] = useState<Workflow | null>(null);
  const [viewingWorkflow, setViewingWorkflow] = useState<Workflow | null>(null);
  const [viewingExecution, setViewingExecution] = useState<WorkflowExecution | null>(null);
  const [activeTab, setActiveTab] = useState('workflows');
  const [form] = Form.useForm();

  // 设置当前页面
  useEffect(() => {
    if (typeof window !== 'undefined' && window.__CLAW_TEST__) {
      window.__CLAW_TEST__.setCurrentPage('workflows');
    }
  }, []);

  // 暴露测试函数
  useEffect(() => {
    if (typeof window !== 'undefined') {
      (window as any).__TEST_WORKFLOWS__ = {
        openModal: () => setModalVisible(true),
        closeModal: () => setModalVisible(false),
        getWorkflows: () => workflows,
        setEditingWorkflow: (wf: Workflow | null) => setEditingWorkflow(wf),
      };
    }
  }, [workflows]);

  const fetchWorkflows = async () => {
    setLoading(true);
    try {
      const res = await workflowApi.list(1, 100);
      if (res.code === 0) {
        // 后端返回的数据格式是 { list: [...], total: n, page: 1, page_size: 20, total_page: 1 }
        const workflowList = res.data.list || res.data.items || [];
        setWorkflows(workflowList);
      }
    } catch (error) {
      console.error('获取工作流列表失败:', error);
    } finally {
      setLoading(false);
    }
  };

  const fetchExecutions = async () => {
    try {
      // 获取所有执行记录需要遍历所有工作流
      const allExecutions: WorkflowExecution[] = [];
      for (const workflow of workflows) {
        const res = await workflowApi.getExecutions(workflow.id, 1, 50);
        if (res.code === 0 && res.data.items) {
          allExecutions.push(...res.data.items);
        }
      }
      setExecutions(allExecutions);
    } catch (error) {
      console.error('获取执行记录失败:', error);
    }
  };

  useEffect(() => {
    fetchWorkflows();
  }, []);

  useEffect(() => {
    if (workflows.length > 0) {
      fetchExecutions();
    }
  }, [workflows]);

  const handleCreate = () => {
    setEditingWorkflow(null);
    form.resetFields();
    setModalVisible(true);
  };

  const handleEdit = (record: Workflow) => {
    setEditingWorkflow(record);
    form.setFieldsValue(record);
    setModalVisible(true);
  };

  const handleView = (record: Workflow) => {
    setViewingWorkflow(record);
    setDetailModalVisible(true);
  };

  const handleDelete = async (id: string) => {
    try {
      const res = await workflowApi.delete(id);
      if (res.code === 0) {
        message.success('删除成功');
        fetchWorkflows();
      } else {
        message.error(res.message || '删除失败');
      }
    } catch (error) {
      console.error('删除工作流失败:', error);
    }
  };

  const handleExecute = async (id: string) => {
    try {
      const res = await workflowApi.trigger(id, {});
      if (res.code === 0) {
        message.success('工作流执行已启动');
        fetchExecutions();
      } else {
        message.error(res.message || '执行失败');
      }
    } catch (error) {
      console.error('执行工作流失败:', error);
    }
  };

  const handleToggleStatus = async (record: Workflow) => {
    try {
      const newStatus = record.status === 'active' ? 'inactive' : 'active';
      const res = await workflowApi.updateStatus(record.id, newStatus);
      if (res.code === 0) {
        message.success(`工作流已${newStatus === 'active' ? '启用' : '禁用'}`);
        fetchWorkflows();
      } else {
        message.error(res.message || '操作失败');
      }
    } catch (error) {
      console.error('切换工作流状态失败:', error);
    }
  };

  const handleViewExecution = (record: WorkflowExecution) => {
    setViewingExecution(record);
    setExecutionModalVisible(true);
  };

  const handleModalOk = async () => {
    try {
      const values = await form.validateFields();
      if (editingWorkflow) {
        const res = await workflowApi.update(editingWorkflow.id, values);
        if (res.code === 0) {
          message.success('更新成功');
          setModalVisible(false);
          fetchWorkflows();
        } else {
          message.error(res.message || '更新失败');
        }
      } else {
        const res = await workflowApi.create(values);
        if (res.code === 0) {
          message.success('创建成功');
          setModalVisible(false);
          fetchWorkflows();
        } else {
          message.error(res.message || '创建失败');
        }
      }
    } catch (error) {
      console.error('保存工作流失败:', error);
    }
  };

  const getStatusColor = (status: string) => {
    const colors: Record<string, string> = {
      active: 'green',
      inactive: 'red',
    };
    return colors[status] || 'default';
  };

  const getStatusText = (status: string) => {
    const texts: Record<string, string> = {
      active: '启用',
      inactive: '禁用',
    };
    return texts[status] || status;
  };

  const getTriggerTypeText = (type: string) => {
    const texts: Record<string, string> = {
      manual: '手动',
      keyword: '关键词',
      webhook: 'Webhook',
      schedule: '定时',
    };
    return texts[type] || type;
  };

  const getExecutionStatusColor = (status: string) => {
    const colors: Record<string, string> = {
      running: 'blue',
      completed: 'green',
      failed: 'red',
      cancelled: 'orange',
    };
    return colors[status] || 'default';
  };

  const getExecutionStatusText = (status: string) => {
    const texts: Record<string, string> = {
      running: '运行中',
      completed: '已完成',
      failed: '失败',
      cancelled: '已取消',
    };
    return texts[status] || status;
  };

  const workflowColumns = [
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
      render: (text: string, record: Workflow) => (
        <Button
          type="link"
          onClick={() => handleView(record)}
          style={{ padding: 0 }}
          data-testid={`workflow-name-${record.id}`}
        >
          {text}
        </Button>
      ),
    },
    {
      title: '触发方式',
      dataIndex: 'trigger_type',
      key: 'trigger_type',
      render: (type: string) => <Tag>{getTriggerTypeText(type)}</Tag>,
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
      title: '步骤数',
      dataIndex: 'steps',
      key: 'steps',
      render: (steps: WorkflowStep[]) => steps?.length || 0,
    },
    {
      title: '更新时间',
      dataIndex: 'updated_at',
      key: 'updated_at',
      render: (date: string) => dayjs(date).format('YYYY-MM-DD HH:mm'),
    },
    {
      title: '操作',
      key: 'action',
      render: (_: any, record: Workflow) => (
        <Space size="middle">
          <Button
            type="primary"
            icon={<PlayCircleOutlined />}
            onClick={() => handleExecute(record.id)}
            disabled={record.status === 'inactive'}
            data-testid={`workflow-execute-btn-${record.id}`}
            data-action="execute"
            data-entity="workflow"
          >
            执行
          </Button>
          <Button
            icon={<EditOutlined />}
            onClick={() => handleEdit(record)}
            data-testid={`workflow-edit-btn-${record.id}`}
            data-action="edit"
            data-entity="workflow"
          >
            编辑
          </Button>
          <Button
            icon={<ReloadOutlined />}
            onClick={() => handleToggleStatus(record)}
            data-testid={`workflow-toggle-btn-${record.id}`}
            data-action="toggle-status"
            data-entity="workflow"
          >
            {record.status === 'active' ? '禁用' : '启用'}
          </Button>
          <Popconfirm
            title="确定删除该工作流吗？"
            onConfirm={() => handleDelete(record.id)}
            okText="确定"
            cancelText="取消"
          >
            <Button
              danger
              icon={<DeleteOutlined />}
              data-testid={`workflow-delete-btn-${record.id}`}
              data-action="delete"
              data-entity="workflow"
            >
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  const executionColumns = [
    {
      title: '工作流ID',
      dataIndex: 'workflow_id',
      key: 'workflow_id',
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <Tag color={getExecutionStatusColor(status)}>{getExecutionStatusText(status)}</Tag>
      ),
    },
    {
      title: '开始时间',
      dataIndex: 'started_at',
      key: 'started_at',
      render: (date: string) => dayjs(date).format('YYYY-MM-DD HH:mm:ss'),
    },
    {
      title: '完成时间',
      dataIndex: 'completed_at',
      key: 'completed_at',
      render: (date: string) => (date ? dayjs(date).format('YYYY-MM-DD HH:mm:ss') : '-'),
    },
    {
      title: '操作',
      key: 'action',
      render: (_: any, record: WorkflowExecution) => (
        <Button
          type="link"
          onClick={() => handleViewExecution(record)}
          data-testid={`execution-view-btn-${record.id}`}
          data-action="view-execution"
          data-entity="workflow"
        >
          查看详情
        </Button>
      ),
    },
  ];

  // 统计
  const activeCount = workflows.filter((w) => w.status === 'active').length;
  const inactiveCount = workflows.filter((w) => w.status === 'inactive').length;
  const runningCount = executions.filter((e) => e.status === 'running').length;

  return (
    <PageContainer
      data-testid="page-workflows"
      data-page="workflows"
      loading={loading}
    >
      <div style={{ padding: '24px' }}>
        <Row gutter={[16, 16]} style={{ marginBottom: '24px' }}>
          <Col xs={24} sm={8}>
            <Card data-testid="workflow-stat-active">
              <Statistic title="启用工作流" value={activeCount} valueStyle={{ color: '#3f8600' }} />
            </Card>
          </Col>
          <Col xs={24} sm={8}>
            <Card data-testid="workflow-stat-inactive">
              <Statistic title="禁用工作流" value={inactiveCount} valueStyle={{ color: '#cf1322' }} />
            </Card>
          </Col>
          <Col xs={24} sm={8}>
            <Card data-testid="workflow-stat-running">
              <Statistic title="运行中" value={runningCount} valueStyle={{ color: '#1890ff' }} />
            </Card>
          </Col>
        </Row>

        <Tabs activeKey={activeTab} onChange={setActiveTab} data-testid="workflow-tabs">
          <TabPane tab="工作流列表" key="workflows">
            <div style={{ marginBottom: '16px', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <h1>工作流管理</h1>
              <Button
                type="primary"
                icon={<PlusOutlined />}
                onClick={handleCreate}
                data-testid="workflow-create-btn"
                data-action="create"
                data-entity="workflow"
              >
                新建工作流
              </Button>
            </div>

            <Table
              columns={workflowColumns}
              dataSource={workflows}
              rowKey="id"
              data-testid="workflow-table"
              data-entity="workflow"
              onRow={(record) => ({
                'data-testid': `workflow-row-${record.id}`,
                'data-workflow-id': record.id,
              } as any)}
            />
          </TabPane>

          <TabPane tab="执行记录" key="executions">
            <div style={{ marginBottom: '16px' }}>
              <h1>执行记录</h1>
            </div>

            <Table
              columns={executionColumns}
              dataSource={executions}
              rowKey="id"
              data-testid="execution-table"
              data-entity="workflow-execution"
            />
          </TabPane>
        </Tabs>

        {/* 编辑/创建模态框 */}
        <Modal
          title={editingWorkflow ? '编辑工作流' : '新建工作流'}
          open={modalVisible}
          onOk={handleModalOk}
          onCancel={() => setModalVisible(false)}
          destroyOnClose
          width={700}
          data-testid="workflow-modal"
        >
          <Form form={form} layout="vertical">
            <Form.Item
              label="名称"
              name="name"
              rules={[{ required: true, message: '请输入工作流名称' }]}
            >
              <Input
                placeholder="请输入工作流名称"
                data-testid="input-workflow-name"
                data-input-name="workflow-name"
              />
            </Form.Item>
            <Form.Item
              label="描述"
              name="description"
            >
              <TextArea
                rows={4}
                placeholder="请输入工作流描述"
                data-testid="input-workflow-description"
                data-input-name="workflow-description"
              />
            </Form.Item>
            <Form.Item
              label="触发方式"
              name="trigger_type"
              rules={[{ required: true, message: '请选择触发方式' }]}
            >
              <Select
                placeholder="请选择触发方式"
                data-testid="input-workflow-trigger-type"
                data-input-name="workflow-trigger-type"
              >
                <Option value="manual">手动</Option>
                <Option value="keyword">关键词</Option>
                <Option value="webhook">Webhook</Option>
                <Option value="schedule">定时</Option>
              </Select>
            </Form.Item>
            <Form.Item
              label="状态"
              name="status"
              valuePropName="checked"
            >
              <Switch
                checkedChildren="启用"
                unCheckedChildren="禁用"
                data-testid="input-workflow-status"
                data-input-name="workflow-status"
              />
            </Form.Item>
          </Form>
        </Modal>

        {/* 详情模态框 */}
        <Modal
          title="工作流详情"
          open={detailModalVisible}
          onOk={() => setDetailModalVisible(false)}
          onCancel={() => setDetailModalVisible(false)}
          width={800}
          data-testid="workflow-detail-modal"
        >
          {viewingWorkflow && (
            <div>
              <p><strong>名称：</strong>{viewingWorkflow.name}</p>
              <p><strong>描述：</strong>{viewingWorkflow.description || '无'}</p>
              <p>
                <strong>触发方式：</strong>
                <Tag>{getTriggerTypeText(viewingWorkflow.trigger_type)}</Tag>
              </p>
              <p>
                <strong>状态：</strong>
                <Tag color={getStatusColor(viewingWorkflow.status)}>
                  {getStatusText(viewingWorkflow.status)}
                </Tag>
              </p>
              <p><strong>创建时间：</strong>{dayjs(viewingWorkflow.created_at).format('YYYY-MM-DD HH:mm')}</p>
              <p><strong>更新时间：</strong>{dayjs(viewingWorkflow.updated_at).format('YYYY-MM-DD HH:mm')}</p>

              <h3 style={{ marginTop: '24px' }}>工作流步骤</h3>
              {viewingWorkflow.steps && viewingWorkflow.steps.length > 0 ? (
                <Steps direction="vertical" current={-1}>
                  {viewingWorkflow.steps.map((step) => (
                    <Step
                      key={step.id}
                      title={step.name}
                      description={`类型: ${step.type}`}
                    />
                  ))}
                </Steps>
              ) : (
                <Empty description="暂无步骤" />
              )}
            </div>
          )}
        </Modal>

        {/* 执行详情模态框 */}
        <Modal
          title="执行详情"
          open={executionModalVisible}
          onOk={() => setExecutionModalVisible(false)}
          onCancel={() => setExecutionModalVisible(false)}
          data-testid="execution-detail-modal"
        >
          {viewingExecution && (
            <div>
              <p><strong>执行ID：</strong>{viewingExecution.id}</p>
              <p><strong>工作流ID：</strong>{viewingExecution.workflow_id}</p>
              <p>
                <strong>状态：</strong>
                <Tag color={getExecutionStatusColor(viewingExecution.status)}>
                  {getExecutionStatusText(viewingExecution.status)}
                </Tag>
              </p>
              <p><strong>开始时间：</strong>{dayjs(viewingExecution.started_at).format('YYYY-MM-DD HH:mm:ss')}</p>
              <p><strong>完成时间：</strong>{viewingExecution.completed_at ? dayjs(viewingExecution.completed_at).format('YYYY-MM-DD HH:mm:ss') : '-'}</p>
              {viewingExecution.result && (
                <div>
                  <strong>执行结果：</strong>
                  <pre style={{ background: '#f5f5f5', padding: '12px', borderRadius: '4px', marginTop: '8px' }}>
                    {JSON.stringify(viewingExecution.result, null, 2)}
                  </pre>
                </div>
              )}
            </div>
          )}
        </Modal>
      </div>
    </PageContainer>
  );
};

export default Workflows;
