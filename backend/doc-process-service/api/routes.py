from fastapi import APIRouter, Depends, HTTPException, UploadFile, File, Form, BackgroundTasks
from sqlalchemy.orm import Session
from typing import List, Dict, Any, Optional
import uuid
import os
import logging

from models.db import get_db
from services.document_processing_service import document_processing_service
from api.document_controller import (
    upload_document,
    get_document_status,
    get_task_status,
    cancel_task,
    list_user_documents,
    get_document_details,
    delete_document,
    get_document_chunks,
    health_check
)

# 创建API路由器
api_router = APIRouter()

# 添加文档处理相关路由
api_router.post("/documents/upload", response_model=Dict[str, Any])(upload_document)
api_router.get("/documents/{document_id}/status", response_model=Dict[str, Any])(get_document_status)
api_router.get("/tasks/{task_id}/status", response_model=Dict[str, Any])(get_task_status)
api_router.post("/tasks/{task_id}/cancel", response_model=Dict[str, Any])(cancel_task)
api_router.get("/users/{user_id}/documents", response_model=List[Dict[str, Any]])(list_user_documents)
api_router.get("/documents/{document_id}", response_model=Dict[str, Any])(get_document_details)
api_router.delete("/documents/{document_id}", response_model=Dict[str, Any])(delete_document)
api_router.get("/documents/{document_id}/chunks", response_model=List[Dict[str, Any]])(get_document_chunks)
api_router.get("/health", response_model=Dict[str, Any])(health_check)

# 创建额外的API路由
def create_additional_routes(router: APIRouter):
    """创建额外的API路由"""
    
    @router.get("/api/version", response_model=Dict[str, str])
    async def get_api_version():
        """获取API版本信息"""
        return {"version": "1.0.0", "service": "document-processing-service"}
    
    @router.get("/api/metrics", response_model=Dict[str, Any])
    async def get_service_metrics(db: Session = Depends(get_db)):
        """获取服务指标"""
        try:
            from models.document import Document
            from models.processing_task import ProcessingTask
            
            # 获取文档统计
            document_stats = {
                "total_documents": db.query(Document).count(),
                "completed_documents": db.query(Document).filter(Document.status == "completed").count(),
                "processing_documents": db.query(Document).filter(Document.status == "processing").count(),
                "failed_documents": db.query(Document).filter(Document.status == "failed").count()
            }
            
            # 获取任务统计
            task_stats = {
                "total_tasks": db.query(ProcessingTask).count(),
                "completed_tasks": db.query(ProcessingTask).filter(ProcessingTask.status == "completed").count(),
                "processing_tasks": db.query(ProcessingTask).filter(ProcessingTask.status == "processing").count(),
                "failed_tasks": db.query(ProcessingTask).filter(ProcessingTask.status == "failed").count(),
                "canceled_tasks": db.query(ProcessingTask).filter(ProcessingTask.status == "canceled").count()
            }
            
            return {
                "documents": document_stats,
                "tasks": task_stats,
                "service_status": "healthy"
            }
        except Exception as e:
            logging.error(f"获取服务指标失败: {str(e)}")
            raise HTTPException(status_code=500, detail="获取服务指标失败")
    
    @router.post("/documents/batch/upload", response_model=Dict[str, Any])
    async def batch_upload_documents(
        files: List[UploadFile] = File(...),
        user_id: str = Form(...),
        background_tasks: BackgroundTasks = BackgroundTasks(),
        db: Session = Depends(get_db)
    ):
        """批量上传文档"""
        try:
            from tasks.document_tasks import batch_process_documents
            
            # 验证文件
            if not files or len(files) == 0:
                raise HTTPException(status_code=400, detail="请上传至少一个文件")
            
            if len(files) > 10:  # 限制批量上传数量
                raise HTTPException(status_code=400, detail="一次性最多上传10个文件")
            
            document_processing_service.validate_files(files)
            
            # 处理每个文件
            document_tasks = []
            for file in files:
                # 保存临时文件
                temp_path = os.path.join("/tmp", f"{uuid.uuid4()}_{file.filename}")
                with open(temp_path, "wb") as buffer:
                    content = await file.read()
                    buffer.write(content)
                
                document_tasks.append({
                    "file_path": temp_path,
                    "file_name": file.filename,
                    "user_id": user_id
                })
            
            # 提交批量处理任务
            task = batch_process_documents.apply_async(args=[document_tasks])
            
            return {
                "success": True,
                "batch_task_id": task.id,
                "file_count": len(files),
                "message": "批量处理任务已提交"
            }
        except HTTPException:
            raise
        except Exception as e:
            logging.error(f"批量上传文档失败: {str(e)}")
            raise HTTPException(status_code=500, detail=str(e))
    
    @router.post("/documents/{document_id}/reprocess", response_model=Dict[str, Any])
    async def reprocess_document(document_id: str, db: Session = Depends(get_db)):
        """重新处理文档"""
        try:
            # 检查文档是否存在
            from models.document import Document
            document = db.query(Document).filter(Document.id == document_id).first()
            if not document:
                raise HTTPException(status_code=404, detail="文档不存在")
            
            # 重新处理文档
            result = document_processing_service.reprocess_document(document_id)
            
            return {
                "success": True,
                "document_id": document_id,
                "task_id": result["task_id"],
                "message": "文档重新处理已提交"
            }
        except HTTPException:
            raise
        except Exception as e:
            logging.error(f"重新处理文档失败: {str(e)}")
            raise HTTPException(status_code=500, detail=str(e))
    
    @router.get("/search/documents", response_model=Dict[str, Any])
    async def search_documents(
        query: str,
        user_id: Optional[str] = None,
        limit: int = 10,
        offset: int = 0,
        db: Session = Depends(get_db)
    ):
        """搜索文档"""
        try:
            if not query or len(query.strip()) < 3:
                raise HTTPException(status_code=400, detail="搜索查询至少需要3个字符")
            
            results = document_processing_service.search_documents(
                query=query,
                user_id=user_id,
                limit=limit,
                offset=offset
            )
            
            return {
                "success": True,
                "results": results["documents"],
                "total": results["total"],
                "query": query,
                "limit": limit,
                "offset": offset
            }
        except HTTPException:
            raise
        except Exception as e:
            logging.error(f"搜索文档失败: {str(e)}")
            raise HTTPException(status_code=500, detail=str(e))
    
    @router.post("/admin/maintenance/cleanup", response_model=Dict[str, Any])
    async def perform_maintenance_cleanup(days: int = 7):
        """执行维护清理"""
        try:
            from tasks.document_tasks import cleanup_old_tasks
            
            task = cleanup_old_tasks.apply_async(args=[days])
            
            return {
                "success": True,
                "task_id": task.id,
                "message": f"清理任务已提交，将清理{days}天前的任务记录"
            }
        except Exception as e:
            logging.error(f"执行维护清理失败: {str(e)}")
            raise HTTPException(status_code=500, detail=str(e))

# 创建所有路由
create_additional_routes(api_router)