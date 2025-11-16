/*
 * @Author: lixinda
 * @Description: 
 * @File: 
 * @Date: 2025-11-13 20:05:18
 */
import axios from 'axios'

// 创建axios实例
const api = axios.create({
  baseURL: '/api/auth',
  timeout: 10000,
})

// 请求拦截器 - 添加token
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers['Authorization'] = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
api.interceptors.response.use(
  (response) => {
    return response
  },
  (error) => {
    // 处理401错误 - 未授权
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

// 登录接口
export const login = (credentials: { username: string; password: string }) => {
  return api.post('/login', credentials)
}

// 登出接口
export const logout = () => {
  return api.post('/logout')
}

// 获取用户信息接口
export const getUserInfo = () => {
  return api.get('/userinfo')
}

// 刷新token接口
export const refreshToken = () => {
  return api.post('/refresh')
}

// 验证权限接口
export const checkPermission = (resource: string, action: string) => {
  return api.get(`/permission/check?resource=${resource}&action=${action}`)
}

export default api