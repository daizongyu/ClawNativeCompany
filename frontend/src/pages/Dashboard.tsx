import React, { useEffect, useState } from 'react';
import { Card, Row, Col, Statistic, List, Tag } from 'antd';
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
import { PageContainer } from '../components/common';

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

  // 设置当前页面
  useEffect(() => {
    if (typeof window !== 'undefined' && window.__CLAW_TEST__) {
      window.__CLAW_TEST__.setCurrentPage('dashboard');
    }
  }, []);

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
      const tasksRes = await taskApi.getMyTasks();
      if (tasksRes.code === 0) {
        // 后端返回的数据格式是 { list: [...], total: n, page: 1, page_size: 20, total_page: 1 }
        const taskList = tasksRes.data.list || tasksRes.data.items || [];
        setRecentTasks(taskList.slice(0, 5));
      }
    } catch (error) {
      console.error('获取数据失败:', error);
    } finally {
      setLoading(false);
    }
  };

  const getStatusColor = (status: string) => {
    const colors: Record<string, string> = {
      pending: 'default',
      in_progress: 'processing',
      completed: 'success',
      cancelled: 'error',
    };
    return colors[status] || 'default';
  };

  const getStatusText = (status: string) => {
    const texts: Record<string, string> = {
      pending: '待处理',
      in_progress: '进行中',
      completed: '已完成',
      cancelled: '已取消',
    };
    return texts[status] || status;
  };

  const getPriorityColor = (priority: string) => {
    const colors: Record<string, string> = {
      low: 'success',
      medium: 'warning',
      high: 'error',
    };
    return colors[priority] || 'default';
  };

  return (
    <PageContainer
      data-testid="page-dashboard"
      data-page="dashboard"
      loading={loading}
    >
      <div style={{ padding: '24px' }}>
        <h1 style={{ marginBottom: '24px' }}>仪表盘</h1>

        {/* 统计卡片 */}
        <Row gutter={[16, 16]} style={{ marginBottom: '24px' }}>
          <Col xs={24} sm={12} lg={6}>
            <Card data-testid="stat-employee-card">
              <Statistic
                data-testid="stat-total-employees"
                title="员工总数"
                value={stats.employee_count || 0}
                prefix={<TeamOutlined />}
                valueStyle={{ color: '#3f8600' }}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card data-testid="stat-channel-card">
              <Statistic
                data-testid="stat-total-channels"
                title="频道总数"
                value={stats.channel_count || 0}
                prefix={<MessageOutlined />}
                valueStyle={{ color: '#1890ff' }}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card data-testid="stat-task-card">
              <Statistic
                data-testid="stat-total-tasks"
                title="任务总数"
                value={stats.task_count || 0}
                prefix={<CheckCircleOutlined />}
                valueStyle={{ color: '#722ed1' }}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card data-testid="stat-workflow-card">
              <Statistic
                data-testid="stat-total-workflows"
                title="工作流总数"
                value={stats.workflow_count || 0}
                prefix={<NodeIndexOutlined />}
                valueStyle={{ color: '#fa8c16' }}
              />
            </Card>
          </Col>
        </Row>

        {/* 任务状态分布 */}
        <Row gutter={[16, 16]} style={{ marginBottom: '24px' }}>
          <Col xs={24} lg={12}>
            <Card title="任务状态分布" data-testid="task-status-card">
              <Row gutter={16}>
                <Col span={12}>
                  <Statistic
                    data-testid="task-pending"
                    title="待处理"
                    value={stats.pending_tasks || 0}
                    prefix={<ClockCircleOutlined />}
                  />
                </Col>
                <Col span={12}>
                  <Statistic
                    data-testid="task-in-progress"
                    title="进行中"
                    value={stats.in_progress_tasks || 0}
                    prefix={<ArrowUpOutlined />}
                  />
                </Col>
              </Row>
              <div style={{ marginTop: '16px' }}>
                <div style={{ textAlign: 'center', marginTop: '8px' }} data-testid="task-completion-rate">
                  任务完成率: {stats.task_completion_rate || 0}%
                </div>
              </div>
            </Card>
          </Col>
          <Col xs={24} lg={12}>
            <Card title="最近任务" data-testid="recent-tasks-card">
              <List
                data-testid="recent-tasks-section"
                dataSource={recentTasks}
                renderItem={(task) => (
                  <List.Item
                    key={task.id}
                    actions={[
                      <Tag color={getPriorityColor(task.priority)} key="priority">
                        {task.priority === 'high' ? '高' : task.priority === 'medium' ? '中' : '低'}
                      </Tag>,
                      <Tag color={getStatusColor(task.status)} key="status">
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
        </Row>
      </div>
    </PageContainer>
  );
};

export default Dashboard;
