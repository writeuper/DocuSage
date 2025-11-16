# DocuSage - 文档处理服务

## 项目介绍

DocuSage 是一个先进的文档智能处理系统，专注于自动化文档解析、向量化和检索。本服务是系统的核心组件，负责文档的上传、解析、分块、向量化和存储管理。

## 技术架构

### 核心技术栈

- **后端框架**: FastAPI + Python 3.9
- **数据库**: PostgreSQL 15（结构化数据）
- **向量数据库**: Milvus 2.3.4（向量存储和相似度搜索）
- **缓存**: Redis 7（任务队列、缓存和限流）
- **对象存储**: MinIO（文档和中间文件存储）
- **消息队列**: Kafka 3.4（异步任务处理）
- **任务调度**: Celery（分布式任务处理）
- **容器化**: Docker + Docker Compose

### 系统架构图

```
┌─────────────┐     ┌─────────────────────┐     ┌─────────────┐
│  客户端应用  │────▶│  文档处理服务 API   │────▶│  查询服务    │
└─────────────┘     └───────────┬─────────┘     └─────────────┘
                               │
         ┌─────────────────────┼─────────────────────┐
         │                     │                     │
┌────────▼─────┐     ┌─────────▼──────────┐     ┌────▼───────────┐
│  PostgreSQL  │     │      Milvus       │     │     MinIO      │
│  (结构化数据) │     │   (向量数据库)    │     │  (对象存储)    │
└──────────────┘     └────────────────────┘     └────────────────┘
         │                     │                     │
         └─────────────────────┼─────────────────────┘
                               │
                      ┌────────▼─────────┐
                      │      Redis       │
                      │ (缓存、任务队列) │
                      └────────┬────────┘
                               │
                      ┌────────▼─────────┐
                      │      Kafka       │
                      │  (消息队列)      │
                      └────────┬────────┘
                               │
                      ┌────────▼─────────┐
                      │      Celery      │
                      │  (任务调度)      │
                      └──────────────────┘
```

## 功能特性

### 文档处理
- 支持多种格式文档上传（PDF、Word、Excel、PPT、TXT、Markdown等）
- 文档智能分块，保留语义完整性
- 并行文档处理，支持高并发

### 向量处理
- 文档内容向量化转换
- 向量相似度搜索
- 高效的向量索引管理

### 存储管理
- 文档元数据和内容分离存储
- 自动清理过期数据
- 支持文件版本管理

### 监控与安全
- 完善的健康检查机制
- 性能指标监控
- 访问控制和数据加密

## 快速开始

### 前置要求

- Docker 20.10+
- Docker Compose 1.29+
- Git

### 安装与部署

#### 1. 克隆代码库

```bash
git clone https://github.com/your-org/DocuSage.git
cd DocuSage/backend/doc-process-service
```

#### 2. 配置环境变量

```bash
cp .env.example .env
# 编辑.env文件，根据实际情况修改配置
```

#### 3. 使用Docker Compose启动服务

```bash
docker-compose up -d
```

这将启动所有必要的服务：
- 文档处理服务 (端口 8000)
- PostgreSQL数据库 (端口 5432)
- Redis缓存 (端口 6379)
- Milvus向量数据库 (端口 19530)
- MinIO对象存储 (端口 9000)
- Kafka消息队列 (端口 9092)
- Adminer数据库管理工具 (端口 8080)

#### 4. 验证服务是否正常运行

```bash
curl http://localhost:8000/health
```

## 核心API接口

### 文档上传与处理

- **POST /api/v1/documents/upload** - 上传新文档
- **GET /api/v1/documents/{document_id}** - 获取文档详情
- **DELETE /api/v1/documents/{document_id}** - 删除文档
- **GET /api/v1/documents** - 获取文档列表

### 文档处理状态

- **GET /api/v1/documents/{document_id}/status** - 获取文档处理状态
- **GET /api/v1/tasks/{task_id}** - 获取任务详情

### 搜索接口

- **POST /api/v1/search** - 执行向量搜索
- **POST /api/v1/search/hybrid** - 执行混合搜索（关键词+向量）

## 数据存储配置

### 1. 数据库配置

PostgreSQL数据库用于存储文档元数据、用户信息和处理任务信息。

**配置项**：
- `DATABASE_URL`: 数据库连接字符串
- `DATABASE_POOL_SIZE`: 数据库连接池大小
- `DATABASE_MAX_OVERFLOW`: 最大溢出连接数
- `DATABASE_POOL_RECYCLE`: 连接回收时间（秒）

### 2. 向量数据库配置

Milvus用于存储文档嵌入向量，支持高效的相似度搜索。

**配置项**：
- `MILVUS_URI`: Milvus服务地址
- `MILVUS_COLLECTION_NAME`: 集合名称
- `MILVUS_INDEX_TYPE`: 索引类型（推荐HNSW）
- `MILVUS_METRIC_TYPE`: 距离度量类型（COSINE/IP/L2）
- `MILVUS_TLS_ENABLED`: 是否启用TLS

### 3. 对象存储配置

MinIO用于存储原始文档文件和处理过程中的中间文件。

**配置项**：
- `MINIO_ENDPOINT`: MinIO服务地址
- `MINIO_ACCESS_KEY`: 访问密钥
- `MINIO_SECRET_KEY`: 密钥
- `MINIO_BUCKET_NAME`: 存储桶名称
- `MINIO_SECURE`: 是否使用HTTPS

### 4. 缓存配置

Redis用于缓存热点数据、管理任务队列和实现限流。

**配置项**：
- `REDIS_URL`: Redis连接地址
- `REDIS_MAX_CONNECTIONS`: 最大连接数
- `REDIS_MIN_IDLE_CONNECTIONS`: 最小空闲连接数
- `REDIS_CONNECT_TIMEOUT`: 连接超时（秒）
- `REDIS_HEALTH_CHECK_INTERVAL`: 健康检查间隔（秒）

## 性能优化

### 连接池配置

所有数据库和缓存连接均使用连接池，配置合理的连接池大小可以显著提升性能：

- 调整`DATABASE_POOL_SIZE`根据并发量
- 调整`REDIS_MAX_CONNECTIONS`以避免连接过多
- 监控连接池使用情况，及时调整参数

### 向量索引优化

Milvus向量索引配置对搜索性能至关重要：

- 使用HNSW索引类型获取最佳的搜索性能
- 调整索引参数以平衡查询速度和准确性
- 定期重建索引以保持查询效率

### 文件处理优化

- 使用异步处理大量文件
- 调整`CHUNK_SIZE`和`CHUNK_OVERLAP`参数以优化文本分块
- 启用文件压缩减少存储空间

## 监控与维护

### 健康检查

系统提供全面的健康检查端点：

- `/health` - 整体服务健康状态
- `/health/db` - 数据库连接健康状态
- `/health/redis` - Redis连接健康状态
- `/health/milvus` - Milvus连接健康状态
- `/health/minio` - MinIO连接健康状态

### 日志管理

- 日志存储在`./logs`目录
- 可配置日志级别（DEBUG/INFO/WARNING/ERROR）
- 支持结构化日志格式

### 常见问题排查

1. **数据库连接失败**
   - 检查数据库服务是否运行
   - 验证连接字符串和凭证
   - 检查网络连接和防火墙设置

2. **向量搜索性能差**
   - 检查索引是否正确创建
   - 优化索引参数
   - 考虑增加Milvus资源

3. **文件上传失败**
   - 检查MinIO服务状态
   - 验证文件大小是否超过限制
   - 检查存储空间是否充足

## 开发指南

### 本地开发环境设置

1. 安装Python 3.9和Poetry

2. 安装依赖

```bash
poetry install
```

3. 启动开发服务器

```bash
poetry run uvicorn main:app --reload --host 0.0.0.0 --port 8000
```

### 代码贡献

- 遵循PEP 8编码规范
- 为新功能编写单元测试
- 提交代码前运行测试

```bash
poetry run pytest
```

## 部署最佳实践

### 生产环境配置

1. 使用专用的环境变量文件
2. 启用TLS加密所有服务间通信
3. 配置适当的资源限制
4. 设置自动备份策略

### 扩展建议

- 水平扩展文档处理服务以支持更多并发
- 增加Milvus节点提升向量搜索性能
- 使用Redis集群提高缓存容量和可靠性

## 许可证

[MIT License](LICENSE)

## 联系方式

- 项目维护者: [team@docusage.com](mailto:team@docusage.com)
- 问题反馈: [GitHub Issues](https://github.com/your-org/DocuSage/issues)

---

*DocuSage - 智能文档处理系统*