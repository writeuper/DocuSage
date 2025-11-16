import React, { ReactNode } from 'react'
import { Navigate, useLocation } from 'react-router-dom'
import { useSelector } from 'react-redux'
import { RootState } from '@/store'

interface PrivateRouteProps {
  children: ReactNode
  role?: string
}

const PrivateRoute: React.FC<PrivateRouteProps> = ({ children, role }) => {
  const { isAuthenticated, userInfo } = useSelector((state: RootState) => state.user)
  const location = useLocation()

  // 如果未登录，重定向到登录页
  if (!isAuthenticated) {
    return <Navigate to="/login" state={{ from: location }} replace />
  }

  // 如果需要特定角色，检查用户角色
  if (role && userInfo && userInfo.role !== role) {
    // 根据用户当前所在路径决定重定向目标
    const redirectPath = userInfo.role === 'admin' ? '/admin' : '/'
    return <Navigate to={redirectPath} replace />
  }

  // 验证通过，渲染子组件
  return <>{children}</>
}

export default PrivateRoute