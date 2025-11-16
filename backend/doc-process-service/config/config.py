import os
from typing import List, Dict, Any
from pydantic_settings import BaseSettings
from functools import lru_cache


class Settings(BaseSettings):
    # 环境配置
    environment: str = "development"
    debug: bool = True
    
    # 服务器配置
    server_host: str = "0.0.0.0"
    server_port: int = 8088
    server_read_timeout: int = 30
    server_write_timeout: int = 30
    server_idle_timeout: int = 60
    server_shutdown_timeout: int = 10
    
    # 数据库配置
    db_host: str = "localhost"
    db_port: int = 3306
    db_user: str = "admin"
    db_password: str = "password"
    db_name: str = "docu_sage"
    # 数据库连接池配置
    db_max_idle_conns: int = 10
    db_max_open_conns: int = 100
    db_conn_max_lifetime: int = 30  # 分钟
    db_conn_max_idle_time: int = 10  # 分钟
    
    # Redis配置
    redis_host: str = "localhost"
    redis_port: int = 6379
    redis_password: str = ""
    redis_db: int = 0
    # Redis连接池配置
    redis_pool_size: int = 100
    redis_min_idle_conns: int = 10
    redis_max_idle_conns: int = 20
    redis_dial_timeout: int = 5000  # 毫秒
    redis_read_timeout: int = 3000  # 毫秒
    redis_write_timeout: int = 3000  # 毫秒
    redis_idle_timeout: int = 60  # 秒
    redis_max_retries: int = 3
    redis_enable_tls: bool = False
    redis_skip_tls_verify: bool = False
    
    # Kafka配置
    kafka_bootstrap_servers: List[str] = ["localhost:9092"]
    kafka_input_topic: str = "document_uploads"
    kafka_output_topic: str = "document_processed"
    
    # Celery配置
    celery_broker_url: str = "redis://localhost:6379/1"
    celery_result_backend: str = "redis://localhost:6379/2"
    
    # MinIO对象存储配置
    minio_endpoint: str = "localhost:9000"
    minio_access_key: str = "minioadmin"
    minio_secret_key: str = "minioadmin"
    minio_bucket_name: str = "docu-sage-documents"
    minio_use_ssl: bool = False
    minio_region: str = "us-east-1"
    
    # 文件存储配置
    upload_dir: str = "/tmp/uploads"
    max_file_size_mb: int = 100
    allowed_extensions: List[str] = [
        ".pdf", ".doc", ".docx", ".txt", ".md", ".html", ".xml"
    ]
    
    # Milvus向量数据库配置
    milvus_host: str = "localhost"
    milvus_port: int = 19530
    milvus_collection_name: str = "document_chunks"
    milvus_dim: int = 1536  # 嵌入向量维度
    milvus_metric_type: str = "IP"  # 内积相似度
    milvus_consistency_level: str = "Strong"
    milvus_timeout: int = 30  # 秒
    
    # 处理配置
    chunk_size: int = 1000  # 文本分块大小
    overlap_size: int = 100  # 块重叠大小
    max_workers: int = 4  # 处理工作线程数
    
    # 嵌入服务配置
    embed_service_url: str = "http://localhost:8089/embed"
    
    # 向量搜索配置
    vector_service_url: str = "http://localhost:8087"
    vector_search_top_k: int = 5
    vector_search_timeout: int = 5  # 秒
    vector_search_threshold: float = 0.75  # 相似度阈值
    
    # JWT配置
    jwt_secret_key: str = "your-secret-key-change-in-production"
    jwt_algorithm: str = "HS256"
    jwt_expiration_minutes: int = 30
    
    # 日志配置
    log_level: str = "INFO"
    log_format: str = "%(asctime)s - %(name)s - %(levelname)s - %(message)s"
    
    # CORS配置
    cors_origins: List[str] = ["http://localhost:3000", "http://localhost:3001"]
    
    # 性能和限流配置
    enable_rate_limit: bool = True
    rate_limit_per_min: int = 100
    
    # 安全配置
    secure_cookies: bool = True
    same_site: str = "lax"
    
    # 监控配置
    enable_metrics: bool = True
    
    @property
    def db_connection_url(self) -> str:
        return f"mysql+pymysql://{self.db_user}:{self.db_password}@{self.db_host}:{self.db_port}/{self.db_name}?charset=utf8mb4"
    
    @property
    def redis_url(self) -> str:
        if self.redis_password:
            return f"redis://:{self.redis_password}@{self.redis_host}:{self.redis_port}/{self.redis_db}"
        return f"redis://{self.redis_host}:{self.redis_port}/{self.redis_db}"
    
    @property
    def milvus_uri(self) -> str:
        return f"http://{self.milvus_host}:{self.milvus_port}"
    
    @property
    def minio_url(self) -> str:
        protocol = "https" if self.minio_use_ssl else "http"
        return f"{protocol}://{self.minio_endpoint}"
    
    class Config:
        env_file = ".env"
        case_sensitive = False


@lru_cache()
def get_settings() -> Settings:
    """获取配置单例"""
    return Settings()


def setup_directories():
    """设置必要的目录"""
    settings = get_settings()
    # 创建上传目录
    if not os.path.exists(settings.upload_dir):
        os.makedirs(settings.upload_dir)


# 导出默认配置实例
settings = get_settings()