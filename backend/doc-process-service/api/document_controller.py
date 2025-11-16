import logging
import os
from typing import Dict, List, Any, Optional
from datetime import datetime
from fastapi import APIRouter, UploadFile, File, Form, Depends, HTTPException, status
from fastapi.responses import JSONResponse
from sqlalchemy.orm import Session

from config.config import settings
from models.db import get_db
from services.document_processing_service import document_processing_service
from utils.file_handler import file_handler
from utils.redis_client import redis_client

logger = logging.getLogger(__name__)

# 创建路由器
router = APIRouter(prefix="/api/documents", tags=["documents"])


@router.post("/upload", response_model=Dict[str, Any])
async def upload_document(
    file: UploadFile = File(...),
    user_id: str = Form(...),
    description: Optional[str] = Form(None),
    category: Optional[str] = Form(None)
) -> Dict[str, Any]:
    """上传并处理文档"""
    try:
        # 验证文件类型
        if not file_handler.is_allowed_file(file.filename):
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail=f"不支持的文件类型。允许的类型: {', '.join(settings.allowed_file_types)}"
            )
        
        # 验证文件大小
        contents = await file.read()
        if len(contents) > settings.max_file_size:
            raise HTTPException(
                status_code=status.HTTP_413_REQUEST_ENTITY_TOO_LARGE,
                detail=f"文件大小超过限制。最大允许大小: {settings.max_file_size / 1024 / 1024} MB"
            )
        
        # 保存临时文件
        temp_dir = os.path.join(settings.document_storage_path, "temp")
        os.makedirs(temp_dir, exist_ok=True)
        temp_file_path = os.path.join(temp_dir, file.filename)
        
        with open(temp_file_path, "wb") as buffer:
            buffer.write(contents)
        
        try:
            # 异步处理文档
            result = await document_processing_service.process_document(
                temp_file_path,
                file.filename,
                user_id
            )
            
            if result["success"]:
                return JSONResponse(
                    status_code=status.HTTP_202_ACCEPTED,
                    content={
                        "message": "文档已接受并正在处理",
                        "document_id": result["document_id"],
                        "task_id": result["task_id"],
                        "file_name": result["file_name"],
                        "status": "processing"
                    }
                )
            else:
                raise HTTPException(
                    status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
                    detail=f"文档处理失败: {result.get('error', '未知错误')}"
                )
                
        finally:
            # 清理临时文件
            if os.path.exists(temp_file_path):
                try:
                    os.remove(temp_file_path)
                except Exception as e:
                    logger.error(f"清理临时文件失败: {str(e)}")
                    
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"上传文档时出错: {str(e)}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"文档上传失败: {str(e)}"
        )


@router.get("/status/{document_id}", response_model=Dict[str, Any])
async def get_document_status(document_id: str) -> Dict[str, Any]:
    """获取文档处理状态"""
    try:
        # 先从缓存获取
        cached_status = redis_client.get_document_status(document_id)
        if cached_status:
            return {
                "document_id": document_id,
                "status": cached_status["status"],
                "updated_at": cached_status.get("updated_at", datetime.now().isoformat()),
                "metadata": {k: v for k, v in cached_status.items() if k not in ["status", "updated_at"]}
            }
        
        # 如果缓存中没有，从数据库查询相关任务
        from models.processing_task import ProcessingTask
        from models.db import get_db_context
        
        with get_db_context() as db:
            task = db.query(ProcessingTask).filter(
                ProcessingTask.document_id == document_id
            ).first()
            
            if not task:
                raise HTTPException(
                    status_code=status.HTTP_404_NOT_FOUND,
                    detail="文档不存在"
                )
            
            return {
                "document_id": document_id,
                "status": task.status,
                "updated_at": task.updated_at.isoformat(),
                "file_name": task.file_name,
                "error_message": task.error_message
            }
            
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"获取文档状态时出错: {str(e)}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"获取文档状态失败: {str(e)}"
        )


@router.get("/task/{task_id}", response_model=Dict[str, Any])
async def get_task_status(task_id: str) -> Dict[str, Any]:
    """获取处理任务状态"""
    try:
        status = await document_processing_service.get_task_status(task_id)
        
        if not status:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail="任务不存在"
            )
        
        return status
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"获取任务状态时出错: {str(e)}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"获取任务状态失败: {str(e)}"
        )


@router.post("/task/{task_id}/cancel", response_model=Dict[str, Any])
async def cancel_task(task_id: str) -> Dict[str, Any]:
    """取消处理任务"""
    try:
        success = await document_processing_service.cancel_task(task_id)
        
        if not success:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail="任务无法取消（任务不存在或已完成/失败）"
            )
        
        return {
            "message": "任务已取消",
            "task_id": task_id
        }
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"取消任务时出错: {str(e)}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"取消任务失败: {str(e)}"
        )


@router.get("/user/{user_id}", response_model=Dict[str, Any])
async def list_user_documents(
    user_id: str,
    skip: int = 0,
    limit: int = 100,
    db: Session = Depends(get_db)
) -> Dict[str, Any]:
    """列出用户的所有文档"""
    try:
        from models.document import Document
        
        # 查询用户文档
        documents = db.query(Document).filter(
            Document.user_id == user_id
        ).order_by(Document.created_at.desc()).offset(skip).limit(limit).all()
        
        # 获取总数
        total = db.query(Document).filter(Document.user_id == user_id).count()
        
        # 格式化结果
        result = []
        for doc in documents:
            result.append({
                "id": doc.id,
                "file_name": doc.file_name,
                "status": doc.status,
                "total_pages": doc.total_pages,
                "created_at": doc.created_at.isoformat(),
                "updated_at": doc.updated_at.isoformat()
            })
        
        return {
            "documents": result,
            "total": total,
            "skip": skip,
            "limit": limit
        }
        
    except Exception as e:
        logger.error(f"列出用户文档时出错: {str(e)}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"获取文档列表失败: {str(e)}"
        )


@router.get("/{document_id}", response_model=Dict[str, Any])
async def get_document_details(
    document_id: str,
    db: Session = Depends(get_db)
) -> Dict[str, Any]:
    """获取文档详细信息"""
    try:
        from models.document import Document, DocumentStats
        
        # 查询文档
        document = db.query(Document).filter(Document.id == document_id).first()
        
        if not document:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail="文档不存在"
            )
        
        # 查询统计信息
        stats = db.query(DocumentStats).filter(
            DocumentStats.document_id == document_id
        ).first()
        
        # 构建响应
        result = {
            "id": document.id,
            "file_name": document.file_name,
            "user_id": document.user_id,
            "status": document.status,
            "total_pages": document.total_pages,
            "metadata": document.metadata,
            "created_at": document.created_at.isoformat(),
            "updated_at": document.updated_at.isoformat()
        }
        
        if stats:
            result["stats"] = {
                "total_chunks": stats.total_chunks,
                "total_tokens": stats.total_tokens,
                "total_words": stats.total_words,
                "updated_at": stats.updated_at.isoformat()
            }
        
        return result
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"获取文档详情时出错: {str(e)}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"获取文档详情失败: {str(e)}"
        )


@router.delete("/{document_id}", response_model=Dict[str, Any])
async def delete_document(
    document_id: str,
    db: Session = Depends(get_db)
) -> Dict[str, Any]:
    """删除文档"""
    try:
        from models.document import Document, DocumentChunk, ChunkMetadata, DocumentStats
        from models.processing_task import ProcessingTask, SubTask, TaskQueue, ProcessingLog
        
        # 开启事务
        try:
            # 删除文档相关数据
            # 1. 删除文档分块和元数据
            chunks = db.query(DocumentChunk).filter(
                DocumentChunk.document_id == document_id
            ).all()
            
            for chunk in chunks:
                if chunk.metadata_id:
                    metadata = db.query(ChunkMetadata).filter(
                        ChunkMetadata.id == chunk.metadata_id
                    ).first()
                    if metadata:
                        db.delete(metadata)
                db.delete(chunk)
            
            # 2. 删除文档统计信息
            stats = db.query(DocumentStats).filter(
                DocumentStats.document_id == document_id
            ).first()
            if stats:
                db.delete(stats)
            
            # 3. 查询并删除相关任务
            task = db.query(ProcessingTask).filter(
                ProcessingTask.document_id == document_id
            ).first()
            
            if task:
                # 删除子任务
                db.query(SubTask).filter(SubTask.task_id == task.id).delete()
                # 删除处理日志
                db.query(ProcessingLog).filter(ProcessingLog.task_id == task.id).delete()
                # 删除队列项
                queue_item = db.query(TaskQueue).filter(TaskQueue.task_id == task.id).first()
                if queue_item:
                    db.delete(queue_item)
                # 删除任务
                db.delete(task)
            
            # 4. 删除文档
            document = db.query(Document).filter(Document.id == document_id).first()
            if not document:
                raise HTTPException(
                    status_code=status.HTTP_404_NOT_FOUND,
                    detail="文档不存在"
                )
            
            db.delete(document)
            db.commit()
            
            # 5. 删除缓存
            redis_client.delete(f"doc_status:{document_id}")
            
            return {
                "message": "文档已成功删除",
                "document_id": document_id
            }
            
        except HTTPException:
            db.rollback()
            raise
        except Exception as e:
            db.rollback()
            logger.error(f"删除文档事务失败: {str(e)}")
            raise HTTPException(
                status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
                detail=f"删除文档失败: {str(e)}"
            )
            
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"删除文档时出错: {str(e)}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"删除文档失败: {str(e)}"
        )


@router.get("/chunks/{document_id}", response_model=Dict[str, Any])
async def get_document_chunks(
    document_id: str,
    skip: int = 0,
    limit: int = 100,
    db: Session = Depends(get_db)
) -> Dict[str, Any]:
    """获取文档分块"""
    try:
        from models.document import DocumentChunk, ChunkMetadata
        
        # 查询文档是否存在
        chunk_count = db.query(DocumentChunk).filter(
            DocumentChunk.document_id == document_id
        ).count()
        
        if chunk_count == 0:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail="文档不存在或没有分块"
            )
        
        # 查询分块
        chunks_with_metadata = db.query(DocumentChunk, ChunkMetadata).join(
            ChunkMetadata, DocumentChunk.metadata_id == ChunkMetadata.id,
            isouter=True
        ).filter(
            DocumentChunk.document_id == document_id
        ).order_by(DocumentChunk.created_at).offset(skip).limit(limit).all()
        
        # 格式化结果
        result = []
        for chunk, metadata in chunks_with_metadata:
            chunk_data = {
                "id": chunk.id,
                "content": chunk.content,
                "embedding_model": chunk.embedding_model,
                "created_at": chunk.created_at.isoformat()
            }
            
            if metadata:
                chunk_data["metadata"] = {
                    "page_number": metadata.page_number,
                    "start_pos": metadata.start_pos,
                    "end_pos": metadata.end_pos,
                    "word_count": metadata.word_count,
                    "token_count": metadata.token_count
                }
            
            result.append(chunk_data)
        
        return {
            "chunks": result,
            "total": chunk_count,
            "skip": skip,
            "limit": limit
        }
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"获取文档分块时出错: {str(e)}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"获取文档分块失败: {str(e)}"
        )


# 健康检查端点
@router.get("/health", response_model=Dict[str, Any])
async def health_check() -> Dict[str, Any]:
    """健康检查"""
    return {
        "status": "healthy",
        "service": "document-processing-service",
        "timestamp": datetime.now().isoformat()
    }