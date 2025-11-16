# 开发指南

本文档提供了DocuSage文档处理服务的开发环境配置、代码结构说明和贡献指南。

## 环境要求

- **Python**: 3.9 或更高版本
- **数据库**: PostgreSQL 13+ (用于关系数据)
- **向量数据库**: Milvus 2.2+ (用于向量存储和搜索)
- **对象存储**: MinIO (用于文档文件存储)
- **缓存**: Redis 6.0+ (用于缓存和队列)
- **消息队列**: Apache Kafka (用于异步任务)
- **任务队列**: Celery 5.2+ (用于后台任务处理)

## 开发环境设置

### 1. 克隆代码库

```bash
git clone https://github.com/your-org/DocuSage.git
cd DocuSage/backend/doc-process-service
```

### 2. 安装依赖管理工具

```bash
# 安装 Poetry
curl -sSL https://install.python-poetry.org | python3 -

# 或使用 pip
pip install poetry
```

### 3. 安装项目依赖

```bash
poetry install
```

### 4. 配置环境变量

复制示例环境配置文件并根据需要修改：

```bash
cp .env.example .env
```

编辑 `.env` 文件，配置以下关键参数：

```dotenv
# 数据库连接信息
DATABASE_URL="postgresql://admin:password@localhost:5432/docu_sage"

# Redis连接信息
REDIS_URL="redis://localhost:6379/0"

# Milvus连接信息
MILVUS_HOST="localhost"
MILVUS_PORT="19530"
MILVUS_USERNAME="root"
MILVUS_PASSWORD="Milvus"

# MinIO连接信息
MINIO_ENDPOINT="localhost:9000"
MINIO_ACCESS_KEY="minioadmin"
MINIO_SECRET_KEY="minioadmin"
MINIO_BUCKET="docu-sage"

# Kafka连接信息
KAFKA_BOOTSTRAP_SERVERS="localhost:9092"

# Celery配置
CELERY_BROKER_URL="redis://localhost:6379/1"
CELERY_RESULT_BACKEND="redis://localhost:6379/2"
```

### 5. 使用Docker Compose启动依赖服务

```bash
docker-compose up -d postgres redis milvus minio kafka
```

等待所有服务启动完成，然后可以验证连接。

### 6. 初始化数据库

```bash
poetry run python -m models.db init
```

### 7. 启动开发服务器

```bash
poetry run uvicorn app.main:app --host 0.0.0.0 --port 8000 --reload
```

开发服务器启动后，可以访问以下地址：
- API文档: http://localhost:8000/docs
- ReDoc文档: http://localhost:8000/redoc
- 健康检查: http://localhost:8000/health

### 8. 启动Celery worker

在另一个终端中：

```bash
poetry run celery -A app.tasks worker --loglevel=info --concurrency=4
```

## 代码结构

```
doc-process-service/
├── app/                    # 应用主目录
│   ├── api/                # API路由和端点
│   │   ├── __init__.py
│   │   ├── health.py       # 健康检查接口
│   │   ├── documents.py    # 文档管理接口
│   │   ├── search.py       # 搜索接口
│   │   ├── tasks.py        # 任务管理接口
│   │   └── stats.py        # 统计接口
│   ├── core/               # 核心功能
│   │   ├── __init__.py
│   │   ├── config.py       # 配置管理
│   │   ├── security.py     # 安全相关
│   │   └── dependencies.py # 依赖注入
│   ├── services/           # 业务逻辑层
│   │   ├── __init__.py
│   │   ├── document_service.py
│   │   ├── search_service.py
│   │   └── task_service.py
│   ├── tasks/              # Celery任务
│   │   ├── __init__.py
│   │   └── document_tasks.py
│   ├── models/             # 数据模型
│   │   ├── __init__.py
│   │   ├── db.py           # 数据库连接
│   │   └── schemas/        # Pydantic模型
│   ├── utils/              # 工具函数
│   │   ├── __init__.py
│   │   ├── redis_client.py
│   │   ├── milvus_client.py
│   │   ├── minio_client.py
│   │   ├── document_parser.py
│   │   └── chunker.py
│   └── main.py             # 应用入口
├── config/                 # 配置文件
│   └── config.py
├── docs/                   # 文档目录
│   ├── api_documentation.md
│   ├── configuration_guide.md
│   └── development_guide.md
├── tests/                  # 测试目录
│   ├── unit/               # 单元测试
│   ├── integration/        # 集成测试
│   └── conftest.py
├── .env.example            # 环境变量示例
├── Dockerfile              # Docker构建文件
├── docker-compose.yml      # Docker Compose配置
├── pyproject.toml          # Poetry项目配置
├── poetry.lock             # 依赖版本锁定
└── README.md               # 项目说明
```

## 关键模块说明

### 1. 文档处理流程

文档处理的主要流程如下：

1. **上传文档**：通过`/api/v1/documents/upload`接口上传文档
2. **文档解析**：使用`DocumentParser`解析不同格式的文档
3. **文本分块**：使用`Chunker`将文档切分为适当大小的块
4. **向量嵌入**：生成文档块的向量表示
5. **存储**：
   - 文档元数据存储在PostgreSQL
   - 文档块内容存储在PostgreSQL
   - 向量嵌入存储在Milvus
   - 原始文件存储在MinIO
6. **索引**：在Milvus中建立向量索引

### 2. 向量搜索实现

向量搜索主要通过`SearchService`实现，支持：
- 相似性搜索（基于余弦相似度）
- 混合搜索（结合关键词和向量搜索）
- 过滤搜索（基于元数据过滤）

### 3. 异步任务处理

系统使用Celery和Kafka处理异步任务：
- 文档处理任务异步执行
- 长时间运行的任务不会阻塞API响应
- 任务状态和进度通过Redis缓存

## 开发工作流

### 代码风格

项目使用以下工具确保代码质量：

- **Black**: 代码格式化
- **isort**: 导入排序
- **flake8**: 代码检查
- **mypy**: 静态类型检查

运行代码检查：

```bash
poetry run black .
poetry run isort .
poetry run flake8
poetry run mypy
```

### 编写测试

为新功能编写测试是贡献流程的重要部分：

1. 单元测试放在`tests/unit/`目录
2. 集成测试放在`tests/integration/`目录

运行测试：

```bash
poetry run pytest
```

### 代码审查

提交Pull Request前，请确保：
1. 代码通过所有测试
2. 代码符合项目的代码风格
3. 添加了必要的文档注释
4. 提供了测试用例

## 贡献指南

### 报告问题

使用GitHub Issues报告问题时，请包括：

1. 问题描述
2. 重现步骤
3. 期望行为
4. 实际行为
5. 环境信息（操作系统、Python版本等）
6. 错误日志（如果有）

### 提交功能请求

提交功能请求时，请详细描述：

1. 功能概述
2. 预期使用场景
3. 对现有功能的影响（如果有）
4. 可能的实现方法（可选）

### Pull Request流程

1. Fork仓库
2. 创建功能分支
3. 实现功能或修复问题
4. 编写测试
5. 运行代码检查和测试
6. 提交Pull Request

## 部署指南

### 使用Docker部署

```bash
docker-compose up -d
```

这将启动所有必要的服务，包括文档处理服务、PostgreSQL、Redis、Milvus、MinIO和Kafka。

### 环境变量配置

在生产环境中，请确保正确配置以下关键环境变量：

- `ENVIRONMENT=production`
- `DEBUG=false`
- 使用强密码保护数据库和其他服务
- 配置适当的连接池大小
- 配置TLS/SSL（如果需要）

### 性能优化

生产环境建议的优化措施：

1. 增加Celery worker数量以处理更多并发任务
2. 调整数据库连接池大小
3. 为Milvus配置适当的索引参数
4. 配置MinIO的缓存和性能选项
5. 使用Redis集群提高缓存性能

## 监控与日志

### 日志级别

可以通过环境变量`LOG_LEVEL`设置日志级别：
- DEBUG
- INFO (默认)
- WARNING
- ERROR
- CRITICAL

### 健康检查

系统提供了多个健康检查端点：
- `/health`: 整体服务健康状态
- `/health/db`: 数据库连接状态
- `/health/redis`: Redis连接状态
- `/health/milvus`: Milvus连接状态
- `/health/minio`: MinIO连接状态

### 性能监控

推荐使用以下工具监控系统性能：

- Prometheus + Grafana: 监控系统指标
- ELK Stack: 日志聚合和分析
- New Relic或Datadog: 应用性能监控

## 常见问题排查

### 1. 文档上传失败

- 检查文件大小是否超过限制
- 验证文件类型是否支持
- 检查MinIO服务是否正常运行
- 查看服务日志获取详细错误信息

### 2. 文档处理任务卡在某个状态

- 检查Celery worker是否运行
- 验证Kafka服务是否正常
- 查看Redis连接状态
- 检查任务日志获取详细错误

### 3. 搜索结果不准确

- 检查向量索引是否正确创建
- 验证嵌入模型配置是否正确
- 调整搜索参数（如top_k、相似度阈值）
- 检查Milvus连接状态

### 4. 服务性能问题

- 监控数据库连接池使用情况
- 检查Redis内存使用
- 优化Celery worker配置
- 调整Milvus查询参数

## 资源与链接

- [FastAPI文档](https://fastapi.tiangolo.com/)
- [SQLAlchemy文档](https://docs.sqlalchemy.org/)
- [Milvus文档](https://milvus.io/docs/)
- [MinIO文档](https://min.io/docs/)
- [Redis文档](https://redis.io/docs/)
- [Celery文档](https://docs.celeryq.dev/)
- [Kafka文档](https://kafka.apache.org/documentation/)

## 许可证

[在此添加许可证信息]