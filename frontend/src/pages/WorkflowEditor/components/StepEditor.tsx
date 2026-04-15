import React, { useState } from 'react';
import {
  Card,
  Button,
  Form,
  Input,
  Select,
  Space,
  Collapse,
  Tag,
  Popconfirm,
  Row,
  Col,
} from 'antd';
import {
  PlusOutlined,
  DeleteOutlined,
  PlayCircleOutlined,
  CheckCircleOutlined,
  QuestionCircleOutlined,
  BellOutlined,
  RobotOutlined,
} from '@ant-design/icons';

const { Option } = Select;
const { Panel } = Collapse;
const { TextArea } = Input;

interface WorkflowStep {
  id: string;
  name: string;
  type: 'start' | 'end' | 'condition' | 'action' | 'notification';
  config: any;
  order: number;
  next_step_id?: string;
}

interface StepEditorProps {
  steps: WorkflowStep[];
  onChange: (steps: WorkflowStep[]) => void;
}

const getStepIcon = (type: string) => {
  const iconMap: Record<string, React.ReactNode> = {
    start: <PlayCircleOutlined />,
    end: <CheckCircleOutlined />,
    condition: <QuestionCircleOutlined />,
    action: <RobotOutlined />,
    notification: <BellOutlined />,
  };
  return iconMap[type] || <PlayCircleOutlined />;
};

const getStepLabel = (type: string): string => {
  const labelMap: Record<string, string> = {
    start: '开始',
    end: '结束',
    condition: '条件判断',
    action: '执行动作',
    notification: '发送通知',
  };
  return labelMap[type] || type;
};

const getStepColor = (type: string): string => {
  const colorMap: Record<string, string> = {
    start: 'green',
    end: 'red',
    condition: 'orange',
    action: 'blue',
    notification: 'purple',
  };
  return colorMap[type] || 'default';
};

export const StepEditor: React.FC<StepEditorProps> = ({ steps, onChange }) => {
  const [editingStep, setEditingStep] = useState<string | null>(null);

  const generateStepId = () => `step_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;

  const addStep = (type: WorkflowStep['type']) => {
    const newStep: WorkflowStep = {
      id: generateStepId(),
      name: getStepLabel(type),
      type,
      config: {},
      order: steps.length,
    };
    onChange([...steps, newStep]);
    setEditingStep(newStep.id);
  };

  const updateStep = (stepId: string, updates: Partial<WorkflowStep>) => {
    const newSteps = steps.map((step) =>
      step.id === stepId ? { ...step, ...updates } : step
    );
    onChange(newSteps);
  };

  const deleteStep = (stepId: string) => {
    const newSteps = steps.filter((step) => step.id !== stepId);
    // 重新排序
    newSteps.forEach((step, index) => {
      step.order = index;
    });
    onChange(newSteps);
  };

  const renderStepConfig = (step: WorkflowStep) => {
    switch (step.type) {
      case 'start':
        return (
          <Form.Item label="初始输入">
            <TextArea
              placeholder="可选：定义初始输入数据（JSON格式）"
              value={step.config?.input || ''}
              onChange={(e) =>
                updateStep(step.id, { config: { ...step.config, input: e.target.value } })
              }
              rows={3}
            />
          </Form.Item>
        );

      case 'condition':
        return (
          <Space direction="vertical" style={{ width: '100%' }}>
            <Form.Item label="条件表达式" required>
              <Input
                placeholder="例如：input.amount > 1000"
                value={step.config?.expression || ''}
                onChange={(e) =>
                  updateStep(step.id, { config: { ...step.config, expression: e.target.value } })
                }
              />
            </Form.Item>
            <Form.Item label="满足条件时的下一步">
              <Select
                placeholder="选择下一步"
                value={step.config?.true_next}
                onChange={(value) =>
                  updateStep(step.id, { config: { ...step.config, true_next: value } })
                }
                allowClear
              >
                {steps
                  .filter((s) => s.id !== step.id)
                  .map((s) => (
                    <Option key={s.id} value={s.id}>
                      {s.name}
                    </Option>
                  ))}
              </Select>
            </Form.Item>
            <Form.Item label="不满足条件时的下一步">
              <Select
                placeholder="选择下一步"
                value={step.config?.false_next}
                onChange={(value) =>
                  updateStep(step.id, { config: { ...step.config, false_next: value } })
                }
                allowClear
              >
                {steps
                  .filter((s) => s.id !== step.id)
                  .map((s) => (
                    <Option key={s.id} value={s.id}>
                      {s.name}
                    </Option>
                  ))}
              </Select>
            </Form.Item>
          </Space>
        );

      case 'action':
        return (
          <Space direction="vertical" style={{ width: '100%' }}>
            <Form.Item label="动作类型" required>
              <Select
                value={step.config?.action_type}
                onChange={(value) =>
                  updateStep(step.id, { config: { ...step.config, action_type: value } })
                }
              >
                <Option value="assign_task">分配任务</Option>
                <Option value="send_message">发送消息</Option>
                <Option value="call_api">调用API</Option>
                <Option value="update_data">更新数据</Option>
              </Select>
            </Form.Item>
            <Form.Item label="动作配置">
              <TextArea
                placeholder="动作配置（JSON格式）"
                value={step.config?.action_config || ''}
                onChange={(e) =>
                  updateStep(step.id, { config: { ...step.config, action_config: e.target.value } })
                }
                rows={4}
              />
            </Form.Item>
          </Space>
        );

      case 'notification':
        return (
          <Space direction="vertical" style={{ width: '100%' }}>
            <Form.Item label="通知方式" required>
              <Select
                value={step.config?.channel}
                onChange={(value) =>
                  updateStep(step.id, { config: { ...step.config, channel: value } })
                }
              >
                <Option value="email">邮件</Option>
                <Option value="sms">短信</Option>
                <Option value="app">应用内</Option>
                <Option value="channel">频道消息</Option>
              </Select>
            </Form.Item>
            <Form.Item label="接收人">
              <Input
                placeholder="接收人ID或邮箱"
                value={step.config?.recipient || ''}
                onChange={(e) =>
                  updateStep(step.id, { config: { ...step.config, recipient: e.target.value } })
                }
              />
            </Form.Item>
            <Form.Item label="通知内容">
              <TextArea
                placeholder="通知内容模板"
                value={step.config?.content || ''}
                onChange={(e) =>
                  updateStep(step.id, { config: { ...step.config, content: e.target.value } })
                }
                rows={3}
              />
            </Form.Item>
          </Space>
        );

      case 'end':
        return (
          <Form.Item label="结束状态">
            <Select
              value={step.config?.status || 'success'}
              onChange={(value) =>
                updateStep(step.id, { config: { ...step.config, status: value } })
              }
            >
              <Option value="success">成功</Option>
              <Option value="failed">失败</Option>
              <Option value="cancelled">取消</Option>
            </Select>
          </Form.Item>
        );

      default:
        return null;
    }
  };

  return (
    <Card title="步骤配置" data-testid="step-editor">
      {/* 步骤列表 */}
      <Collapse
        activeKey={editingStep || []}
        onChange={(keys) => setEditingStep(keys[0] as string)}
      >
        {steps.map((step, index) => (
          <Panel
            key={step.id}
            header={
              <Space>
                <Tag color={getStepColor(step.type)} icon={getStepIcon(step.type)}>
                  {getStepLabel(step.type)}
                </Tag>
                <span>{step.name}</span>
                <span style={{ color: '#999', fontSize: '12px' }}>步骤 {index + 1}</span>
              </Space>
            }
            extra={
              step.type !== 'start' && step.type !== 'end' ? (
                <Popconfirm
                  title="确认删除"
                  description="删除后无法恢复，是否继续？"
                  onConfirm={() => deleteStep(step.id)}
                >
                  <Button type="text" danger icon={<DeleteOutlined />} size="small">
                    删除
                  </Button>
                </Popconfirm>
              ) : null
            }
          >
            <Form layout="vertical">
              <Form.Item label="步骤名称">
                <Input
                  value={step.name}
                  onChange={(e) => updateStep(step.id, { name: e.target.value })}
                  placeholder="步骤名称"
                />
              </Form.Item>
              {renderStepConfig(step)}
            </Form>
          </Panel>
        ))}
      </Collapse>

      {/* 添加步骤按钮 */}
      <Row gutter={8} style={{ marginTop: 16 }}>
        <Col>
          <Button type="dashed" onClick={() => addStep('condition')} icon={<PlusOutlined />}>
            添加条件
          </Button>
        </Col>
        <Col>
          <Button type="dashed" onClick={() => addStep('action')} icon={<PlusOutlined />}>
            添加动作
          </Button>
        </Col>
        <Col>
          <Button type="dashed" onClick={() => addStep('notification')} icon={<PlusOutlined />}>
            添加通知
          </Button>
        </Col>
      </Row>
    </Card>
  );
};

export default StepEditor;
