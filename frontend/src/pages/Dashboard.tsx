import React, { useEffect, useState } from 'react';
import { Card, Row, Col, Statistic, List, Tag, Spin, Progress } from 'antd';
import { 
  TeamOutlined, 
  MessageOutlined, 
  CheckCircleOutlined,
  ClockCircleOutlined,
  NodeIndexOutlined,
  ArrowUpOutlined
} from '@ant-design/icons';
import { dashboardApi } from '../api/dashboard';
import { taskApi } from '../api/task';
import { Link } from 'react-router-dom';
import dayjs from 'dayjs';

interface Task {
  id: string;
  title: string;
  status: string;
  priority: string;
  due_date: string;
}

const Dashboard: React.FC = () => {
  const [loading, setLoading] = useState(true);
  const [stats, setStats] = useState<any>({});
  const [recentTasks, setRecentTasks] = useState<Task[]>([]);

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    setLoading(true);
    try {
      // 获取统计数据
      const statsRes = await dashboardApi.getStats();
      if (statsRes.code === 0) {
        setStats(statsRes.data);
      }

      // 获取最近任务
      const tasksRes = await taskApi.list(1, 5);
      if (tasksRes.code === 0) {
        setRecentTasks(tasksRes.data.list || []);
      }


    } finally {
      setLoading(false);
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'pending': return 'default';
      case 'claimed': return 'processing';
      case 'completed': return 'success';
      case 'cancelled': return 'error';
      default: return 'default';
    }
  };

  const getStatusText = (status: string) => {
    switch (status) {
      case 'pending': return '待处理';
      case 'claimed': return '已认领';
      case 'completed': return '已完成';
      case 'cancelled': return '已取消';
      default: return status;
    }
  };

  const getPriorityColor = (priority: string) => {
    switch (priority) {
      case 'high': return 'red';
      case 'medium': return 'orange';
      case 'low': return 'green';
      default: return 'default';
    }
  };

  const getPriorityText = (priority: string) => {
    switch (priority) {
      case 'high': return '高';
      case 'medium': return '中';
      case 'low': return '低';
      default: return priority;
    }
  };

  if (loading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100%' }}>
        <Spin size="large" />
      </div>
    );
  }

  return (
    <div>
      <h1 style={{ marginBottom: 24 }}>仪表盘</h1>
      
      {/* 统计卡片 */}
      <Row gutter={16} style={{ marginBottom: 24 }}>
        <Col span={6}>
          <Card>
            <Statistic
              title="员工总数"
              value={stats.employee_count || 0}
              prefix={<TeamOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="频道数量"
              value={stats.channel_count || 0}
              prefix={<MessageOutlined />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="待处理任务"
              value={stats.pending_task_count || 0}
              prefix={<ClockCircleOutlined />}
              valueStyle={{ color: '#faad14' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="活跃工作流"
              value={stats.active_workflow_count || 0}
              prefix={<NodeIndexOutlined />}
              valueStyle={{ color: '#722ed1' }}
            />
          </Card>
        </Col>
      </Row>

      {/* 任务统计 */}
      <Row gutter={16} style={{ marginBottom: 24 }}>
        <Col span={12}>
          <Card title="任务概览">
            <Row gutter={16}>
              <Col span={8}>
                <Statistic
                  title="总任务"
                  value={stats.total_tasks || 0}
                />
              </Col>
              <Col span={8}>
                <Statistic
                  title="已完成"
                  value={stats.completed_tasks || 0}
                  valueStyle={{ color: '#52c41a' }}
                />
              </Col>
              <Col span={8}>
                <Statistic
                  title="完成率"
                  value={stats.task_completion_rate || 0}
                  suffix="%"
                  valueStyle={{ color: '#1890ff' }}
                />
              </Col>
            </Row>
            <div style={{ marginTop: 16 }}>
              <Progress
                percent={stats.task_completion_rate || 0}
                status="active"
                strokeColor={{ from: '#108ee9', to: '#87d068' }}
              />
            </div>
          </Card>
        </Col>
        <Col span={12}>
          <Card title="今日动态">
            <Row gutter={16}>
              <Col span={12}>
                <Statistic
                  title="今日新增任务"
                  value={stats.today_new_tasks || 0}
                  prefix={<ArrowUpOutlined style={{ color: '#52c41a' }} />}
                />
              </Col>
              <Col span={12}>
                <Statistic
                  title="今日完成"
                  value={stats.today_completed_tasks || 0}
                  prefix={<CheckCircleOutlined style={{ color: '#1890ff' }} />}
                />
              </Col>
            </Row>
          </Card>
        </Col>
      </Row>

      {/* 最近任务和工作流 */}
      <Row gutter={16}>
        <Col span={12}>
          <Card 
            title="最近任务" 
            extra={<Link to="/tasks">查看全部</Link>}
          >
            <List
              dataSource={recentTasks}
              renderItem={(task) => (
                <List.Item
                  key={task.id}
                  actions={[
                    <Tag color={getPriorityColor(task.priority)}>
                      {getPriorityText(task.priority)}
                    </Tag>,
                    <Tag color={getStatusColor(task.status)}>
                      {getStatusText(task.status)}
                    </Tag>,
                  ]}
                >
                  <List.Item.Meta
                    title={<Link to={`/tasks`}>{task.title}</Link>}
                    description={task.due_date ? `截止: ${dayjs(task.due_date).format('MM-DD')}` : '无截止日期'}
                  />
                </List.Item>
              )}
            />
          </Card>
        </Col>
        <Col span={12}>
          <Card 
            title="工作流状态" 
            extra={<Link to="/workflows">查看全部</Link>}
          >
            <Row gutter={[16, 16]}>
              <Col span={12}>
                <Card size="small">
                  <Statistic
                    title="运行中"
                    value={stats.running_workflows || 0}
                    valueStyle={{ color: '#52c41a' }}
                  />
                </Card>
              </Col>
              <Col span={12}>
                <Card size="small">
                  <Statistic
                    title="已暂停"
                    value={stats.paused_workflows || 0}
                    valueStyle={{ color: '#faad14' }}
                  />
                </Card>
              </Col>
              <Col span={12}>
                <Card size="small">
                  <Statistic
                    title="失败"
                    value={stats.failed_workflows || 0}
                    valueStyle={{ color: '#f5222d' }}
                  />
                </Card>
              </Col>
              <Col span={12}>
                <Card size="small">
                  <Statistic
                    title="总执行次数"
                    value={stats.total_executions || 0}
                    valueStyle={{ color: '#1890ff' }}
                  />
                </Card>
              </Col>
            </Row>
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default Dashboard;
