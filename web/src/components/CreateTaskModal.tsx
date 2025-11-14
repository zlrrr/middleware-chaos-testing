import { useState, useEffect } from 'react';
import { Modal, Form, Select, Input, InputNumber, message } from 'antd';
import type { TaskRequest, MiddlewareInfo } from '../types';
import { api } from '../services/api';

interface CreateTaskModalProps {
  visible: boolean;
  onCancel: () => void;
  onSuccess: () => void;
}

export const CreateTaskModal: React.FC<CreateTaskModalProps> = ({
  visible,
  onCancel,
  onSuccess,
}) => {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [middlewares, setMiddlewares] = useState<MiddlewareInfo[]>([]);
  const [selectedMiddleware, setSelectedMiddleware] = useState<string>('');

  useEffect(() => {
    const fetchMiddlewares = async () => {
      try {
        const data = await api.getMiddlewares();
        setMiddlewares(data);
      } catch (error: any) {
        message.error(error.message || 'Failed to fetch middlewares');
      }
    };
    fetchMiddlewares();
  }, []);

  const selectedMiddlewareInfo = middlewares.find((m) => m.name === selectedMiddleware);

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      setLoading(true);

      const request: TaskRequest = {
        middleware: values.middleware,
        duration: values.duration,
        operations: values.operations,
        concurrency: values.concurrency,
        config: {},
      };

      // Extract config fields
      if (selectedMiddlewareInfo) {
        Object.keys(selectedMiddlewareInfo.config_spec).forEach((key) => {
          if (values[key] !== undefined) {
            request.config[key] = String(values[key]);
          }
        });
      }

      await api.createTask(request);
      message.success('Task created successfully');
      form.resetFields();
      onSuccess();
    } catch (error: any) {
      message.error(error.message || 'Failed to create task');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Modal
      title="Create New Test Task"
      open={visible}
      onCancel={onCancel}
      onOk={handleSubmit}
      confirmLoading={loading}
      width={600}
    >
      <Form form={form} layout="vertical">
        <Form.Item
          label="Middleware"
          name="middleware"
          rules={[{ required: true, message: 'Please select middleware' }]}
        >
          <Select
            placeholder="Select middleware"
            onChange={(value) => setSelectedMiddleware(value)}
          >
            {middlewares.map((m) => (
              <Select.Option key={m.name} value={m.name}>
                {m.display_name}
              </Select.Option>
            ))}
          </Select>
        </Form.Item>

        {selectedMiddlewareInfo && (
          <>
            {Object.entries(selectedMiddlewareInfo.config_spec).map(([key, spec]) => (
              <Form.Item
                key={key}
                label={key.charAt(0).toUpperCase() + key.slice(1)}
                name={key}
                rules={[{ required: spec.required, message: `Please input ${key}` }]}
                tooltip={spec.description}
              >
                {spec.type === 'number' ? (
                  <InputNumber style={{ width: '100%' }} placeholder={spec.example} />
                ) : (
                  <Input placeholder={spec.example} />
                )}
              </Form.Item>
            ))}
          </>
        )}

        <Form.Item
          label="Duration"
          name="duration"
          rules={[{ required: true, message: 'Please input duration' }]}
          tooltip="e.g., 30s, 1m, 5m"
        >
          <Input placeholder="30s" />
        </Form.Item>

        <Form.Item label="Operations" name="operations" initialValue={5000}>
          <InputNumber min={1} max={1000000} style={{ width: '100%' }} />
        </Form.Item>

        <Form.Item label="Concurrency" name="concurrency" initialValue={10}>
          <InputNumber min={1} max={1000} style={{ width: '100%' }} />
        </Form.Item>
      </Form>
    </Modal>
  );
};
