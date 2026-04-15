import React, { useEffect, useState } from 'react';
import { Button, Row, Col, Modal, Form, Input, Select, message, Empty, Space, Tabs, Statistic } from 'antd';
import { PlusOutlined, ReloadOutlined, PlayCircleOutlined, CheckCircleOutlined, PauseCircleOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { workflowApi } from '../../api/workflow';
import { PageContainer } from '../../components/common';
import { WorkflowCard } from './components/WorkflowCard';

const { TextArea } = Input;
const { Option } = Select;
const { TabPane } = Tabs;

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

const Workflows: React.FC = () => {
  const navigate = useNavigate();
  const [workflows, setWorkflows] = useState<Workflow[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [editingWorkflow, setEditingWorkflow] = useState<Workflow | null>(null);
  const [form] = Form.useForm();
  const [activeTab, setActiveTab] = useState('all');

  useEffect(() => {
    fetchWorkflows();
  }, []);

  const fetchWorkflows = async () => {
    setLoading(true);
    try {
      const res = await workflowApi.list();
      if (res.code === 0) {
        const list = res.data?.list || res.data?.items || [];
        setWorkflows(list);
      }
    } catch (error) {
      console.error('获取工作流列表失败:', error);
      message.error('获取工作流列表失败');
    } finally {
      setLoading(false);
    }
  };

  const handleCreate = () => {
    setEditingWorkflow(null);
    form.resetFields();
    setModalVisible(true);
  };

  const handleEdit = (id: string) => {
    navigate(`/workflows/editor/${id}`);
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
      message.error('删除失败');
    }
  };

  const handleToggleStatus = async (id: string, checked: boolean) => {
    try {
      const workflow = workflows.find(w => w.id === id);
      if (!workflow) return;

      const newStatus = checked ? 'active' : 'inactive';
      const res = await workflowApi.update(id, { status: newStatus });

      if (res.code === 0) {
        message.success(checked ? '工作流已激活' : '工作流已停用');
        fetchWorkflows();
      } else {
        message.error(res.message || '操作失败');
      }
    } catch (error) {
      console.error('切换状态失败:', error);
      message.error('操作失败');
    }
  };

  const handleExecute = async (id: string) => {
    try {
      const res = await workflowApi.execute(id);
      if (res.code === 0) {
        message.success('工作流执行已启动');
      } else {
        message.error(res.message || '执行失败');
      }
    } catch (error) {
      console.error('执行工作流失败:', error);
      message.error('执行失败');
    }
  };

  const handleViewExecutions = (workflowId: string) => {
    navigate(`/workflows/executions/${workflowId}`);
  };

  const handleModalOk = async () => {
    try {
      const values = await form.validateFields();

      // 构建完整的工作流数据
      const workflowData = {
        ...values,
        trigger_config: values.trigger_config || {},
        steps: values.steps || [
          { id: 'start', name: '开始', type: 'start', config: {}, order: 0 },
          { id: 'end', name: '结束', type: 'end', config: {}, order: 1 }
        ],
        status: 'active'
      };

      if (editingWorkflow) {
        const res = await workflowApi.update(editingWorkflow.id, workflowData);
        if (res.code === 0) {
          message.success('更新成功');
          setModalVisible(false);
          fetchWorkflows();
        } else {
          message.error(res.message || '更新失败');
        }
      } else {
        const res = await workflowApi.create(workflowData);
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

  // 筛选工作流
  const filteredWorkflows = workflows.filter(w => {
    if (activeTab === 'active') return w.status === 'active';
    if (activeTab === 'inactive') return w.status === 'inactive';
    return true;
  });

  // 统计
  const activeCount = workflows.filter(w => w.status === 'active').length;
  const inactiveCount = workflows.filter(w => w.status === 'inactive').length;

  return (
    <PageContainer
      data-testid="page-workflows"
      data-page="workflows"
      loading={loading}
    >
      {/* 统计卡片 */}
      <Row gutter={16} style={{ marginBottom: 24 }}>
        <Col span={8}>
          <Statistic
            title="总工作流"
            value={workflows.length}
            prefix={<PlayCircleOutlined />}
          />
        </Col>
        <Col span={8}>
          <Statistic
            title="已激活"
            value={activeCount}
            valueStyle={{ color: '#52c41a' }}
            prefix={<CheckCircleOutlined />}
          />
        </Col>
        <Col span={8}>
          <Statistic
            title="已停用"
            value={inactiveCount}
            valueStyle={{ color: '#999' }}
            prefix={<PauseCircleOutlined />}
          />
        </Col>
      </Row>

      {/* 操作栏 */}
      <Space style={{ marginBottom: 16 }}>
        <Button
          type="primary"
          icon={<PlusOutlined />}
          onClick={handleCreate}
          data-testid="workflow-create-btn"
        >
          新建工作流
        </Button>
        <Button icon={<ReloadOutlined />} onClick={fetchWorkflows}>
          刷新
        </Button>
      </Space>

      {/* 标签页 */}
      <Tabs activeKey={activeTab} onChange={setActiveTab} style={{ marginBottom: 16 }}>
        <TabPane tab="全部工作流" key="all" />
        <TabPane tab={`已激活 (${activeCount})`} key="active" />
        <TabPane tab={`已停用 (${inactiveCount})`} key="inactive" />
      </Tabs>

      {/* 工作流卡片列表 */}
      {filteredWorkflows.length > 0 ? (
        <Row gutter={[16, 16]}>
          {filteredWorkflows.map(workflow => (
            <Col xs={24} sm={12} lg={8} key={workflow.id}>
              <WorkflowCard
                workflow={workflow}
                onEdit={handleEdit}
                onDelete={handleDelete}
                onToggleStatus={handleToggleStatus}
                onExecute={handleExecute}
                onViewExecutions={handleViewExecutions}
              />
            </Col>
          ))}
        </Row>
      ) : (
        <Empty
          description="暂无工作流"
          image={Empty.PRESENTED_IMAGE_SIMPLE}
          data-testid="workflow-empty"
        />
      )}

      {/* 创建/编辑弹窗 */}
      <Modal
        title={editingWorkflow ? '编辑工作流' : '新建工作流'}
        open={modalVisible}
        onOk={handleModalOk}
        onCancel={() => setModalVisible(false)}
        width={600}
        destroyOnClose
      >
        <Form
          form={form}
          layout="vertical"
          initialValues={{ trigger_type: 'manual' }}
        >
          <Form.Item
            name="name"
            label="工作流名称"
            rules={[{ required: true, message: '请输入工作流名称' }]}
          >
            <Input placeholder="请输入工作流名称" maxLength={100} />
          </Form.Item>

          <Form.Item
            name="description"
            label="工作流描述"
          >
            <TextArea
              rows={3}
              placeholder="请输入工作流描述"
              maxLength={500}
              showCount
            />
          </Form.Item>

          <Form.Item
            name="trigger_type"
            label="触发器类型"
            rules={[{ required: true, message: '请选择触发器类型' }]}
          >
            <Select placeholder="请选择触发器类型">
              <Option value="manual">手动触发</Option>
              <Option value="keyword">关键词触发</Option>
              <Option value="schedule">定时触发</Option>
              <Option value="webhook">Webhook触发</Option>
            </Select>
          </Form.Item>

          {/* TODO: 根据触发器类型显示不同的配置选项 */}
        </Form>
      </Modal>
    </PageContainer>
  );
};

export default Workflows;
