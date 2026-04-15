import React from 'react';
import { Form, Input, Select, Card, Alert, Space } from 'antd';

const { Option } = Select;

interface TriggerConfigPanelProps {
  triggerType: string;
  triggerConfig: any;
  onChange: (config: any) => void;
}

export const TriggerConfigPanel: React.FC<TriggerConfigPanelProps> = ({
  triggerType,
  triggerConfig,
  onChange,
}) => {
  const renderConfigForm = () => {
    switch (triggerType) {
      case 'manual':
        return (
          <Alert
            message="手动触发"
            description="工作流需要手动点击执行按钮才会运行。适用于需要人工确认的场景。"
            type="info"
            showIcon
          />
        );

      case 'keyword':
        return (
          <Space direction="vertical" style={{ width: '100%' }}>
            <Alert
              message="关键词触发"
              description="当频道消息中包含指定关键词时，自动触发工作流。"
              type="info"
              showIcon
            />
            <Form.Item
              label="触发关键词"
              required
              help="支持正则表达式，如：报销|费用"
            >
              <Input
                placeholder="请输入触发关键词"
                value={triggerConfig?.keyword || ''}
                onChange={(e) => onChange({ ...triggerConfig, keyword: e.target.value })}
                data-testid="trigger-keyword-input"
              />
            </Form.Item>
          </Space>
        );

      case 'schedule':
        return (
          <Space direction="vertical" style={{ width: '100%' }}>
            <Alert
              message="定时触发"
              description="按照 Cron 表达式定时触发工作流。"
              type="info"
              showIcon
            />
            <Form.Item
              label="Cron 表达式"
              required
              help="例如：0 9 * * 1-5 表示工作日早上9点"
            >
              <Input
                placeholder="0 9 * * 1-5"
                value={triggerConfig?.schedule || ''}
                onChange={(e) => onChange({ ...triggerConfig, schedule: e.target.value })}
                data-testid="trigger-schedule-input"
              />
            </Form.Item>
          </Space>
        );

      case 'webhook':
        return (
          <Space direction="vertical" style={{ width: '100%' }}>
            <Alert
              message="Webhook 触发"
              description="通过 HTTP POST 请求触发工作流。"
              type="info"
              showIcon
            />
            <Form.Item label="Webhook URL">
              <Input
                value={triggerConfig?.webhook_url || '自动生成'}
                disabled
                data-testid="trigger-webhook-url"
              />
            </Form.Item>
          </Space>
        );

      default:
        return null;
    }
  };

  return (
    <Card title="触发器配置" data-testid="trigger-config-panel">
      <Form.Item label="触发器类型" required>
        <Select
          value={triggerType}
          onChange={(_value) => onChange({})}
          data-testid="trigger-type-select"
        >
          <Option value="manual">手动触发</Option>
          <Option value="keyword">关键词触发</Option>
          <Option value="schedule">定时触发</Option>
          <Option value="webhook">Webhook触发</Option>
        </Select>
      </Form.Item>

      {renderConfigForm()}
    </Card>
  );
};

export default TriggerConfigPanel;
