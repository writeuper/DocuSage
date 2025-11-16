import logging
import time
from sqlalchemy import create_engine, text
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import sessionmaker, Session
from sqlalchemy.pool import QueuePool
from contextlib import contextmanager

from config.config import settings

logger = logging.getLogger(__name__)

# 创建数据库引擎
engine = create_engine(
    settings.db_connection_url,
    poolclass=QueuePool,
    pool_pre_ping=True,
    pool_size=settings.db_max_idle_conns,
    max_overflow=settings.db_max_open_conns - settings.db_max_idle_conns,
    pool_recycle=settings.db_conn_max_lifetime * 60,  # 转换为秒
    pool_timeout=30,  # 连接池获取连接的超时时间（秒）
    echo=settings.debug  # 在调试模式下启用SQL回显
)

logger.info(
    f"Database engine created with pool configuration: "
    f"max_idle={settings.db_max_idle_conns}, "
    f"max_open={settings.db_max_open_conns}, "
    f"lifetime={settings.db_conn_max_lifetime}min, "
    f"idle_time={settings.db_conn_max_idle_time}min"
)

# 创建会话工厂
SessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)

# 创建基类
Base = declarative_base()


def get_db() -> Session:
    """获取数据库会话"""
    db = SessionLocal()
    try:
        yield db
    finally:
        db.close()


@contextmanager
def get_db_context() -> Session:
    """数据库会话上下文管理器"""
    db = SessionLocal()
    try:
        yield db
        db.commit()
    except Exception:
        db.rollback()
        raise
    finally:
        db.close()


def init_db():
    """初始化数据库，创建所有表"""
    try:
        # 导入所有模型，确保它们被注册
        from models import document, processing_task
        
        # 创建所有表
        Base.metadata.create_all(bind=engine)
        
        # 进行健康检查
        if check_db_health():
            logger.info("Database initialized and health check passed")
        else:
            raise RuntimeError("Database initialization failed: health check failed")
            
    except Exception as e:
        logger.error(f"Failed to initialize database: {e}")
        raise

def check_db_health() -> bool:
    """检查数据库连接健康状态"""
    try:
        with engine.connect() as conn:
            start_time = time.time()
            # 执行简单查询测试连接
            conn.execute(text("SELECT 1"))
            execution_time = (time.time() - start_time) * 1000  # 转换为毫秒
            logger.debug(f"Database health check passed in {execution_time:.2f}ms")
            return True
    except Exception as e:
        logger.error(f"Database health check failed: {e}")
        return False

def get_db_stats() -> dict:
    """获取数据库连接池统计信息"""
    pool = engine.pool
    stats = {
        "checkedin": pool.checkedin(),
        "checkedout": pool.checkedout(),
        "overflow": pool.overflow(),
        "size": pool.size(),
        "max_size": pool.max_size()
    }
    logger.debug(f"Database pool stats: {stats}")
    return stats


def close_db():
    """关闭数据库连接"""
    try:
        # 获取当前连接池状态用于日志记录
        pool_stats = get_db_stats()
        logger.info(f"Closing database connections, current pool stats: {pool_stats}")
        
        engine.dispose()
        logger.info("Database connection pool disposed successfully")
    except Exception as e:
        logger.error(f"Error closing database connection: {e}")