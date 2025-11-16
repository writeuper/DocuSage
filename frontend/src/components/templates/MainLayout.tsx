import React, { ReactNode } from 'react'
import { Layout, Menu } from 'antd'
import { Link, useLocation } from 'react-router-dom'
import { useSelector } from 'react-redux'
import { RootState } from '@/store'

const { Header, Sider, Content } = Layout

interface MenuItem {
  key: string
  label: string
  icon: ReactNode
}

interface MainLayoutProps {
  children: ReactNode
  menuItems: MenuItem[]
}

const MainLayout: React.FC<MainLayoutProps> = ({ children, menuItems }) => {
  const location = useLocation()
  const user = useSelector((state: RootState) => state.user.userInfo)
  
  // 从当前路径获取选中的菜单项键
  const getSelectedKey = () => {
    const pathParts = location.pathname.split('/')
    return pathParts[pathParts.length - 1] || 'home'
  }

  return (
    <Layout className="main-layout" style={{ minHeight: '100vh' }}>
      <Header className="header" style={{ background: '#fff', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <div className="logo" style={{ fontSize: '20px', fontWeight: 'bold', color: '#1890ff' }}>
          DocuSage 智能文档助手
        </div>
        <div className="user-info">
          {user && `${user.username} (${user.role})`}
        </div>
      </Header>
      <Layout>
        <Sider width={200} className="site-layout-background">
          <Menu
            mode="inline"
            selectedKeys={[getSelectedKey()]}
            style={{ height: '100%', borderRight: 0 }}
          >
            {menuItems.map(item => (
              <Menu.Item key={item.key} icon={item.icon}>
                <Link to={item.key === 'home' ? '/' : `${location.pathname.includes('admin') ? '/admin/' : '/'}${item.key}`}>
                  {item.label}
                </Link>
              </Menu.Item>
            ))}
          </Menu>
        </Sider>
        <Layout className="site-layout">
          <Content
            className="site-layout-background"
            style={{
              margin: '24px 16px',
              padding: 24,
              minHeight: 280,
              background: '#fff',
            }}
          >
            {children}
          </Content>
        </Layout>
      </Layout>
    </Layout>
  )
}

export default MainLayout