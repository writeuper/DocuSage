// 用户相关类型
export interface User {
  id: string
  username: string
  role: string
  department: string
  permissions: string[]
  createdAt: string
  updatedAt: string
}

// 搜索结果类型
export interface SearchResultItem {
  id: string
  title: string
  content: string
  highlightedText: string
  similarity: number
  docType: string
  department: string
  uploadDate: string
  url: string
  version: string
}

// 完整搜索结果响应
export interface SearchResult {
  answer: string
  results: SearchResultItem[]
  total: number
  processingTime: number
  queryId: string
  conversationId?: string
}

// 文档相关类型
export interface Document {
  id: string
  title: string
  filename: string
  docType: string
  department: string
  size: number
  version: string
  status: string // 'processing', 'completed', 'failed'
  uploadDate: string
  processedDate?: string
  qualityScore?: number
  chunksCount?: number
  uploadedBy: string
}

// 查询历史类型
export interface SearchHistory {
  id: string
  query: string
  timestamp: string
  resultsCount: number
  processingTime: number
}

// 权限相关类型
export interface Permission {
  resource: string
  action: string
  description: string
}

// 部门类型
export interface Department {
  id: string
  name: string
  description: string
  parentId?: string
}

// 统计数据类型
export interface Statistics {
  totalDocuments: number
  totalUsers: number
  totalQueries: number
  dailyQueries: number
  monthlyNewDocs: number
  activeUsers: number
}