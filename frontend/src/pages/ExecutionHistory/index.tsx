import React, { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
  Table,
  Tag,
  Button,
  Drawer,
  Space,
  message,
  Row,
  Col,
  Card,
  Statistic,
  Select,
} from 'antd';
import {
  ReloadOutlined,
  EyeOutlined,
  RedoOutlined,
  ArrowLeftOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  ClockCircleOutlined,
  LoadingOutlined,
} from '@ant-design/icons';
import dayjs from 'dayjs';
import { workflowApi } from '../../api/workflow';
import { PageContainer } from '../../components/common';
import { ExecutionTimeline } from './components/ExecutionTimeline';

const { Option } = Select;

interface WorkflowExecution {
  id: string;
  workflow_id: string;
  workflow_name: string;
  status: 'running' | 'completed' | 'failed' | 'cancelled';
  started_at: string;
  completed_at?: string;
  duration?: number;
  error?: string;
  steps?: ExecutionStep[];
}

interface ExecutionStep {
  id: string;
  step_id: string;
  step_name: string;
  step_type: string;
  status: 'pending' | 'running' | 'completed' | 'failed' | 'skipped';
  input?: any;
  output?: any;
  error?: string;
  started_at?: string;
  completed_at?: string;
}

const ExecutionHistory: React.FC = () => {
  const { workflowId } = useParams<{ workflowId: string }>();
  const navigate = useNavigate();
  const [executions, setExecutions] = useState<WorkflowExecution[]>([]);
  const [loading, setLoading] = useState(false);
  const [selectedExecution, setSelectedExecution] = useState<WorkflowExecution | null>(null);
  const [drawerVisible, setDrawerVisible] = useState(false);
  const [statusFilter, setStatusFilter] = useState<string>('');

  useEffect(() => {
    fetchExecutions();
  }, [workflowId, statusFilter]);

  const fetchExecutions = async () => {
    setLoading(true);
    try {
      const params: any = { page: 1, pageSize: 50 };
      if (statusFilter) params.status = statusFilter;

      let res;
      if (workflowId) {
        res = await workflowApi.getExecutions(workflowId, 1, 50);
      } else {
        // TODO: 获取所有执行记录
        res = { code: 0, data: { list: [] } };
      }

      if (res.code === 0) {
        const list = res.data?.list || res.data?.items || [];
        setExecutions(list);
      }
    } catch (error) {
      console.error('获取执行记录失败:', error);
      message.error('获取执行记录失败');
    } finally {
      setLoading(false);
    }
  };

  const handleViewDetail = async (execution: WorkflowExecution) => {
    try {
      // 获取执行详情
      const res = await workflowApi.getExecutionDetail(execution.id);
      if (res.code === 0) {
        setSelectedExecution({ ...execution, steps: res.data?.steps || [] });
        setDrawerVisible(true);
      }
    } catch (error) {
      console.error('获取执行详情失败:', error);
      // 如果没有详情API，直接显示基本信息
      setSelectedExecution(execution);
      setDrawerVisible(true);
    }
  };

  const handleRetry = async (executionId: string) => {
    try {
      const res = await workflowApi.retryExecution(executionId);
      if (res.code === 0) {
        message.success('重试成功');
        fetchExecutions();
      } else {
        message.error(res.message || '重试失败');
      }
    } catch (error) {
      console.error('重试失败:', error);
      message.error('重试失败');
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'completed':
        return <CheckCircleOutlined style={{ color: '#52c41a' }} />;
      case 'failed':
        return <CloseCircleOutlined style={{ color: '#ff4d4f' }} />;
      case 'running':
        return <LoadingOutlined style={{ color: '#1890ff' }} />;
      case 'cancelled':
        return <ClockCircleOutlined style={{ color: '#faad14' }} />;
      default:
        return null;
    }
  };

  const getStatusColor = (status: string): string => {
    const colorMap: Record<string, string> = {
      completed: 'success',
      failed: 'error',
      running: 'processing',
      cancelled: 'warning',
    };
    return colorMap[status] || 'default';
  };

  const getStatusLabel = (status: string): string => {
    const labelMap: Record<string, string> = {
      completed: '已完成',
      failed: '失败',
      running: '执行中',
      cancelled: '已取消',
    };
    return labelMap[status] || status;
  };

  const formatDuration = (ms?: number): string => {
    if (!ms) return '-';
    if (ms < 1000) return `${ms}ms`;
    if (ms < 60000) return `${Math.round(ms / 1000)}s`;
    return `${Math.round(ms / 60000)}m ${Math.round((ms % 60000) / 1000)}s`;
  };

  const columns = [
    {
      title: '执行ID',
      dataIndex: 'id',
      key: 'id',
      render: (id: string) => <code>{id.slice(0, 8)}...</code>,
    },
    {
      title: '工作流',
      dataIndex: 'workflow_name',
      key: 'workflow_name',
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <Tag color={getStatusColor(status)} icon={getStatusIcon(status)}>
          {getStatusLabel(status)}
        </Tag>
      ),
    },
    {
      title: '开始时间',
      dataIndex: 'started_at',
      key: 'started_at',
      render: (time: string) => dayjs(time).format('YYYY-MM-DD HH:mm:ss'),
    },
    {
      title: '耗时',
      key: 'duration',
      render: (_: any, record: WorkflowExecution) => formatDuration(record.duration),
    },
    {
      title: '操作',
      key: 'action',
      render: (_: any, record: WorkflowExecution) => (
        <Space size="small">
          <Button
            type="link"
            size="small"
            icon={<EyeOutlined />}
            onClick={() => handleViewDetail(record)}
          >
            详情
          </Button>
          {record.status === 'failed' && (
            <Button
              type="link"
              size="small"
              icon={<RedoOutlined />}
              onClick={() => handleRetry(record.id)}
            >
              重试
            </Button>
          )}
        </Space>
      ),
    },
  ];

  // 统计
  const completedCount = executions.filter(e => e.status === 'completed').length;
  const failedCount = executions.filter(e => e.status === 'failed').length;

  return (
    <PageContainer
      data-testid="page-execution-history"
      data-page="execution-history"
      loading={loading}
    >
      {/* 顶部操作栏 */}
      <Row justify="space-between" style={{ marginBottom: 24 }}>
        <Col>
          <Space>
            <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/workflows')}>
              返回工作流
            </Button>
            {workflowId && <span>工作流ID: {workflowId}</span>}
          </Space>
        </Col>
        <Col>
          <Space>
            <Select
              placeholder="筛选状态"
              allowClear
              value={statusFilter}
              onChange={setStatusFilter}
              style={{ width: 120 }}
            >
              <Option value="running">执行中</Option>
              <Option value="completed">已完成</Option>
              <Option value="failed">失败</Option>
              <Option value="cancelled">已取消</Option>
            </Select>
            <Button icon={<ReloadOutlined />} onClick={fetchExecutions}>
              刷新
            </Button>
          </Space>
        </Col>
      </Row>

      {/* 统计卡片 */}
      <Row gutter={16} style={{ marginBottom: 24 }}>
        <Col span={8}>
          <Card>
            <Statistic
              title="总执行次数"
              value={executions.length}
              prefix={<ClockCircleOutlined />}
            />
          </Card>
        </Col>
        <Col span={8}>
          <Card>
            <Statistic
              title="成功"
              value={completedCount}
              valueStyle={{ color: '#52c41a' }}
              prefix={<CheckCircleOutlined />}
            />
          </Card>
        </Col>
        <Col span={8}>
          <Card>
            <Statistic
              title="失败"
              value={failedCount}
              valueStyle={{ color: '#ff4d4f' }}
              prefix={<CloseCircleOutlined />}
            />
          </Card>
        </Col>
      </Row>

      {/* 执行记录列表 */}
      <Table
        columns={columns}
        dataSource={executions}
        rowKey="id"
        pagination={{ pageSize: 10 }}
        data-testid="execution-table"
      />

      {/* 执行详情抽屉 */}
      <Drawer
        title="执行详情"
        placement="right"
        width={600}
        onClose={() => setDrawerVisible(false)}
        open={drawerVisible}
      >
        {selectedExecution && (
          <Space direction="vertical" style={{ width: '100%' }} size="large">
            <Card size="small" title="基本信息">
              <p>
                <strong>执行ID:</strong> {selectedExecution.id}
              </p>
              <p>
                <strong>工作流:</strong> {selectedExecution.workflow_name}
              </p>
              <p>
                <strong>状态:</strong>{' '}
                <Tag color={getStatusColor(selectedExecution.status)}>
                  {getStatusLabel(selectedExecution.status)}
                </Tag>
              </p>
              <p>
                <strong>开始时间:</strong>{' '}
                {dayjs(selectedExecution.started_at).format('YYYY-MM-DD HH:mm:ss')}
              </p>
              {selectedExecution.completed_at && (
                <p>
                  <strong>完成时间:</strong>{' '}
                  {dayjs(selectedExecution.completed_at).format('YYYY-MM-DD HH:mm:ss')}
                </p>
              )}
              <p>
                <strong>耗时:</strong> {formatDuration(selectedExecution.duration)}
              </p>
              {selectedExecution.error && (
                <p style={{ color: '#ff4d4f' }}>
                  <strong>错误:</strong> {selectedExecution.error}
                </p>
              )}
            </Card>

            {selectedExecution.steps && selectedExecution.steps.length > 0 && (
              <Card size="small" title="执行步骤">
                <ExecutionTimeline
                  steps={selectedExecution.steps}
                  onRetry={(stepId) => console.log('Retry step:', stepId)}
                />
              </Card>
            )}

            {selectedExecution.status === 'failed' && (
              <Button
                type="primary"
                icon={<RedoOutlined />}
                onClick={() => handleRetry(selectedExecution.id)}
                block
              >
                重试执行
              </Button>
            )}
          </Space>
        )}
      </Drawer>
    </PageContainer>
  );
};

export default ExecutionHistory;
