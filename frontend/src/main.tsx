/*
 * @Author: lixinda
 * @Description: 
 * @File: 
 * @Date: 2025-11-13 20:00:44
 */
import React, { useEffect } from 'react'
import ReactDOM from 'react-dom/client'
import { ConfigProvider } from 'antd'
import zhCN from 'antd/locale/zh_CN'
import { Provider } from 'react-redux'
import { QueryClient, QueryClientProvider } from 'react-query'
import App from './App'
import store from './store'
import './assets/styles/global.css'
import { getUserInfo } from '@/services/authService'
import { login } from './store/slices/userSlice'

// 创建React Query客户端
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      refetchOnWindowFocus: false,
    },
  },
})

// 应用初始化组件，用于验证token和获取用户信息
const AppWithInitialization: React.FC = () => {
  useEffect(() => {
    const initializeApp = async () => {
      const token = localStorage.getItem('token')
      if (token) {
        try {
          // 调用getUserInfo接口获取完整用户信息
          const userInfoResponse = await getUserInfo()
          
          // 构建完整的用户信息对象，包含token
          const userData = {
            ...userInfoResponse.data,
            token: token
          }
          
          // 更新Redux状态
          store.dispatch(login(userData))
        } catch (error) {
          // Token无效或请求失败，清除存储
          localStorage.removeItem('token')
        }
      }
    }

    initializeApp()
  }, [])


  return <App />;
};

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <Provider store={store}>
      <QueryClientProvider client={queryClient}>
        <ConfigProvider locale={zhCN}>
          <AppWithInitialization />
        </ConfigProvider>
      </QueryClientProvider>
    </Provider>
  </React.StrictMode>,
)