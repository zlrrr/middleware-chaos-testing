import { useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { Card, Row, Col, Statistic, List, Tag, Spin, Alert } from 'antd';
import { ClockCircleOutlined, ThunderboltOutlined, CheckCircleOutlined } from '@ant-design/icons';
import { useResultStore } from '../stores/resultStore';
import { ScoreCard } from '../components/ScoreCard';

export const TaskResult: React.FC = () => {
  const { taskId } = useParams<{ taskId: string }>();
  const { fetchResult, getResult, loading, error } = useResultStore();

  useEffect(() => {
    if (taskId) {
      fetchResult(taskId);
    }
  }, [taskId, fetchResult]);

  const result = taskId ? getResult(taskId) : null;

  if (loading) {
    return (
      <div style={{ textAlign: 'center', padding: '100px 0' }}>
        <Spin size="large" />
      </div>
    );
  }

  if (error) {
    return <Alert message="Error" description={error} type="error" showIcon />;
  }

  if (!result) {
    return <Alert message="No result found" type="info" showIcon />;
  }

  const formatDuration = (ns: number) => {
    const ms = ns / 1000000;
    if (ms < 1000) return `${ms.toFixed(2)}ms`;
    return `${(ms / 1000).toFixed(2)}s`;
  };

  return (
    <div>
      <h2>Test Result - {result.middleware.toUpperCase()}</h2>

      <Row gutter={[16, 16]} style={{ marginTop: 24 }}>
        <Col xs={24} lg={8}>
          <ScoreCard result={result} />
        </Col>

        <Col xs={24} lg={16}>
          <Card title="Key Metrics" bordered={false}>
            <Row gutter={[16, 16]}>
              <Col span={8}>
                <Statistic
                  title="Availability"
                  value={(result.metrics.availability * 100).toFixed(2)}
                  suffix="%"
                  prefix={<CheckCircleOutlined />}
                  valueStyle={{ color: '#3f8600' }}
                />
              </Col>
              <Col span={8}>
                <Statistic
                  title="P95 Latency"
                  value={formatDuration(result.metrics.p95_latency)}
                  prefix={<ClockCircleOutlined />}
                />
              </Col>
              <Col span={8}>
                <Statistic
                  title="Throughput"
                  value={result.metrics.throughput.toFixed(2)}
                  suffix="ops/s"
                  prefix={<ThunderboltOutlined />}
                />
              </Col>
              <Col span={8}>
                <Statistic
                  title="Error Rate"
                  value={(result.metrics.error_rate * 100).toFixed(3)}
                  suffix="%"
                  valueStyle={{ color: result.metrics.error_rate > 0.01 ? '#cf1322' : '#3f8600' }}
                />
              </Col>
              <Col span={8}>
                <Statistic
                  title="P99 Latency"
                  value={formatDuration(result.metrics.p99_latency)}
                />
              </Col>
              <Col span={8}>
                <Statistic
                  title="MTTR"
                  value={formatDuration(result.metrics.mttr)}
                />
              </Col>
            </Row>
          </Card>
        </Col>
      </Row>

      {result.issues && result.issues.length > 0 && (
        <Card title="Issues" style={{ marginTop: 16 }} bordered={false}>
          <List
            dataSource={result.issues}
            renderItem={(issue) => (
              <List.Item>
                <List.Item.Meta
                  title={
                    <div>
                      <Tag color={issue.severity === 'CRITICAL' ? 'red' : issue.severity === 'HIGH' ? 'orange' : 'blue'}>
                        {issue.severity}
                      </Tag>
                      <span>{issue.type.replace(/_/g, ' ').toUpperCase()}</span>
                    </div>
                  }
                  description={
                    <div>
                      <div>{issue.message}</div>
                      <div style={{ marginTop: 8 }}>
                        Current: <strong>{issue.current}</strong> | Expected: <strong>{issue.expected}</strong>
                      </div>
                    </div>
                  }
                />
              </List.Item>
            )}
          />
        </Card>
      )}

      {result.recommendations && result.recommendations.length > 0 && (
        <Card title="Recommendations" style={{ marginTop: 16 }} bordered={false}>
          <List
            dataSource={result.recommendations}
            renderItem={(rec) => (
              <List.Item>
                <List.Item.Meta
                  title={
                    <div>
                      <Tag color={rec.priority === 'HIGH' ? 'red' : rec.priority === 'MEDIUM' ? 'orange' : 'blue'}>
                        {rec.priority}
                      </Tag>
                      <Tag>{rec.category}</Tag>
                      <span style={{ marginLeft: 8 }}>{rec.title}</span>
                    </div>
                  }
                  description={
                    <div>
                      <div>{rec.message}</div>
                      {rec.actions && rec.actions.length > 0 && (
                        <div style={{ marginTop: 8 }}>
                          <strong>Actions:</strong>
                          <ul>
                            {rec.actions.map((action, idx) => (
                              <li key={idx}>{action}</li>
                            ))}
                          </ul>
                        </div>
                      )}
                    </div>
                  }
                />
              </List.Item>
            )}
          />
        </Card>
      )}

      {result.rationale && (
        <Card title="Rationale" style={{ marginTop: 16 }} bordered={false}>
          <p>{result.rationale}</p>
        </Card>
      )}
    </div>
  );
};
