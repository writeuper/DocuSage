import React from 'react'
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom'
import { Layout } from 'antd'
import { UserOutlined, SettingOutlined, FileTextOutlined } from '@ant-design/icons'
import MainLayout from './components/templates/MainLayout'
import Login from './pages/common/Login'
import Home from './pages/user/Home'
import DocumentSearch from './pages/user/DocumentSearch'
import SceneTools from './pages/user/SceneTools'
import UserCenter from './pages/user/UserCenter'
import KnowledgeBaseManagement from './pages/admin/KnowledgeBaseManagement'
import SystemConfig from './pages/admin/SystemConfig'
import MonitorPanel from './pages/admin/MonitorPanel'
import OperationTools from './pages/admin/OperationTools'
import PrivateRoute from './utils/PrivateRoute'

const { Content } = Layout

const App: React.FC = () => {
  return (
    <Router>
      <Routes>
        {/* 公共路由 */}
        <Route path="/login" element={<Login />} />
        
        {/* 用户端路由 */}
        <Route path="/" element={
          <PrivateRoute>
            <MainLayout 
              menuItems={[
                { key: 'home', label: '首页', icon: <FileTextOutlined /> },
                { key: 'search', label: '文档检索', icon: <UserOutlined /> },
                { key: 'tools', label: '场景工具', icon: <SettingOutlined /> },
                { key: 'center', label: '个人中心', icon: <UserOutlined /> },
              ]}
            >
              <Content>
                <Route index element={<Home />} />
                <Route path="search" element={<DocumentSearch />} />
                <Route path="tools" element={<SceneTools />} />
                <Route path="center" element={<UserCenter />} />
              </Content>
            </MainLayout>
          </PrivateRoute>
        } />
        
        {/* 管理员端路由 */}
        <Route path="/admin" element={
          <PrivateRoute role="admin">
            <MainLayout 
              menuItems={[
                { key: 'knowledge', label: '知识库管理', icon: <FileTextOutlined /> },
                { key: 'config', label: '系统配置', icon: <SettingOutlined /> },
                { key: 'monitor', label: '监控面板', icon: <UserOutlined /> },
                { key: 'operation', label: '运维工具', icon: <SettingOutlined /> },
              ]}
            >
              <Content>
                <Route index element={<KnowledgeBaseManagement />} />
                <Route path="config" element={<SystemConfig />} />
                <Route path="monitor" element={<MonitorPanel />} />
                <Route path="operation" element={<OperationTools />} />
              </Content>
            </MainLayout>
          </PrivateRoute>
        } />
      </Routes>
    </Router>
  )
}

export default App