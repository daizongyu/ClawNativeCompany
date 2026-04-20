import React, { useState, useEffect } from 'react';
import {
  Card,
  Button,
  Tag,
  Modal,
  Form,
  Input,
  Select,
  Switch,
  Alert,
  Tooltip,
  Row,
  Col,
  Divider,
  message,
} from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  ReloadOutlined,
  SettingOutlined,
} from '@ant-design/icons';
import { gatewayConfigApi, GatewayConfig, GatewayType, GatewayStatus } from '../../services/gatewayConfigApi';

interface GatewayConfigPanelProps {
  employeeId: string;
}

const GatewayConfigPanel: React.FC<GatewayConfigPanelProps> = ({ employeeId: _employeeId }) => {
  const [configs, setConfigs] = useState<GatewayConfig[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingConfig, setEditingConfig] = useState<GatewayConfig | null>(null);
  const [form] = Form.useForm();

  // 加载配置列表
  const loadConfigs = async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await gatewayConfigApi.list();
      setConfigs(res.data?.list || []);
    } catch (err: any) {
      setError(err.message || '加载配置失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadConfigs();
  }, []);

  // 打开创建对话框
  const handleOpenCreate = () => {
    setEditingConfig(null);
    form.resetFields();
    form.setFieldsValue({ type: 'dingtalk', is_default: false });
    setIsModalOpen(true);
  };

  // 打开编辑对话框
  const handleOpenEdit = (config: GatewayConfig) => {
    setEditingConfig(config);
    form.setFieldsValue({
      ...config,
      // 敏感字段不显示，需要重新输入
      app_secret: '',
      bot_token: '',
      auth_token: '',
    });
    setIsModalOpen(true);
  };

  // 关闭对话框
  const handleCloseModal = () => {
    setIsModalOpen(false);
    setEditingConfig(null);
    form.resetFields();
  };

  // 保存配置
  const handleSave = async (values: any) => {
    try {
      if (editingConfig) {
        await gatewayConfigApi.update(editingConfig.id, values);
        message.success('配置已更新');
      } else {
        await gatewayConfigApi.create(values);
        message.success('配置已创建');
      }
      handleCloseModal();
      loadConfigs();
    } catch (err: any) {
      message.error(err.message || '保存失败');
    }
  };

  // 删除配置
  const handleDelete = async (id: string) => {
    Modal.confirm({
      title: '确认删除',
      content: '确定要删除此配置吗？',
      onOk: async () => {
        try {
          await gatewayConfigApi.delete(id);
          message.success('配置已删除');
          loadConfigs();
        } catch (err: any) {
          message.error(err.message || '删除失败');
        }
      },
    });
  };

  // 验证配置
  const handleVerify = async (id: string) => {
    try {
      await gatewayConfigApi.verify(id);
      message.success('验证成功！');
      loadConfigs();
    } catch (err: any) {
      message.error(err.message || '验证失败');
    }
  };

  // 发送测试消息
  const handleTest = async (id: string) => {
    try {
      await gatewayConfigApi.test(id);
      message.success('测试消息已发送！');
    } catch (err: any) {
      message.error(err.message || '发送失败');
    }
  };

  // 设为默认
  const handleSetDefault = async (id: string) => {
    try {
      await gatewayConfigApi.setDefault(id);
      message.success('已设为默认配置');
      loadConfigs();
    } catch (err: any) {
      message.error(err.message || '设置失败');
    }
  };

  // 获取类型标签
  const getTypeLabel = (type: GatewayType) => {
    switch (type) {
      case 'dingtalk':
        return '钉钉';
      case 'slack':
        return 'Slack';
      case 'custom':
        return '自定义';
      default:
        return type;
    }
  };

  // 获取状态标签
  const getStatusTag = (status: GatewayStatus) => {
    switch (status) {
      case 'active':
        return <Tag color="success" icon={<CheckCircleOutlined />}>正常</Tag>;
      case 'inactive':
        return <Tag>禁用</Tag>;
      case 'error':
        return <Tag color="error" icon={<CloseCircleOutlined />}>错误</Tag>;
      default:
        return <Tag>{status}</Tag>;
    }
  };

  return (
    <div data-testid="gateway-config-panel">
      {/* 标题和操作按钮 */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
        <h3>
          <SettingOutlined style={{ marginRight: 8 }} />
          消息推送配置
        </h3>
        <Button
          type="primary"
          icon={<PlusOutlined />}
          onClick={handleOpenCreate}
          data-testid="gateway-config-add-btn"
        >
          添加配置
        </Button>
      </div>

      {/* 错误提示 */}
      {error && (
        <Alert
          message={error}
          type="error"
          closable
          onClose={() => setError(null)}
          style={{ marginBottom: 16 }}
        />
      )}

      {/* 配置列表 */}
      <Row gutter={[16, 16]}>
        {configs.map((config) => (
          <Col xs={24} md={12} key={config.id}>
            <Card
              bordered
              style={{
                borderColor: config.is_default ? '#1890ff' : undefined,
                position: 'relative',
              }}
              data-testid={`gateway-config-card-${config.id}`}
              title={
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <span>{config.name}</span>
                  {config.is_default && <Tag color="blue">默认</Tag>}
                </div>
              }
              extra={getStatusTag(config.status)}
            >
              <p style={{ color: '#666', marginBottom: 16 }}>
                {getTypeLabel(config.type)}
              </p>

              <div style={{ display: 'flex', gap: 8 }}>
                <Tooltip title="编辑">
                  <Button
                    size="small"
                    icon={<EditOutlined />}
                    onClick={() => handleOpenEdit(config)}
                    data-testid={`gateway-config-edit-${config.id}`}
                  />
                </Tooltip>
                <Tooltip title="验证">
                  <Button
                    size="small"
                    icon={<ReloadOutlined />}
                    onClick={() => handleVerify(config.id)}
                    data-testid={`gateway-config-verify-${config.id}`}
                  />
                </Tooltip>
                <Tooltip title="测试">
                  <Button
                    size="small"
                    icon={<CheckCircleOutlined />}
                    onClick={() => handleTest(config.id)}
                    data-testid={`gateway-config-test-${config.id}`}
                  />
                </Tooltip>
                {!config.is_default && (
                  <Tooltip title="设为默认">
                    <Button
                      size="small"
                      icon={<SettingOutlined />}
                      onClick={() => handleSetDefault(config.id)}
                      data-testid={`gateway-config-default-${config.id}`}
                    />
                  </Tooltip>
                )}
                <Tooltip title="删除">
                  <Button
                    size="small"
                    danger
                    icon={<DeleteOutlined />}
                    onClick={() => handleDelete(config.id)}
                    data-testid={`gateway-config-delete-${config.id}`}
                  />
                </Tooltip>
              </div>
            </Card>
          </Col>
        ))}
      </Row>

      {configs.length === 0 && !loading && (
        <div style={{ textAlign: 'center', padding: '40px 0', color: '#999' }}>
          <p>暂无配置</p>
          <p>点击上方按钮添加消息推送配置</p>
        </div>
      )}

      {/* 创建/编辑对话框 */}
      <Modal
        title={editingConfig ? '编辑配置' : '添加配置'}
        open={isModalOpen}
        onCancel={handleCloseModal}
        onOk={() => form.submit()}
        width={600}
        data-testid="gateway-config-dialog"
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSave}
        >
          {/* 配置名称 */}
          <Form.Item
            name="name"
            label="配置名称"
            rules={[{ required: true, message: '请输入配置名称' }]}
          >
            <Input data-testid="gateway-config-name-input" />
          </Form.Item>

          {/* 配置类型 */}
          <Form.Item
            name="type"
            label="配置类型"
            rules={[{ required: true, message: '请选择配置类型' }]}
          >
            <Select disabled={!!editingConfig} data-testid="gateway-config-type-select">
              <Select.Option value="dingtalk">钉钉</Select.Option>
              <Select.Option value="slack">Slack</Select.Option>
              <Select.Option value="custom">自定义 Webhook</Select.Option>
            </Select>
          </Form.Item>

          {/* 设为默认 */}
          <Form.Item
            name="is_default"
            valuePropName="checked"
          >
            <Switch data-testid="gateway-config-default-switch" />
            <span style={{ marginLeft: 8 }}>设为默认配置</span>
          </Form.Item>

          <Divider />

          {/* 动态表单字段 - 根据类型显示 */}
          <Form.Item noStyle shouldUpdate={(prev, curr) => prev.type !== curr.type}>
            {({ getFieldValue }) => {
              const type = getFieldValue('type');

              if (type === 'dingtalk') {
                return (
                  <div data-testid="gateway-config-dingtalk-fields">
                    <Form.Item
                      name="app_key"
                      label="AppKey"
                      rules={[{ required: true, message: '请输入 AppKey' }]}
                    >
                      <Input data-testid="gateway-config-appkey-input" />
                    </Form.Item>
                    <Form.Item
                      name="app_secret"
                      label={editingConfig ? 'AppSecret (留空保持不变)' : 'AppSecret'}
                      rules={editingConfig ? [] : [{ required: true, message: '请输入 AppSecret' }]}
                    >
                      <Input.Password data-testid="gateway-config-appsecret-input" />
                    </Form.Item>
                    <Form.Item
                      name="agent_id"
                      label="AgentID"
                    >
                      <Input data-testid="gateway-config-agentid-input" />
                    </Form.Item>
                  </div>
                );
              }

              if (type === 'slack') {
                return (
                  <div data-testid="gateway-config-slack-fields">
                    <Form.Item
                      name="webhook_url"
                      label="Webhook URL"
                      rules={[{ required: true, message: '请输入 Webhook URL' }]}
                    >
                      <Input data-testid="gateway-config-webhook-input" />
                    </Form.Item>
                    <Form.Item
                      name="bot_token"
                      label={editingConfig ? 'Bot Token (留空保持不变)' : 'Bot Token'}
                    >
                      <Input.Password data-testid="gateway-config-bottoken-input" />
                    </Form.Item>
                    <Form.Item
                      name="default_channel"
                      label="默认频道"
                    >
                      <Input placeholder="#general" data-testid="gateway-config-channel-input" />
                    </Form.Item>
                  </div>
                );
              }

              if (type === 'custom') {
                return (
                  <div data-testid="gateway-config-custom-fields">
                    <Form.Item
                      name="webhook_url"
                      label="Webhook URL"
                      rules={[{ required: true, message: '请输入 Webhook URL' }]}
                    >
                      <Input data-testid="gateway-config-custom-url-input" />
                    </Form.Item>
                    <Form.Item
                      name="auth_type"
                      label="认证类型"
                      initialValue="none"
                    >
                      <Select data-testid="gateway-config-authtype-select">
                        <Select.Option value="none">无认证</Select.Option>
                        <Select.Option value="bearer">Bearer Token</Select.Option>
                        <Select.Option value="basic">Basic Auth</Select.Option>
                      </Select>
                    </Form.Item>
                    <Form.Item noStyle shouldUpdate={(prev, curr) => prev.auth_type !== curr.auth_type}>
                      {({ getFieldValue: getAuthType }) => {
                        const authType = getAuthType('auth_type');
                        if (authType !== 'none') {
                          return (
                            <Form.Item
                              name="auth_token"
                              label={editingConfig ? '认证令牌 (留空保持不变)' : '认证令牌'}
                              rules={editingConfig ? [] : [{ required: true, message: '请输入认证令牌' }]}
                            >
                              <Input.Password data-testid="gateway-config-authtoken-input" />
                            </Form.Item>
                          );
                        }
                        return null;
                      }}
                    </Form.Item>
                  </div>
                );
              }

              return null;
            }}
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default GatewayConfigPanel;
