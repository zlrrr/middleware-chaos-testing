import { useEffect } from 'react';
import { Card, Row, Col, Statistic, Table } from 'antd';
import { CheckCircleOutlined, ClockCircleOutlined, CloseCircleOutlined } from '@ant-design/icons';
import { useTaskStore } from '../stores/taskStore';
import type { ColumnsType } from 'antd/es/table';
import type { Task } from '../types';
import dayjs from 'dayjs';

export const Dashboard: React.FC = () => {
  const { tasks, fetchTasks } = useTaskStore();

  useEffect(() => {
    fetchTasks({ page: 1, page_size: 10 });
  }, [fetchTasks]);

  const completedTasks = tasks.filter((t) => t.status === 'completed').length;
  const runningTasks = tasks.filter((t) => t.status === 'running').length;
  const failedTasks = tasks.filter((t) => t.status === 'failed').length;

  const recentTasks = tasks.slice(0, 5);

  const columns: ColumnsType<Task> = [
    {
      title: 'Task ID',
      dataIndex: 'id',
      key: 'id',
      render: (id: string) => id.slice(0, 8),
    },
    {
      title: 'Middleware',
      dataIndex: 'middleware',
      key: 'middleware',
      render: (middleware: string) => middleware.toUpperCase(),
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
    },
    {
      title: 'Created',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (time: string) => dayjs(time).format('YYYY-MM-DD HH:mm'),
    },
  ];

  return (
    <div>
      <h2>Dashboard</h2>

      <Row gutter={[16, 16]} style={{ marginTop: 24 }}>
        <Col xs={24} sm={8}>
          <Card>
            <Statistic
              title="Completed Tasks"
              value={completedTasks}
              prefix={<CheckCircleOutlined />}
              valueStyle={{ color: '#3f8600' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={8}>
          <Card>
            <Statistic
              title="Running Tasks"
              value={runningTasks}
              prefix={<ClockCircleOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={8}>
          <Card>
            <Statistic
              title="Failed Tasks"
              value={failedTasks}
              prefix={<CloseCircleOutlined />}
              valueStyle={{ color: '#cf1322' }}
            />
          </Card>
        </Col>
      </Row>

      <Card title="Recent Tasks" style={{ marginTop: 24 }}>
        <Table
          columns={columns}
          dataSource={recentTasks}
          rowKey="id"
          pagination={false}
        />
      </Card>
    </div>
  );
};
