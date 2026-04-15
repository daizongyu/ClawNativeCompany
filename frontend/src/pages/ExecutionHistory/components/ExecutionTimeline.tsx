import React from 'react';
import { Timeline, Tag, Card, Space, Typography, Button } from 'antd';
import {
  CheckCircleOutlined,
  CloseCircleOutlined,
  ClockCircleOutlined,
  LoadingOutlined,
  PlayCircleOutlined,
  QuestionCircleOutlined,
} from '@ant-design/icons';
import dayjs from 'dayjs';

const { Text } = Typography;

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

interface ExecutionTimelineProps {
  steps: ExecutionStep[];
  onRetry?: (stepId: string) => void;
}

const getStatusIcon = (status: string) => {
  switch (status) {
    case 'completed':
      return <CheckCircleOutlined style={{ color: '#52c41a' }} />;
    case 'failed':
      return <CloseCircleOutlined style={{ color: '#ff4d4f' }} />;
    case 'running':
      return <LoadingOutlined style={{ color: '#1890ff' }} />;
    case 'pending':
      return <ClockCircleOutlined style={{ color: '#bfbfbf' }} />;
    case 'skipped':
      return <QuestionCircleOutlined style={{ color: '#faad14' }} />;
    default:
      return <PlayCircleOutlined />;
  }
};

const getStatusColor = (status: string): string => {
  const colorMap: Record<string, string> = {
    completed: 'success',
    failed: 'error',
    running: 'processing',
    pending: 'default',
    skipped: 'warning',
  };
  return colorMap[status] || 'default';
};

const getStatusLabel = (status: string): string => {
  const labelMap: Record<string, string> = {
    completed: '已完成',
    failed: '失败',
    running: '执行中',
    pending: '等待中',
    skipped: '已跳过',
  };
  return labelMap[status] || status;
};

const getStepTypeLabel = (type: string): string => {
  const labelMap: Record<string, string> = {
    start: '开始',
    end: '结束',
    condition: '条件判断',
    action: '执行动作',
    notification: '发送通知',
  };
  return labelMap[type] || type;
};

export const ExecutionTimeline: React.FC<ExecutionTimelineProps> = ({
  steps,
  onRetry,
}) => {
  const getDuration = (step: ExecutionStep): string => {
    if (!step.started_at) return '-';
    const end = step.completed_at ? dayjs(step.completed_at) : dayjs();
    const duration = end.diff(dayjs(step.started_at), 'milliseconds');
    if (duration < 1000) return `${duration}ms`;
    if (duration < 60000) return `${Math.round(duration / 1000)}s`;
    return `${Math.round(duration / 60000)}m`;
  };

  const timelineItems = steps.map((step) => ({
    key: step.id,
    dot: getStatusIcon(step.status),
    color: getStatusColor(step.status),
    children: (
      <Card size="small" style={{ marginBottom: 8 }}>
        <Space direction="vertical" style={{ width: '100%' }}>
          <Space>
            <Text strong>{step.step_name}</Text>
            <Tag color={getStatusColor(step.status)}>
              {getStatusLabel(step.status)}
            </Tag>
            <Tag>{getStepTypeLabel(step.step_type)}</Tag>
          </Space>

          <Space>
            <Text type="secondary">
              耗时: {getDuration(step)}
            </Text>
            {step.started_at && (
              <Text type="secondary">
                开始: {dayjs(step.started_at).format('HH:mm:ss')}
              </Text>
            )}
          </Space>

          {step.input && (
            <div>
              <Text type="secondary">输入:</Text>
              <pre style={{ fontSize: '12px', background: '#f5f5f5', padding: 8, borderRadius: 4 }}>
                {JSON.stringify(step.input, null, 2)}
              </pre>
            </div>
          )}

          {step.output && (
            <div>
              <Text type="secondary">输出:</Text>
              <pre style={{ fontSize: '12px', background: '#f5f5f5', padding: 8, borderRadius: 4 }}>
                {JSON.stringify(step.output, null, 2)}
              </pre>
            </div>
          )}

          {step.error && (
            <div style={{ color: '#ff4d4f' }}>
              <Text type="danger">错误: {step.error}</Text>
            </div>
          )}

          {step.status === 'failed' && onRetry && (
            <Button
              type="link"
              size="small"
              onClick={() => onRetry(step.step_id)}
            >
              重试此步骤
            </Button>
          )}
        </Space>
      </Card>
    ),
  }));

  return (
    <div data-testid="execution-timeline">
      <Timeline mode="left" items={timelineItems} />
    </div>
  );
};

export default ExecutionTimeline;
