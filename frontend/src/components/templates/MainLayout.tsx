import React, { ReactNode, useState } from 'react'
import { Layout, Menu, Dropdown, Avatar, Modal, message } from 'antd'
import { UserOutlined, SettingOutlined, LogoutOutlined, FileTextOutlined } from '@ant-design/icons'
import { Link, useLocation, useNavigate } from 'react-router-dom'
import { useSelector, useDispatch } from 'react-redux'
import { RootState } from '@/store'
import { logout } from '@/store/slices/userSlice'

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
  const navigate = useNavigate()
  const dispatch = useDispatch()
  const user = useSelector((state: RootState) => state.user.userInfo)
  const [logoutModalVisible, setLogoutModalVisible] = useState(false)
  
  // 从当前路径获取选中的菜单项键
  const getSelectedKey = () => {
    const pathParts = location.pathname.split('/')
    return pathParts[pathParts.length - 1] || 'home'
  }

  // 退出登录
  const handleLogout = () => {
    setLogoutModalVisible(false)
    dispatch(logout())
    message.success('已安全退出')
    navigate('/login')
  }

  // 跳转到个人中心
  const goToUserCenter = () => {
    const userCenterPath = location.pathname.includes('admin') ? '/admin/center' : '/center'
    navigate(userCenterPath)
  }

  // 用户下拉菜单项
  const userMenuItems = [
    {
      key: 'profile',
      icon: <UserOutlined />,
      label: '个人中心',
      onClick: goToUserCenter,
    },
    {
      key: 'settings',
      icon: <SettingOutlined />,
      label: '账户设置',
      onClick: () => message.info('账户设置功能开发中'),
    },
    {
      type: 'divider' as const,
    },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: '退出登录',
      onClick: () => setLogoutModalVisible(true),
    },
  ]

  return (
    <Layout className="main-layout" style={{ minHeight: '100vh' }}>
      <Header className="header" style={{ background: '#fff', display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '0 24px' }}>
        <div className="logo" style={{ fontSize: '20px', fontWeight: 'bold', color: '#1890ff' }}>
          DocuSage 智能文档助手
        </div>
        <div className="user-info" style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
          <span style={{ color: '#666' }}>
            {user ? `${user.username || '未知用户'} (${user.role || '未知角色'})` : ''}
          </span>
          <Dropdown
            menu={{ items: userMenuItems }}
            placement="bottomRight"
            arrow
          >
            <Avatar 
              icon={<UserOutlined />} 
              style={{ 
                cursor: 'pointer',
                backgroundColor: user?.role === 'admin' ? '#f50' : '#1890ff' 
              }}
            />
          </Dropdown>
        </div>
      </Header>
      <Layout>
        <Sider width={200} className="site-layout-background">
          <Menu
            mode="inline"
            selectedKeys={[getSelectedKey()]}
            style={{ height: '100%', borderRight: 0 }}
            items={menuItems.map(item => ({
              key: item.key,
              icon: item.icon,
              label: (
                <Link to={item.key === 'home' ? '/' : `${location.pathname.includes('admin') ? '/admin/' : '/'}${item.key}`}>
                  {item.label}
                </Link>
              )
            }))}
          />
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
      
      {/* 退出确认对话框 */}
      <Modal
        title="确认退出"
        open={logoutModalVisible}
        onOk={handleLogout}
        onCancel={() => setLogoutModalVisible(false)}
        okText="确认退出"
        cancelText="取消"
        okButtonProps={{ danger: true }}
      >
        <p>您确定要退出登录吗？</p>
      </Modal>
    </Layout>
  )
}

export default MainLayout