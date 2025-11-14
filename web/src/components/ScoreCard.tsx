import { Card, Progress, Tag } from 'antd';
import type { TestResult } from '../types';

interface ScoreCardProps {
  result: TestResult;
}

const getGradeColor = (grade: string) => {
  switch (grade) {
    case 'EXCELLENT':
      return '#52c41a';
    case 'GOOD':
      return '#1890ff';
    case 'FAIR':
      return '#faad14';
    case 'POOR':
      return '#fa8c16';
    case 'FAILED':
      return '#f5222d';
    default:
      return '#d9d9d9';
  }
};

const getStatusColor = (status: string) => {
  switch (status) {
    case 'PASS':
      return 'success';
    case 'WARNING':
      return 'warning';
    case 'FAIL':
      return 'error';
    default:
      return 'default';
  }
};

export const ScoreCard: React.FC<ScoreCardProps> = ({ result }) => {
  const color = getGradeColor(result.grade);

  return (
    <Card
      title="Overall Score"
      bordered={false}
      style={{ textAlign: 'center' }}
    >
      <div style={{ fontSize: 48, fontWeight: 'bold', color, marginBottom: 16 }}>
        {result.score.toFixed(1)}
        <span style={{ fontSize: 24, color: '#999' }}> / 100</span>
      </div>

      <div style={{ marginBottom: 24 }}>
        <Tag color={color} style={{ fontSize: 16, padding: '4px 16px' }}>
          {result.grade}
        </Tag>
        <Tag color={getStatusColor(result.status)} style={{ fontSize: 16, padding: '4px 16px' }}>
          {result.status}
        </Tag>
      </div>

      <div style={{ marginTop: 24 }}>
        <div style={{ marginBottom: 8 }}>
          <span>Availability</span>
          <Progress
            percent={Math.round((result.scores.availability / 30) * 100)}
            strokeColor="#52c41a"
            format={() => `${result.scores.availability.toFixed(1)}/30`}
          />
        </div>
        <div style={{ marginBottom: 8 }}>
          <span>Performance</span>
          <Progress
            percent={Math.round((result.scores.performance / 25) * 100)}
            strokeColor="#1890ff"
            format={() => `${result.scores.performance.toFixed(1)}/25`}
          />
        </div>
        <div style={{ marginBottom: 8 }}>
          <span>Reliability</span>
          <Progress
            percent={Math.round((result.scores.reliability / 25) * 100)}
            strokeColor="#722ed1"
            format={() => `${result.scores.reliability.toFixed(1)}/25`}
          />
        </div>
        <div>
          <span>Resilience</span>
          <Progress
            percent={Math.round((result.scores.resilience / 20) * 100)}
            strokeColor="#fa8c16"
            format={() => `${result.scores.resilience.toFixed(1)}/20`}
          />
        </div>
      </div>
    </Card>
  );
};
