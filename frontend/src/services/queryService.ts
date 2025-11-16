import axios from 'axios'
import { SearchResult } from '@/types'

const api = axios.create({
  baseURL: '/api/query',
  timeout: 30000, // 查询可能需要更长时间
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

// 搜索文档接口
export const searchDocuments = (params: {
  query: string
  docType?: string
  department?: string
  dateRange?: any[]
  topK?: number
}) => {
  // 转换日期范围为字符串格式
  const formattedParams = { ...params }
  if (params.dateRange && params.dateRange.length === 2) {
    formattedParams.dateRange = [
      params.dateRange[0]?.format('YYYY-MM-DD'),
      params.dateRange[1]?.format('YYYY-MM-DD')
    ]
  }
  
  return api.get<SearchResult>('/search', { params: formattedParams })
}

// 获取查询历史接口
export const getSearchHistory = (limit: number = 10) => {
  return api.get('/history', { params: { limit } })
}

// 清除查询历史接口
export const clearSearchHistory = () => {
  return api.delete('/history')
}

// 获取推荐问题接口
export const getRecommendedQuestions = (keywords?: string) => {
  return api.get('/recommendations', { params: { keywords } })
}

// 对查询结果进行反馈接口
export const feedbackSearchResult = (data: {
  queryId: string
  resultId: string
  rating: number // 1-5分
  comment?: string
}) => {
  return api.post('/feedback', data)
}

// 多轮对话接口
export const continueConversation = (data: {
  conversationId: string
  query: string
}) => {
  return api.post('/conversation/continue', data)
}

export default api