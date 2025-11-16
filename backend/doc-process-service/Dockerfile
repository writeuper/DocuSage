FROM python:3.9-slim

# 设置工作目录
WORKDIR /app

# 设置环境变量
ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1 \
    POETRY_VERSION=1.7.1

# 安装系统依赖
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    curl \
    git \
    gcc \
    libpq-dev \
    && rm -rf /var/lib/apt/lists/*

# 安装Poetry
RUN curl -sSL https://install.python-poetry.org | python3 -

# 将Poetry添加到PATH
ENV PATH="/root/.local/bin:$PATH"

# 配置Poetry（不创建虚拟环境）
RUN poetry config virtualenvs.create false

# 复制pyproject.toml和poetry.lock（如果存在）
COPY pyproject.toml poetry.lock* ./

# 安装依赖
RUN poetry install --no-interaction --no-ansi

# 复制应用代码
COPY . .

# 确保start.sh有执行权限
RUN chmod +x /app/start.sh

# 暴露服务端口
EXPOSE 8000

# 健康检查
HEALTHCHECK --interval=30s --timeout=30s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8000/health || exit 1

# 启动服务
CMD ["/app/start.sh"]