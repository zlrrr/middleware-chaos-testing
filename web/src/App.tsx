import { BrowserRouter, Routes, Route, Navigate, Link } from 'react-router-dom';
import { Layout, Menu } from 'antd';
import { DashboardOutlined, UnorderedListOutlined } from '@ant-design/icons';
import { Dashboard } from './pages/Dashboard';
import { Tasks } from './pages/Tasks';
import { TaskResult } from './pages/TaskResult';

const { Header, Content } = Layout;

function App() {
  return (
    <BrowserRouter>
      <Layout style={{ minHeight: '100vh' }}>
        <Header style={{ display: 'flex', alignItems: 'center' }}>
          <div style={{ color: 'white', fontSize: 20, fontWeight: 'bold', marginRight: 50 }}>
            MCT Platform
          </div>
          <Menu
            theme="dark"
            mode="horizontal"
            defaultSelectedKeys={['dashboard']}
            style={{ flex: 1, minWidth: 0 }}
          >
            <Menu.Item key="dashboard" icon={<DashboardOutlined />}>
              <Link to="/dashboard">Dashboard</Link>
            </Menu.Item>
            <Menu.Item key="tasks" icon={<UnorderedListOutlined />}>
              <Link to="/tasks">Tasks</Link>
            </Menu.Item>
          </Menu>
        </Header>
        <Content style={{ padding: '24px 50px' }}>
          <div style={{ background: '#fff', padding: 24, minHeight: 'calc(100vh - 112px)' }}>
            <Routes>
              <Route path="/" element={<Navigate to="/dashboard" replace />} />
              <Route path="/dashboard" element={<Dashboard />} />
              <Route path="/tasks" element={<Tasks />} />
              <Route path="/tasks/:taskId/result" element={<TaskResult />} />
            </Routes>
          </div>
        </Content>
      </Layout>
    </BrowserRouter>
  );
}

export default App;
