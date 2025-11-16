import os
import logging
from fastapi import FastAPI, Request
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse
import time

# 配置日志
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler("app.log"),
        logging.StreamHandler()
    ]
)

logger = logging.getLogger(__name__)

# 导入配置和路由
from config.config import settings, setup_directories, get_settings
from api.routes import api_router
from models.db import init_db, close_db

# 创建FastAPI应用实例
app = FastAPI(
    title="Document Processing Service API",
    description="文档处理服务API - 用于上传、解析和处理文档",
    version="1.0.0",
    docs_url="/docs",
    redoc_url="/redoc"
)

# 配置CORS
app.add_middleware(
    CORSMiddleware,
    allow_origins=settings.cors_origins,
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# 中间件：请求处理时间
@app.middleware("http")
async def add_process_time_header(request: Request, call_next):
    start_time = time.time()
    response = await call_next(request)
    process_time = time.time() - start_time
    response.headers["X-Process-Time"] = str(process_time)
    return response

# 全局异常处理
@app.exception_handler(Exception)
async def global_exception_handler(request: Request, exc: Exception):
    logger.error(f"全局异常: {str(exc)}", exc_info=True)
    return JSONResponse(
        status_code=500,
        content={
            "success": False,
            "error": "内部服务器错误",
            "message": "服务器内部发生错误，请稍后重试",
            "timestamp": time.time()
        }
    )

# 添加API路由
app.include_router(api_router, prefix="", tags=["document-processing"])

# 启动事件
@app.on_event("startup")
async def startup_event():
    logger.info("服务启动中...")
    
    try:
        # 设置目录结构
        setup_directories()
        logger.info("目录结构设置完成")
        
        # 初始化数据库
        init_db()
        logger.info("数据库初始化完成")
        
        # 检查配置
        config = get_settings()
        logger.info(f"配置加载完成: 环境={config.environment}")
        
        # 检查必要的目录权限
        check_directory_permissions()
        
        logger.info("服务启动成功")
    except Exception as e:
        logger.error(f"服务启动失败: {str(e)}", exc_info=True)
        raise

# 关闭事件
@app.on_event("shutdown")
async def shutdown_event():
    logger.info("服务关闭中...")
    try:
        # 关闭数据库连接
        close_db()
        logger.info("数据库连接已关闭")
    except Exception as e:
        logger.error(f"关闭资源时出错: {str(e)}")
    logger.info("服务已关闭")

# 根路径
@app.get("/")
async def root():
    return {
        "message": "文档处理服务API",
        "version": "1.0.0",
        "docs": "/docs",
        "status": "running"
    }

# 健康检查（额外的端点，除了路由中的）
@app.get("/healthz")
async def healthz():
    return {
        "status": "healthy",
        "timestamp": time.time(),
        "service": "document-processing-service"
    }

# 检查目录权限
def check_directory_permissions():
    """检查必要目录的权限"""
    directories_to_check = [
        settings.document_storage_path,
        settings.document_storage_path + "/uploads",
        settings.document_storage_path + "/processed",
        settings.log_directory
    ]
    
    for directory in directories_to_check:
        if not os.path.exists(directory):
            try:
                os.makedirs(directory, exist_ok=True)
                logger.info(f"创建目录: {directory}")
            except Exception as e:
                logger.error(f"无法创建目录 {directory}: {str(e)}")
                raise
        
        if not os.access(directory, os.W_OK):
            logger.error(f"目录没有写权限: {directory}")
            raise PermissionError(f"目录没有写权限: {directory}")
        
        if not os.access(directory, os.R_OK):
            logger.error(f"目录没有读权限: {directory}")
            raise PermissionError(f"目录没有读权限: {directory}")

# 主函数
if __name__ == "__main__":
    import uvicorn
    
    try:
        uvicorn.run(
            "main:app",
            host=settings.host,
            port=settings.port,
            reload=settings.debug,
            log_level="info"
        )
    except KeyboardInterrupt:
        logger.info("服务被用户中断")
    except Exception as e:
        logger.error(f"服务运行失败: {str(e)}", exc_info=True)