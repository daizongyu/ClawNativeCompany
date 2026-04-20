import React, { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
  Form,
  Input,
  Button,
  Steps,
  Card,
  message,
  Space,
  Row,
  Col,
} from 'antd';
import {
  SaveOutlined,
  ArrowLeftOutlined,
  PlayCircleOutlined,
  SettingOutlined,
  AppstoreOutlined,
  CheckCircleOutlined,
} from '@ant-design/icons';
import { workflowApi } from '../../services/workflow';
import { PageContainer } from '../../components/common';
import { TriggerConfigPanel } from './components/TriggerConfigPanel';
import { StepEditor } from './components/StepEditor';

const { Step } = Steps;
const { TextArea } = Input;

interface WorkflowStep {
  id: string;
  name: string;
  type: 'start' | 'end' | 'condition' | 'action' | 'notification';
  config: any;
  order: number;
  next_step_id?: string;
}

interface Workflow {
  id: string;
  name: string;
  description: string;
  status: 'active' | 'inactive';
  trigger_type: 'manual' | 'keyword' | 'webhook' | 'schedule';
  trigger_config?: any;
  steps: WorkflowStep[];
}

const WorkflowEditor: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [form] = Form.useForm();
  const [currentStep, setCurrentStep] = useState(0);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);

  // 工作流数据
  const [workflow, setWorkflow] = useState<Partial<Workflow>>({
    name: '',
    description: '',
    trigger_type: 'manual',
    trigger_config: {},
    steps: [
      { id: 'start', name: '开始', type: 'start', config: {}, order: 0 },
      { id: 'end', name: '结束', type: 'end', config: {}, order: 1 },
    ],
  });

  const isEditing = !!id;

  useEffect(() => {
    if (isEditing && id) {
      fetchWorkflow(id);
    }
  }, [id]);

  const fetchWorkflow = async (workflowId: string) => {
    setLoading(true);
    try {
      const res = await workflowApi.getById(workflowId);
      if (res.code === 0) {
        const data = res.data;
        setWorkflow({
          ...data,
          steps: data.steps || [
            { id: 'start', name: '开始', type: 'start', config: {}, order: 0 },
            { id: 'end', name: '结束', type: 'end', config: {}, order: 1 },
          ],
        });
        form.setFieldsValue({
          name: data.name,
          description: data.description,
        });
      } else {
        message.error(res.message || '获取工作流失败');
      }
    } catch (error) {
      console.error('获取工作流失败:', error);
      message.error('获取工作流失败');
    } finally {
      setLoading(false);
    }
  };

  const handleSave = async () => {
    try {
      const values = await form.validateFields();
      setSaving(true);

      const workflowData = {
        ...values,
        trigger_type: workflow.trigger_type,
        trigger_config: workflow.trigger_config,
        steps: workflow.steps,
      };

      if (isEditing) {
        const res = await workflowApi.update(id!, workflowData);
        if (res.code === 0) {
          message.success('保存成功');
        } else {
          message.error(res.message || '保存失败');
        }
      } else {
        const res = await workflowApi.create(workflowData);
        if (res.code === 0) {
          message.success('创建成功');
          navigate(`/workflows/editor/${res.data.id}`);
        } else {
          message.error(res.message || '创建失败');
        }
      }
    } catch (error) {
      console.error('保存工作流失败:', error);
    } finally {
      setSaving(false);
    }
  };

  const handleTest = async () => {
    if (!isEditing) {
      message.warning('请先保存工作流');
      return;
    }
    try {
      const res = await workflowApi.execute(id!);
      if (res.code === 0) {
        message.success('测试执行已启动');
      } else {
        message.error(res.message || '测试执行失败');
      }
    } catch (error) {
      console.error('测试执行失败:', error);
      message.error('测试执行失败');
    }
  };

  const steps = [
    {
      title: '基本信息',
      icon: <SettingOutlined />,
      content: (
        <Card>
          <Form
            form={form}
            layout="vertical"
            initialValues={{ name: workflow.name, description: workflow.description }}
          >
            <Form.Item
              name="name"
              label="工作流名称"
              rules={[{ required: true, message: '请输入工作流名称' }]}
            >
              <Input placeholder="请输入工作流名称" maxLength={100} />
            </Form.Item>

            <Form.Item name="description" label="工作流描述">
              <TextArea
                rows={4}
                placeholder="请输入工作流描述"
                maxLength={500}
                showCount
              />
            </Form.Item>
          </Form>
        </Card>
      ),
    },
    {
      title: '触发器配置',
      icon: <PlayCircleOutlined />,
      content: (
        <TriggerConfigPanel
          triggerType={workflow.trigger_type || 'manual'}
          triggerConfig={workflow.trigger_config || {}}
          onChange={(config) => setWorkflow({ ...workflow, trigger_config: config })}
        />
      ),
    },
    {
      title: '步骤配置',
      icon: <AppstoreOutlined />,
      content: (
        <StepEditor
          steps={workflow.steps || []}
          onChange={(steps) => setWorkflow({ ...workflow, steps })}
        />
      ),
    },
    {
      title: '保存发布',
      icon: <CheckCircleOutlined />,
      content: (
        <Card>
          <div style={{ textAlign: 'center', padding: '40px 0' }}>
            <h3>工作流配置完成</h3>
            <p>点击下方按钮保存工作流</p>
            <Space>
              <Button type="primary" icon={<SaveOutlined />} onClick={handleSave} loading={saving}>
                {isEditing ? '保存修改' : '创建工作流'}
              </Button>
              {isEditing && (
                <Button icon={<PlayCircleOutlined />} onClick={handleTest}>
                  测试执行
                </Button>
              )}
            </Space>
          </div>
        </Card>
      ),
    },
  ];

  return (
    <PageContainer
      data-testid="page-workflow-editor"
      data-page="workflow-editor"
      loading={loading}
    >
      {/* 顶部操作栏 */}
      <Row justify="space-between" style={{ marginBottom: 24 }}>
        <Col>
          <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/workflows')}>
            返回列表
          </Button>
        </Col>
        <Col>
          <Space>
            <Button onClick={() => setCurrentStep(Math.max(0, currentStep - 1))}>
              上一步
            </Button>
            <Button
              type="primary"
              onClick={() => setCurrentStep(Math.min(steps.length - 1, currentStep + 1))}
            >
              下一步
            </Button>
          </Space>
        </Col>
      </Row>

      {/* 步骤导航 */}
      <Steps current={currentStep} style={{ marginBottom: 24 }}>
        {steps.map((item) => (
          <Step key={item.title} title={item.title} icon={item.icon} />
        ))}
      </Steps>

      {/* 步骤内容 */}
      <div style={{ minHeight: 400 }}>{steps[currentStep].content}</div>

      {/* 底部操作栏 */}
      <Row justify="center" style={{ marginTop: 24 }}>
        <Col>
          <Space>
            <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/workflows')}>
              取消
            </Button>
            <Button type="primary" icon={<SaveOutlined />} onClick={handleSave} loading={saving}>
              {isEditing ? '保存修改' : '创建工作流'}
            </Button>
            {isEditing && (
              <Button icon={<PlayCircleOutlined />} onClick={handleTest}>
                测试执行
              </Button>
            )}
          </Space>
        </Col>
      </Row>
    </PageContainer>
  );
};

export default WorkflowEditor;
