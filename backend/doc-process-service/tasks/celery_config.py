import os
from celery import Celery
from config.config import settings

# 创建Celery应用实例
celery_app = Celery(
    'document_processing',
    broker=settings.celery_broker_url,
    backend=settings.celery_result_backend
)

# 配置Celery
celery_app.conf.update(
    # 任务结果序列化格式
    result_serializer='json',
    # 接受的内容类型
    accept_content=['json'],
    # 任务序列化格式
    task_serializer='json',
    # 时区设置
    timezone=settings.timezone,
    # 启用UTC
    enable_utc=True,
    # 任务执行超时设置
    task_time_limit=settings.task_time_limit,
    # 任务软超时设置
    task_soft_time_limit=settings.task_soft_time_limit,
    # 并发设置
    worker_concurrency=settings.worker_concurrency,
    # 任务重试设置
    task_retry_backoff=True,
    task_retry_backoff_max=3600,
    task_max_retries=settings.max_retries,
    # 任务路由设置
    task_routes={
        'doc_process_service.tasks.document_tasks.process_document_task': {
            'queue': 'document_processing',
            'routing_key': 'document.process'
        },
        'doc_process_service.tasks.document_tasks.parse_document': {
            'queue': 'document_parsing',
            'routing_key': 'document.parse'
        },
        'doc_process_service.tasks.document_tasks.generate_embeddings': {
            'queue': 'embedding_generation',
            'routing_key': 'embedding.generate'
        }
    },
    # 任务跟踪设置
    task_track_started=True,
    # 结果过期时间（秒）
    result_expires=settings.result_expires,
    # 心跳设置
    broker_heartbeat=10,
    # 连接池设置
    broker_pool_limit=10,
    # 错误处理
    task_acks_late=True,
    worker_prefetch_multiplier=1
)

# 自动发现任务
celery_app.autodiscover_tasks(
    ['doc_process_service.tasks'],
    related_name='tasks'
)

# 配置任务超时和重试
celery_app.conf.task_default_rate_limit = settings.task_default_rate_limit

# 配置错误处理
celery_app.conf.on_failure = True

# 配置日志记录
if settings.log_level == 'DEBUG':
    celery_app.conf.worker_log_format = '%(asctime)s [%(levelname)s] [%(processName)s/%(process)d] [%(task_name)s(%(task_id)s)] %(message)s'
    celery_app.conf.worker_task_log_format = '%(asctime)s [%(levelname)s] [%(processName)s/%(process)d] [%(task_name)s(%(task_id)s)] %(message)s'