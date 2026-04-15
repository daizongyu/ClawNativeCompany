import React from 'react';
import { Card, Tag, Space, Switch, Button, Popconfirm, Tooltip } from 'antd';
import {
  EditOutlined,
  DeleteOutlined,
  PlayCircleOutlined,
  ClockCircleOutlined,
  CommentOutlined,
  GlobalOutlined,
  HistoryOutlined,
} from '@ant-design/icons';
import dayjs from 'dayjs';

interface Workflow {
  id: string;
  name: string;
  description: string;
  status: 'active' | 'inactive';
  trigger_type: 'manual' | 'keyword' | 'webhook' | 'schedule';
  trigger_config?: {
    keyword?: string;
    schedule?: string;
  };
  updated_at: string;
  created_at: string;
}

interface WorkflowCardProps {
  workflow: Workflow;
  onEdit: (id: string) => void;
  onDelete: (id: string) => void;
  onToggleStatus: (id: string, status: boolean) => void;
  onExecute: (id: string) => void;
  onViewExecutions: (id: string) => void;
}

const getStatusColor = (status: string): string => {
  const colorMap: Record<string, string> = {
    active: 'success',
    inactive: 'default',
  };
  return colorMap[status] || 'default';
};

const getStatusLabel = (status: string): string => {
  const labelMap: Record<string, string> = {
    active: '已激活',
    inactive: '已停用',
  };
  return labelMap[status] || status;
};

const getTriggerIcon = (triggerType: string) => {
  const iconMap: Record<string, React.ReactNode> = {
    manual: <PlayCircleOutlined />,
    keyword: <CommentOutlined />,
    schedule: <ClockCircleOutlined />,
    webhook: <GlobalOutlined />,
  };
  return iconMap[triggerType] || <PlayCircleOutlined />;
};

const getTriggerLabel = (triggerType: string): string => {
  const labelMap: Record<string, string> = {
    manual: '手动触发',
    keyword: '关键词触发',
    schedule: '定时触发',
    webhook: 'Webhook触发',
  };
  return labelMap[triggerType] || triggerType;
};

const getTriggerDescription = (workflow: Workflow): string => {
  switch (workflow.trigger_type) {
    case 'keyword':
      return workflow.trigger_config?.keyword || '未配置关键词';
    case 'schedule':
      return workflow.trigger_config?.schedule || '未配置定时规则';
    case 'webhook':
      return '通过 Webhook URL 触发';
    default:
      return '点击执行按钮手动触发';
  }
};

export const WorkflowCard: React.FC<WorkflowCardProps> = ({
  workflow,
  onEdit,
  onDelete,
  onToggleStatus,
  onExecute,
  onViewExecutions,
}) => {
  return (
    <Card
      className="workflow-card"
      data-testid={`workflow-card-${workflow.id}`}
      hoverable
      actions={[
        <Tooltip title="编辑工作流" key="edit">
          <Button
            type="text"
            icon={<EditOutlined />}
            onClick={() => onEdit(workflow.id)}
            data-testid={`workflow-edit-btn-${workflow.id}`}
          >
            编辑
          </Button>
        </Tooltip>,
        <Tooltip title="查看执行记录" key="executions">
          <Button
            type="text"
            icon={<HistoryOutlined />}
            onClick={() => onViewExecutions(workflow.id)}
            data-testid={`workflow-executions-btn-${workflow.id}`}
          >
            执行记录
          </Button>
        </Tooltip>,
        <Popconfirm
          key="delete"
          title="确认删除"
          description="删除后无法恢复，是否继续？"
          onConfirm={() => onDelete(workflow.id)}
        >
          <Button
            type="text"
            danger
            icon={<DeleteOutlined />}
            data-testid={`workflow-delete-btn-${workflow.id}`}
          >
            删除
          </Button>
        </Popconfirm>,
      ]}
    >
      <Card.Meta
        title={
          <Space>
            <span data-testid={`workflow-name-${workflow.id}`}>
              {workflow.name}
            </span>
            <Tag
              color={getStatusColor(workflow.status)}
              data-testid={`workflow-status-${workflow.id}`}
            >
              {getStatusLabel(workflow.status)}
            </Tag>
          </Space>
        }
        description={
          <Space direction="vertical" size="small" style={{ width: '100%' }}>
            <div
              className="workflow-description"
              style={{
                color: 'rgba(0, 0, 0, 0.45)',
                minHeight: '40px',
              }}
              data-testid={`workflow-desc-${workflow.id}`}
            >
              {workflow.description || '暂无描述'}
            </div>

            <Space wrap>
              <Tag
                icon={getTriggerIcon(workflow.trigger_type)}
                data-testid={`workflow-trigger-${workflow.id}`}
              >
                {getTriggerLabel(workflow.trigger_type)}
              </Tag>
              <Tooltip title={getTriggerDescription(workflow)}>
                <span style={{ color: 'rgba(0, 0, 0, 0.45)', fontSize: '12px' }}>
                  {workflow.trigger_type === 'keyword' && workflow.trigger_config?.keyword
                    ? `关键词: ${workflow.trigger_config.keyword}`
                    : workflow.trigger_type === 'schedule' && workflow.trigger_config?.schedule
                    ? `Cron: ${workflow.trigger_config.schedule}`
                    : getTriggerLabel(workflow.trigger_type)}
                </span>
              </Tooltip>
            </Space>

            <Space
              style={{
                width: '100%',
                justifyContent: 'space-between',
                marginTop: '12px',
                paddingTop: '12px',
                borderTop: '1px solid #f0f0f0',
              }}
            >
              <span
                style={{ color: 'rgba(0, 0, 0, 0.45)', fontSize: '12px' }}
                data-testid={`workflow-updated-${workflow.id}`}
              >
                更新于 {dayjs(workflow.updated_at).format('YYYY-MM-DD HH:mm')}
              </span>

              <Space>
                {workflow.status === 'active' && (
                  <Button
                    type="primary"
                    size="small"
                    icon={<PlayCircleOutlined />}
                    onClick={() => onExecute(workflow.id)}
                    data-testid={`workflow-execute-btn-${workflow.id}`}
                  >
                    执行
                  </Button>
                )}
                <Switch
                  checked={workflow.status === 'active'}
                  onChange={(checked) => onToggleStatus(workflow.id, checked)}
                  checkedChildren="激活"
                  unCheckedChildren="停用"
                  data-testid={`workflow-toggle-${workflow.id}`}
                />
              </Space>
            </Space>
          </Space>
        }
      />
    </Card>
  );
};

export default WorkflowCard;
