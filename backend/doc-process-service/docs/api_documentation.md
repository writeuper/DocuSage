# API文档

本文档详细描述了DocuSage文档处理服务提供的所有API接口，包括接口路径、请求参数、响应格式和示例。

## 基本信息

- **基础URL**: `/api/v1`
- **认证方式**: JWT Bearer Token (在Authorization头中)
- **内容类型**: `application/json` (除文件上传接口)

## 健康检查接口

### 获取服务健康状态

```
GET /health
```

**描述**: 检查整体服务健康状态

**响应示例**:

```json
{
  "status": "healthy",
  "timestamp": "2024-01-01T12:00:00Z",
  "services": {
    "database": "healthy",
    "redis": "healthy",
    "milvus": "healthy",
    "minio": "healthy"
  }
}
```

### 获取数据库健康状态

```
GET /health/db
```

**描述**: 检查数据库连接健康状态

**响应示例**:

```json
{
  "status": "healthy",
  "connection_pool": {
    "size": 20,
    "checkedin": 15,
    "checkedout": 5,
    "overflow": 0
  }
}
```

## 文档管理接口

### 上传文档

```
POST /documents/upload
```

**描述**: 上传新文档到系统

**请求参数**: 表单数据
- `file`: 文件对象 (必填)
- `metadata`: JSON字符串，包含文档元数据 (可选)
- `chunk_size`: 分块大小 (可选，默认1000)
- `chunk_overlap`: 块重叠大小 (可选，默认200)

**响应示例**:

```json
{
  "document_id": "doc_123456",
  "filename": "example.pdf",
  "size": 1024000,
  "content_type": "application/pdf",
  "status": "uploaded",
  "task_id": "task_789012",
  "created_at": "2024-01-01T12:00:00Z"
}
```

### 获取文档列表

```
GET /documents
```

**描述**: 获取用户的文档列表

**查询参数**:
- `page`: 页码 (默认1)
- `per_page`: 每页数量 (默认20)
- `status`: 过滤状态 (可选)
- `sort_by`: 排序字段 (默认created_at)
- `sort_order`: 排序方向 (asc/desc，默认desc)

**响应示例**:

```json
{
  "items": [
    {
      "document_id": "doc_123456",
      "filename": "example.pdf",
      "size": 1024000,
      "status": "processed",
      "chunks_count": 42,
      "created_at": "2024-01-01T12:00:00Z",
      "updated_at": "2024-01-01T12:10:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "per_page": 20,
  "total_pages": 1
}
```

### 获取文档详情

```
GET /documents/{document_id}
```

**描述**: 获取指定文档的详细信息

**路径参数**:
- `document_id`: 文档ID

**响应示例**:

```json
{
  "document_id": "doc_123456",
  "filename": "example.pdf",
  "size": 1024000,
  "content_type": "application/pdf",
  "status": "processed",
  "chunks_count": 42,
  "metadata": {
    "author": "John Doe",
    "category": "technical"
  },
  "created_at": "2024-01-01T12:00:00Z",
  "updated_at": "2024-01-01T12:10:00Z",
  "processed_at": "2024-01-01T12:10:00Z"
}
```

### 更新文档信息

```
PUT /documents/{document_id}
```

**描述**: 更新文档的元数据信息

**路径参数**:
- `document_id`: 文档ID

**请求体**:

```json
{
  "metadata": {
    "author": "Jane Doe",
    "category": "marketing"
  }
}
```

**响应示例**:

```json
{
  "document_id": "doc_123456",
  "filename": "example.pdf",
  "metadata": {
    "author": "Jane Doe",
    "category": "marketing"
  },
  "updated_at": "2024-01-01T13:00:00Z"
}
```

### 删除文档

```
DELETE /documents/{document_id}
```

**描述**: 删除指定文档

**路径参数**:
- `document_id`: 文档ID

**响应示例**:

```json
{
  "success": true,
  "message": "文档已成功删除",
  "document_id": "doc_123456"
}
```

### 获取文档处理状态

```
GET /documents/{document_id}/status
```

**描述**: 获取文档的处理状态和进度

**路径参数**:
- `document_id`: 文档ID

**响应示例**:

```json
{
  "document_id": "doc_123456",
  "status": "processing",
  "progress": 65,
  "stage": "embedding",
  "steps": [
    {"name": "upload", "status": "completed", "timestamp": "2024-01-01T12:00:00Z"},
    {"name": "parsing", "status": "completed", "timestamp": "2024-01-01T12:02:00Z"},
    {"name": "chunking", "status": "completed", "timestamp": "2024-01-01T12:05:00Z"},
    {"name": "embedding", "status": "in_progress", "timestamp": "2024-01-01T12:06:00Z"},
    {"name": "indexing", "status": "pending", "timestamp": null}
  ],
  "updated_at": "2024-01-01T12:08:00Z"
}
```

## 文档块管理接口

### 获取文档块

```
GET /documents/{document_id}/chunks
```

**描述**: 获取文档的所有块

**路径参数**:
- `document_id`: 文档ID

**查询参数**:
- `page`: 页码 (默认1)
- `per_page`: 每页数量 (默认50)

**响应示例**:

```json
{
  "items": [
    {
      "chunk_id": "chunk_123",
      "document_id": "doc_123456",
      "content": "这是文档的第一部分内容...",
      "position": 0,
      "created_at": "2024-01-01T12:05:00Z"
    }
  ],
  "total": 42,
  "page": 1,
  "per_page": 50,
  "total_pages": 1
}
```

### 获取单个块

```
GET /chunks/{chunk_id}
```

**描述**: 获取指定的文档块

**路径参数**:
- `chunk_id`: 块ID

**响应示例**:

```json
{
  "chunk_id": "chunk_123",
  "document_id": "doc_123456",
  "content": "这是文档的第一部分内容...",
  "metadata": {
    "page_number": 1,
    "section": "introduction"
  },
  "position": 0,
  "created_at": "2024-01-01T12:05:00Z"
}
```

## 搜索接口

### 向量搜索

```
POST /search
```

**描述**: 根据查询文本执行向量相似度搜索

**请求体**:

```json
{
  "query": "如何优化数据库性能？",
  "top_k": 5,
  "document_ids": ["doc_123456", "doc_789012"],  // 可选，限制搜索范围
  "filters": {
    "category": "technical",
    "min_similarity": 0.7
  }
}
```

**响应示例**:

```json
{
  "query": "如何优化数据库性能？",
  "results": [
    {
      "chunk_id": "chunk_123",
      "document_id": "doc_123456",
      "filename": "数据库优化指南.pdf",
      "content": "数据库性能优化的关键在于正确的索引设计...",
      "similarity": 0.85,
      "metadata": {
        "page_number": 10,
        "section": "索引优化"
      }
    }
  ],
  "total": 5
}
```

### 混合搜索

```
POST /search/hybrid
```

**描述**: 结合关键词搜索和向量搜索的混合搜索

**请求体**:

```json
{
  "query": "数据库索引优化",
  "top_k": 10,
  "keyword_weight": 0.3,  // 关键词搜索权重
  "vector_weight": 0.7,    // 向量搜索权重
  "filters": {
    "created_at": {
      "$gte": "2023-01-01"
    }
  }
}
```

**响应示例**:

```json
{
  "query": "数据库索引优化",
  "results": [
    {
      "chunk_id": "chunk_123",
      "document_id": "doc_123456",
      "filename": "数据库优化指南.pdf",
      "content": "数据库索引优化是提升查询性能的重要手段...",
      "hybrid_score": 0.92,
      "keyword_score": 0.95,
      "vector_score": 0.90,
      "metadata": {
        "page_number": 10
      }
    }
  ]
}
```

## 任务管理接口

### 获取任务列表

```
GET /tasks
```

**描述**: 获取任务列表

**查询参数**:
- `status`: 过滤任务状态 (可选)
- `page`: 页码 (默认1)
- `per_page`: 每页数量 (默认20)

**响应示例**:

```json
{
  "items": [
    {
      "task_id": "task_789012",
      "type": "document_processing",
      "status": "completed",
      "progress": 100,
      "document_id": "doc_123456",
      "created_at": "2024-01-01T12:00:00Z",
      "updated_at": "2024-01-01T12:10:00Z",
      "completed_at": "2024-01-01T12:10:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "per_page": 20
}
```

### 获取任务详情

```
GET /tasks/{task_id}
```

**描述**: 获取指定任务的详细信息

**路径参数**:
- `task_id`: 任务ID

**响应示例**:

```json
{
  "task_id": "task_789012",
  "type": "document_processing",
  "status": "completed",
  "progress": 100,
  "document_id": "doc_123456",
  "filename": "example.pdf",
  "details": {
    "parsed_pages": 20,
    "created_chunks": 42,
    "embedding_time_ms": 1500
  },
  "logs": [
    {"level": "info", "message": "开始处理文档", "timestamp": "2024-01-01T12:00:00Z"},
    {"level": "info", "message": "文档解析完成", "timestamp": "2024-01-01T12:02:00Z"},
    {"level": "info", "message": "文档处理完成", "timestamp": "2024-01-01T12:10:00Z"}
  ],
  "created_at": "2024-01-01T12:00:00Z",
  "updated_at": "2024-01-01T12:10:00Z",
  "completed_at": "2024-01-01T12:10:00Z"
}
```

### 取消任务

```
DELETE /tasks/{task_id}
```

**描述**: 取消正在执行的任务

**路径参数**:
- `task_id`: 任务ID

**响应示例**:

```json
{
  "success": true,
  "message": "任务已取消",
  "task_id": "task_789012"
}
```

## 统计接口

### 获取处理统计

```
GET /stats/processing
```

**描述**: 获取文档处理统计信息

**查询参数**:
- `time_range`: 时间范围 (day/week/month，默认day)

**响应示例**:

```json
{
  "total_documents": 100,
  "total_chunks": 5000,
  "processed_today": 10,
  "failed_today": 1,
  "average_processing_time": 45,
  "time_distribution": {
    "upload": 5,
    "parsing": 15,
    "chunking": 10,
    "embedding": 25,
    "indexing": 5
  }
}
```

### 获取搜索统计

```
GET /stats/search
```

**描述**: 获取搜索统计信息

**查询参数**:
- `time_range`: 时间范围 (day/week/month，默认day)

**响应示例**:

```json
{
  "total_searches": 1000,
  "searches_today": 150,
  "average_results": 4.5,
  "average_search_time_ms": 120,
  "top_queries": [
    {"query": "数据库优化", "count": 50},
    {"query": "API设计", "count": 45}
  ]
}
```

## 错误码说明

| 错误码 | 状态码 | 描述 |
|-------|-------|------|
| `INVALID_REQUEST` | 400 | 请求参数无效 |
| `UNAUTHORIZED` | 401 | 未授权访问 |
| `FORBIDDEN` | 403 | 禁止访问 |
| `NOT_FOUND` | 404 | 资源不存在 |
| `CONFLICT` | 409 | 资源冲突 |
| `VALIDATION_ERROR` | 422 | 请求参数验证失败 |
| `SERVER_ERROR` | 500 | 服务器内部错误 |
| `SERVICE_UNAVAILABLE` | 503 | 服务不可用 |
| `FILE_TOO_LARGE` | 413 | 文件过大 |
| `UNSUPPORTED_FILE_TYPE` | 415 | 不支持的文件类型 |

## 错误响应格式

```json
{
  "error": {
    "code": "INVALID_REQUEST",
    "message": "请求参数无效",
    "details": {
      "field": "file",
      "reason": "文件不能为空"
    }
  },
  "timestamp": "2024-01-01T12:00:00Z"
}
```

## 使用示例

### 上传文档示例

```bash
curl -X POST "http://localhost:8000/api/v1/documents/upload" \
  -H "Authorization: Bearer your-jwt-token" \
  -F "file=@example.pdf" \
  -F "metadata={\"category\":\"technical\"}"
```

### 搜索示例

```bash
curl -X POST "http://localhost:8000/api/v1/search" \
  -H "Authorization: Bearer your-jwt-token" \
  -H "Content-Type: application/json" \
  -d '{"query": "数据库优化", "top_k": 5}'
```

### 获取文档状态示例

```bash
curl -X GET "http://localhost:8000/api/v1/documents/doc_123456/status" \
  -H "Authorization: Bearer your-jwt-token"
```

## 注意事项

1. 所有API接口都需要在请求头中提供有效的JWT令牌
2. 文件上传接口使用`multipart/form-data`格式
3. 对于大量文档的批量操作，建议使用异步接口
4. 搜索结果默认返回前5个最相关的结果，可通过top_k参数调整
5. 系统默认限制单个文件大小为100MB