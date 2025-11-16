/*
 * @Author: lixinda
 * @Description: 
 * @File: 
 * @Date: 2025-11-13 20:00:44
 */
import React from 'react'
import ReactDOM from 'react-dom/client'
import { ConfigProvider } from 'antd'
import zhCN from 'antd/locale/zh_CN'
import { Provider } from 'react-redux'
import { QueryClient, QueryClientProvider } from 'react-query'
import App from './App'
import store from './store'
import './assets/styles/global.css'

// 创建React Query客户端
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      refetchOnWindowFocus: false,
    },
  },
})

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <Provider store={store}>
      <QueryClientProvider client={queryClient}>
        <ConfigProvider locale={zhCN}>
          <App />
        </ConfigProvider>
      </QueryClientProvider>
    </Provider>
  </React.StrictMode>,
)