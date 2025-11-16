# 配置指南

本文档提供了DocuSage文档处理服务的详细配置说明，包括所有可用的环境变量、配置选项及其默认值。

## 配置文件结构

配置通过环境变量进行管理。有两种方式设置配置：

1. 创建并编辑 `.env` 文件（推荐用于开发环境）
2. 直接设置系统环境变量（推荐用于生产环境）

## 核心配置项

### 1. 环境配置

| 环境变量 | 默认值 | 描述 |
|---------|-------|------|
| `ENVIRONMENT` | `development` | 运行环境，可选值：`development`, `testing`, `production` |

### 2. 服务器配置

| 环境变量 | 默认值 | 描述 |
|---------|-------|------|
| `SERVER_HOST` | `0.0.0.0` | 服务器绑定地址 |
| `SERVER_PORT` | `8000` | 服务器监听端口 |
| `SERVER_WORKERS` | `2` | Gunicorn工作进程数 |
| `SERVER_TIMEOUT` | `300` | 服务器超时时间（秒） |

### 3. 数据库配置

| 环境变量 | 默认值 | 描述 |
|---------|-------|------|
| `DATABASE_URL` | `postgresql://admin:password@localhost:5432/docu_sage` | 数据库连接URL |
| `DATABASE_POOL_SIZE` | `20` | 数据库连接池大小 |
| `DATABASE_MAX_OVERFLOW` | `10` | 最大溢出连接数 |
| `DATABASE_POOL_RECYCLE` | `3600` | 连接回收时间（秒） |
| `DATABASE_ECHO` | `false` | 是否打印SQL语句 |

### 4. Redis配置

| 环境变量 | 默认值 | 描述 |
|---------|-------|------|
| `REDIS_URL` | `redis://localhost:6379/0` | Redis连接URL |
| `REDIS_MAX_CONNECTIONS` | `50` | 最大连接数 |
| `REDIS_MIN_IDLE_CONNECTIONS` | `5` | 最小空闲连接数 |
| `REDIS_CONNECT_TIMEOUT` | `5` | 连接超时（秒） |
| `REDIS_SOCKET_TIMEOUT` | `5` | 套接字超时（秒） |
| `REDIS_HEALTH_CHECK_INTERVAL` | `30` | 健康检查间隔（秒） |
| `REDIS_TLS_ENABLED` | `false` | 是否启用TLS |
| `REDIS_TLS_VERIFY_MODE` | `CERT_NONE` | TLS证书验证模式 |

### 5. Milvus配置

| 环境变量 | 默认值 | 描述 |
|---------|-------|------|
| `MILVUS_URI` | `http://localhost:19530` | Milvus服务地址 |
| `MILVUS_TOKEN` | `` | 访问令牌 |
| `MILVUS_COLLECTION_NAME` | `document_chunks` | 集合名称 |
| `MILVUS_INDEX_TYPE` | `HNSW` | 索引类型 |
| `MILVUS_METRIC_TYPE` | `COSINE` | 距离度量类型 |
| `MILVUS_INDEX_PARAMS` | `{"M": 8, "efConstruction": 64}` | 索引参数 |
| `MILVUS_SEARCH_PARAMS` | `{"ef": 64}` | 搜索参数 |
| `MILVUS_TLS_ENABLED` | `false` | 是否启用TLS |
| `MILVUS_TIMEOUT` | `30` | 超时时间（秒） |
| `MILVUS_MAX_RETRIES` | `3` | 最大重试次数 |

### 6. MinIO配置

| 环境变量 | 默认值 | 描述 |
|---------|-------|------|
| `MINIO_ENDPOINT` | `localhost:9000` | MinIO服务地址 |
| `MINIO_ACCESS_KEY` | `minioadmin` | 访问密钥 |
| `MINIO_SECRET_KEY` | `minioadmin` | 密钥 |
| `MINIO_SECURE` | `false` | 是否使用HTTPS |
| `MINIO_BUCKET_NAME` | `docu-sage` | 存储桶名称 |
| `MINIO_REGION` | `us-east-1` | 存储区域 |
| `MINIO_PART_SIZE` | `10485760` | 分段上传大小（10MB） |

### 7. 嵌入服务配置

| 环境变量 | 默认值 | 描述 |
|---------|-------|------|
| `EMBEDDING_API_URL` | `http://embedding-service:8000/v1/embeddings` | 嵌入API地址 |
| `EMBEDDING_API_KEY` | `` | API密钥 |
| `EMBEDDING_DIMENSION` | `768` | 嵌入维度 |

### 8. 文件处理配置

| 环境变量 | 默认值 | 描述 |
|---------|-------|------|
| `MAX_FILE_SIZE` | `104857600` | 最大文件大小（100MB） |
| `ALLOWED_FILE_TYPES` | `pdf,doc,docx,txt,csv,xlsx,pptx,md` | 允许的文件类型 |
| `TEMP_FILE_DIR` | `/tmp` | 临时文件目录 |
| `CHUNK_SIZE` | `1000` | 文本分块大小 |
| `CHUNK_OVERLAP` | `200` | 块重叠大小 |

## 高级配置

### 性能优化配置

| 环境变量 | 默认值 | 描述 |
|---------|-------|------|
| `RATE_LIMIT_PER_MINUTE` | `60` | 每分钟请求限制 |
| `RATE_LIMIT_PER_HOUR` | `1000` | 每小时请求限制 |

### 日志配置

| 环境变量 | 默认值 | 描述 |
|---------|-------|------|
| `LOG_LEVEL` | `INFO` | 日志级别 |
| `LOG_FORMAT` | `json` | 日志格式 |
| `LOG_FILE` | `logs/app.log` | 日志文件路径 |

### CORS配置

| 环境变量 | 默认值 | 描述 |
|---------|-------|------|
| `CORS_ORIGINS` | `*` | 允许的来源 |
| `CORS_ALLOW_CREDENTIALS` | `true` | 是否允许凭证 |
| `CORS_ALLOW_METHODS` | `GET,POST,PUT,DELETE,OPTIONS` | 允许的HTTP方法 |
| `CORS_ALLOW_HEADERS` | `*` | 允许的HTTP头 |

### 安全配置

| 环境变量 | 默认值 | 描述 |
|---------|-------|------|
| `SECRET_KEY` | `your-secret-key-here` | 应用密钥 |
| `JWT_ALGORITHM` | `HS256` | JWT算法 |
| `JWT_EXPIRATION_MINUTES` | `30` | JWT过期时间（分钟） |

## 部署配置

### Docker部署配置

使用Docker Compose部署时，建议根据环境修改以下配置：

1. 在生产环境中，修改 `.env` 文件中的以下值：
   - `ENVIRONMENT=production`
   - `SERVER_WORKERS` 根据CPU核心数调整
   - 更改所有默认密码
   - 启用TLS加密

2. 调整Docker Compose中的资源限制：
   ```yaml
   services:
     doc-process-service:
       deploy:
         resources:
           limits:
             cpus: '4'
             memory: 8G
   ```

### 水平扩展配置

当需要水平扩展服务时，确保以下配置正确：

1. 所有服务实例使用相同的数据库和存储配置
2. Redis配置为所有实例共享
3. 设置适当的负载均衡

## 连接池调优

### 数据库连接池调优

根据并发用户数和查询复杂度调整连接池大小：

- 对于小型应用：`DATABASE_POOL_SIZE=10`
- 对于中型应用：`DATABASE_POOL_SIZE=20-50`
- 对于大型应用：`DATABASE_POOL_SIZE=50-100`

连接池大小计算公式：`(核心数 × 2) + 有效磁盘数`

### Redis连接池调优

Redis连接池大小应根据以下因素调整：

- 并发用户数
- 每个用户的Redis操作数
- 操作频率

一般来说，Redis连接池大小可以设置为数据库连接池大小的2-3倍。

## 常见配置问题排查

### 连接问题

1. **数据库连接失败**
   - 检查`DATABASE_URL`格式是否正确
   - 验证数据库服务是否运行
   - 检查网络连接和防火墙设置

2. **Redis连接失败**
   - 确认Redis服务端口是否正确
   - 检查密码和数据库索引
   - 验证TLS设置（如果启用）

3. **Milvus连接失败**
   - 检查Milvus服务地址和端口
   - 验证令牌（如果设置）
   - 确认etcd和MinIO服务是否正常运行

### 性能问题

1. **数据库性能慢**
   - 增加`DATABASE_POOL_SIZE`
   - 调整`DATABASE_POOL_RECYCLE`值
   - 检查数据库索引是否优化

2. **向量搜索性能差**
   - 调整`MILVUS_INDEX_PARAMS`中的`M`和`efConstruction`参数
   - 优化`MILVUS_SEARCH_PARAMS`中的`ef`参数
   - 考虑增加Milvus资源

## 示例配置文件

以下是一个完整的生产环境配置示例：

```dotenv
# 环境配置
ENVIRONMENT=production

# 服务器配置
SERVER_HOST=0.0.0.0
SERVER_PORT=8000
SERVER_WORKERS=4
SERVER_TIMEOUT=300

# 数据库配置
DATABASE_URL=postgresql://admin:StrongPassword123@postgres:5432/docu_sage
DATABASE_POOL_SIZE=30
DATABASE_MAX_OVERFLOW=15
DATABASE_POOL_RECYCLE=3600
DATABASE_ECHO=false

# Redis配置
REDIS_URL=redis://redis:6379/0
REDIS_MAX_CONNECTIONS=100
REDIS_MIN_IDLE_CONNECTIONS=10
REDIS_CONNECT_TIMEOUT=5
REDIS_SOCKET_TIMEOUT=5
REDIS_HEALTH_CHECK_INTERVAL=30
REDIS_TLS_ENABLED=true
REDIS_TLS_VERIFY_MODE=CERT_REQUIRED

# Milvus配置
MILVUS_URI=https://milvus:19530
MILVUS_TOKEN=your-milvus-token
MILVUS_COLLECTION_NAME=document_chunks
MILVUS_INDEX_TYPE=HNSW
MILVUS_METRIC_TYPE=COSINE
MILVUS_TLS_ENABLED=true
MILVUS_TIMEOUT=60
MILVUS_MAX_RETRIES=3

# MinIO配置
MINIO_ENDPOINT=minio:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=StrongPassword123
MINIO_SECURE=true
MINIO_BUCKET_NAME=docu-sage
MINIO_REGION=us-east-1

# 安全配置
SECRET_KEY=your-secret-key-here-change-in-production
JWT_EXPIRATION_MINUTES=60
```

## 配置备份与管理

建议定期备份配置文件，并使用配置管理工具进行版本控制。在生产环境中，考虑使用：

- AWS Secrets Manager
- HashiCorp Vault
- Kubernetes Secrets

这些工具可以提供更好的安全性和配置管理功能。