import { useEffect, useState } from 'react';
import { Button, Table, Tag, Space, Popconfirm, message, Select } from 'antd';
import { PlusOutlined, DeleteOutlined, PlayCircleOutlined, EyeOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import type { ColumnsType } from 'antd/es/table';
import dayjs from 'dayjs';
import { useTaskStore } from '../stores/taskStore';
import { CreateTaskModal } from '../components/CreateTaskModal';
import type { Task } from '../types';

const getStatusColor = (status: string) => {
  switch (status) {
    case 'pending':
      return 'default';
    case 'running':
      return 'processing';
    case 'completed':
      return 'success';
    case 'failed':
      return 'error';
    default:
      return 'default';
  }
};

export const Tasks: React.FC = () => {
  const navigate = useNavigate();
  const [modalVisible, setModalVisible] = useState(false);
  const [statusFilter, setStatusFilter] = useState<string | undefined>();

  const {
    tasks,
    loading,
    total,
    page,
    pageSize,
    fetchTasks,
    deleteTask,
    runTask,
  } = useTaskStore();

  useEffect(() => {
    fetchTasks({ page: 1, page_size: 10, status: statusFilter as any });
  }, [fetchTasks, statusFilter]);

  const handleDelete = async (taskId: string) => {
    try {
      await deleteTask(taskId);
      message.success('Task deleted successfully');
    } catch (error: any) {
      message.error(error.message || 'Failed to delete task');
    }
  };

  const handleRun = async (taskId: string) => {
    try {
      await runTask(taskId);
      message.success('Task started successfully');
    } catch (error: any) {
      message.error(error.message || 'Failed to start task');
    }
  };

  const handleViewResult = (taskId: string) => {
    navigate(`/tasks/${taskId}/result`);
  };

  const columns: ColumnsType<Task> = [
    {
      title: 'Task ID',
      dataIndex: 'id',
      key: 'id',
      width: 250,
      render: (id: string) => id.slice(0, 8),
    },
    {
      title: 'Middleware',
      dataIndex: 'middleware',
      key: 'middleware',
      width: 120,
      render: (middleware: string) => middleware.toUpperCase(),
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      width: 120,
      render: (status: string) => (
        <Tag color={getStatusColor(status)}>{status.toUpperCase()}</Tag>
      ),
    },
    {
      title: 'Created At',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 180,
      render: (time: string) => dayjs(time).format('YYYY-MM-DD HH:mm:ss'),
    },
    {
      title: 'Actions',
      key: 'actions',
      width: 200,
      render: (_, record) => (
        <Space>
          {record.status === 'pending' && (
            <Button
              type="link"
              icon={<PlayCircleOutlined />}
              onClick={() => handleRun(record.id)}
            >
              Run
            </Button>
          )}
          {record.status === 'completed' && (
            <Button
              type="link"
              icon={<EyeOutlined />}
              onClick={() => handleViewResult(record.id)}
            >
              Result
            </Button>
          )}
          <Popconfirm
            title="Are you sure to delete this task?"
            onConfirm={() => handleDelete(record.id)}
            okText="Yes"
            cancelText="No"
          >
            <Button type="link" danger icon={<DeleteOutlined />}>
              Delete
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div>
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between' }}>
        <Space>
          <Select
            style={{ width: 200 }}
            placeholder="Filter by status"
            allowClear
            onChange={(value) => setStatusFilter(value)}
          >
            <Select.Option value="pending">Pending</Select.Option>
            <Select.Option value="running">Running</Select.Option>
            <Select.Option value="completed">Completed</Select.Option>
            <Select.Option value="failed">Failed</Select.Option>
          </Select>
        </Space>
        <Button
          type="primary"
          icon={<PlusOutlined />}
          onClick={() => setModalVisible(true)}
        >
          Create Task
        </Button>
      </div>

      <Table
        columns={columns}
        dataSource={tasks}
        rowKey="id"
        loading={loading}
        pagination={{
          current: page,
          pageSize,
          total,
          onChange: (newPage, newPageSize) => {
            fetchTasks({ page: newPage, page_size: newPageSize, status: statusFilter as any });
          },
        }}
      />

      <CreateTaskModal
        visible={modalVisible}
        onCancel={() => setModalVisible(false)}
        onSuccess={() => {
          setModalVisible(false);
          fetchTasks({ page: 1, page_size: 10, status: statusFilter as any });
        }}
      />
    </div>
  );
};
