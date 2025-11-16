import logging
import os
import uuid
from typing import Dict, List, Any, Optional
from datetime import datetime

from celery import shared_task
from celery.exceptions import SoftTimeLimitExceeded
from celery.utils.log import get_task_logger

from config.config import settings
from models.db import get_db_context
from models.document import Document, DocumentChunk, ChunkMetadata, DocumentStats
from models.processing_task import ProcessingTask, SubTask, TaskQueue, ProcessingLog
from utils.file_handler import file_handler
from utils.document_parser import document_parser
from utils.text_chunker import text_chunker
from utils.embedding_generator import embedding_generator
from utils.redis_client import redis_client

# 获取任务日志记录器
logger = get_task_logger(__name__)


@shared_task(bind=True, autoretry_for=(Exception,), retry_backoff=True, retry_kwargs={'max_retries': 3})
def process_document_task(self, file_path: str, file_name: str, user_id: str) -> Dict[str, Any]:
    """处理文档的Celery任务"""
    document_id = str(uuid.uuid4())
    task_id = self.request.id
    
    try:
        logger.info(f"开始处理文档: {file_name}, 任务ID: {task_id}")
        
        # 创建任务记录
        create_task_record(task_id, document_id, file_name, user_id)
        
        # 更新任务状态
        update_task_status(task_id, "processing")
        
        # 保存文档文件
        saved_file_path = save_document_file(file_path, file_name, document_id)
        update_task_status(task_id, "file_saved", {"file_path": saved_file_path})
        
        # 解析文档
        parsing_result = parse_document(saved_file_path, file_name)
        update_task_status(task_id, "parsed", parsing_result["metadata"])
        
        # 创建文档记录
        create_document_record(document_id, file_name, parsing_result, user_id)
        
        # 文本分块
        chunks = chunk_document(parsing_result, document_id)
        update_task_status(task_id, "chunked", {"chunk_count": len(chunks)})
        
        # 生成嵌入向量（异步任务）
        embedding_task = generate_embeddings.apply_async(
            args=[chunks],
            link=save_chunks_task.s(document_id, task_id)
        )
        
        logger.info(f"文档处理任务已提交，嵌入向量生成任务ID: {embedding_task.id}")
        
        # 返回任务状态
        return {
            "success": True,
            "document_id": document_id,
            "task_id": task_id,
            "file_name": file_name,
            "status": "processing",
            "embedding_task_id": embedding_task.id
        }
        
    except SoftTimeLimitExceeded:
        error_msg = "任务处理超时"
        logger.error(f"{error_msg}: {task_id}")
        fail_task(task_id, error_msg)
        redis_client.cache_document_status(document_id, "failed", {"error": error_msg})
        return {
            "success": False,
            "document_id": document_id,
            "task_id": task_id,
            "error": error_msg
        }
    except Exception as e:
        error_msg = str(e)
        logger.error(f"处理文档任务失败: {error_msg}")
        fail_task(task_id, error_msg)
        redis_client.cache_document_status(document_id, "failed", {"error": error_msg})
        return {
            "success": False,
            "document_id": document_id,
            "task_id": task_id,
            "error": error_msg
        }


@shared_task(bind=True)
def parse_document(self, file_path: str, file_name: str) -> Dict[str, Any]:
    """解析文档的任务"""
    try:
        logger.info(f"开始解析文档: {file_name}")
        result = document_parser.parse_document(file_path, file_name)
        logger.info(f"文档解析完成: {file_name}, 页数: {result['metadata'].get('total_pages', 0)}")
        return result
    except Exception as e:
        logger.error(f"解析文档失败: {str(e)}")
        raise


@shared_task(bind=True)
def generate_embeddings(self, chunks: List[Dict[str, Any]]) -> List[Dict[str, Any]]:
    """生成嵌入向量的任务"""
    try:
        logger.info(f"开始生成嵌入向量，分块数: {len(chunks)}")
        result = embedding_generator.batch_generate_embeddings(chunks)
        success_count = sum(1 for chunk in result if chunk.get('embedding_success', False))
        logger.info(f"嵌入向量生成完成，成功: {success_count}/{len(result)}")
        return result
    except Exception as e:
        logger.error(f"生成嵌入向量失败: {str(e)}")
        raise


@shared_task(bind=True)
def save_chunks_task(self, chunks: List[Dict[str, Any]], document_id: str, parent_task_id: str) -> Dict[str, Any]:
    """保存文档分块的任务"""
    try:
        logger.info(f"开始保存文档分块，文档ID: {document_id}")
        
        # 保存分块
        save_document_chunks(chunks, document_id)
        
        # 更新文档统计
        update_document_stats(document_id, chunks)
        
        # 完成父任务
        complete_task(parent_task_id, document_id)
        
        # 更新缓存
        redis_client.cache_document_status(document_id, "completed", {
            "chunk_count": len(chunks),
            "document_id": document_id
        })
        
        logger.info(f"文档处理完成，文档ID: {document_id}")
        
        return {
            "success": True,
            "document_id": document_id,
            "chunk_count": len(chunks),
            "status": "completed"
        }
        
    except Exception as e:
        error_msg = str(e)
        logger.error(f"保存文档分块失败: {error_msg}")
        fail_task(parent_task_id, error_msg)
        redis_client.cache_document_status(document_id, "failed", {"error": error_msg})
        raise


@shared_task(bind=True)
def batch_process_documents(self, documents: List[Dict[str, Any]]) -> List[Dict[str, Any]]:
    """批量处理文档的任务"""
    results = []
    
    for doc in documents:
        try:
            # 提交单个文档处理任务
            task_result = process_document_task.apply_async(
                args=[doc['file_path'], doc['file_name'], doc['user_id']]
            )
            results.append({
                "file_name": doc['file_name'],
                "task_id": task_result.id,
                "status": "submitted"
            })
        except Exception as e:
            results.append({
                "file_name": doc['file_name'],
                "status": "failed",
                "error": str(e)
            })
    
    return results


@shared_task(bind=True)
def cleanup_old_tasks(self, days: int = 7) -> Dict[str, int]:
    """清理旧任务记录的任务"""
    try:
        from datetime import timedelta
        cutoff_date = datetime.now() - timedelta(days=days)
        
        with get_db_context() as db:
            # 删除旧的处理日志
            log_count = db.query(ProcessingLog).filter(
                ProcessingLog.created_at < cutoff_date
            ).delete()
            
            # 删除旧的子任务
            subtask_count = db.query(SubTask).filter(
                SubTask.created_at < cutoff_date
            ).delete()
            
            # 删除旧的任务队列记录
            queue_count = db.query(TaskQueue).filter(
                TaskQueue.created_at < cutoff_date
            ).delete()
            
            # 删除旧的任务记录（仅删除已完成或失败的任务）
            task_count = db.query(ProcessingTask).filter(
                ProcessingTask.created_at < cutoff_date,
                ProcessingTask.status.in_(["completed", "failed", "canceled"])
            ).delete()
            
            db.commit()
            
            logger.info(f"清理旧任务完成: 日志 {log_count}, 子任务 {subtask_count}, 队列 {queue_count}, 任务 {task_count}")
            
            return {
                "logs_deleted": log_count,
                "subtasks_deleted": subtask_count,
                "queue_items_deleted": queue_count,
                "tasks_deleted": task_count
            }
            
    except Exception as e:
        logger.error(f"清理旧任务失败: {str(e)}")
        raise


# 辅助函数
def create_task_record(task_id: str, document_id: str, file_name: str, user_id: str):
    """创建任务记录"""
    with get_db_context() as db:
        task = ProcessingTask(
            id=task_id,
            document_id=document_id,
            file_name=file_name,
            user_id=user_id,
            status="created",
            priority=0,
            created_at=datetime.now(),
            updated_at=datetime.now()
        )
        db.add(task)
        
        # 创建初始子任务
        subtask = SubTask(
            id=str(uuid.uuid4()),
            task_id=task_id,
            name="初始化任务",
            status="completed",
            created_at=datetime.now(),
            completed_at=datetime.now()
        )
        db.add(subtask)
        
        # 添加到任务队列
        queue_item = TaskQueue(
            id=str(uuid.uuid4()),
            task_id=task_id,
            priority=0,
            status="queued",
            enqueued_at=datetime.now()
        )
        db.add(queue_item)
        
        db.commit()


def save_document_file(file_path: str, file_name: str, document_id: str) -> str:
    """保存文档文件"""
    user_dir = os.path.join(settings.document_storage_path, "uploads")
    return file_handler.save_file(file_path, file_name, user_dir)


def create_document_record(document_id: str, file_name: str, 
                          parsing_result: Dict[str, Any], user_id: str):
    """创建文档记录"""
    with get_db_context() as db:
        document = Document(
            id=document_id,
            file_name=file_name,
            user_id=user_id,
            status="processing",
            total_pages=parsing_result["metadata"].get("total_pages", 0),
            metadata=parsing_result["metadata"],
            created_at=datetime.now(),
            updated_at=datetime.now()
        )
        db.add(document)
        db.commit()


def chunk_document(parsing_result: Dict[str, Any], document_id: str) -> List[Dict[str, Any]]:
    """将文档内容分块"""
    chunks = []
    
    for page in parsing_result["content"]:
        page_chunks = text_chunker.chunk_text(
            page["text"], 
            page.get("page_number", 1)
        )
        
        for chunk in page_chunks:
            chunk["document_id"] = document_id
        
        chunks.extend(page_chunks)
    
    return chunks


def save_document_chunks(chunks: List[Dict[str, Any]], document_id: str):
    """保存文档分块"""
    with get_db_context() as db:
        for chunk in chunks:
            if chunk.get("embedding_success", False) and chunk.get("embedding"):
                # 创建分块元数据
                metadata = ChunkMetadata(
                    page_number=chunk.get("page_number", 1),
                    start_pos=chunk.get("start_pos", 0),
                    end_pos=chunk.get("end_pos", 0),
                    word_count=chunk.get("word_count", 0),
                    token_count=chunk.get("token_count", 0)
                )
                db.add(metadata)
                db.flush()
                
                # 创建文档分块
                document_chunk = DocumentChunk(
                    id=str(uuid.uuid4()),
                    document_id=document_id,
                    content=chunk["content"],
                    embedding=chunk["embedding"],
                    embedding_model=chunk.get("embedding_model", ""),
                    metadata_id=metadata.id,
                    created_at=datetime.now()
                )
                db.add(document_chunk)
        
        db.commit()


def update_document_stats(document_id: str, chunks: List[Dict[str, Any]]):
    """更新文档统计信息"""
    total_chunks = len(chunks)
    total_tokens = sum(chunk.get("token_count", 0) for chunk in chunks)
    total_words = sum(chunk.get("word_count", 0) for chunk in chunks)
    
    with get_db_context() as db:
        stats = DocumentStats(
            document_id=document_id,
            total_chunks=total_chunks,
            total_tokens=total_tokens,
            total_words=total_words,
            updated_at=datetime.now()
        )
        db.add(stats)
        db.commit()


def update_task_status(task_id: str, status: str, metadata: Optional[Dict[str, Any]] = None):
    """更新任务状态"""
    with get_db_context() as db:
        task = db.query(ProcessingTask).filter(ProcessingTask.id == task_id).first()
        if task:
            task.status = status
            task.updated_at = datetime.now()
            db.commit()
            
            # 创建子任务记录
            subtask = SubTask(
                id=str(uuid.uuid4()),
                task_id=task_id,
                name=status,
                status="completed",
                metadata=metadata or {},
                created_at=datetime.now(),
                completed_at=datetime.now()
            )
            db.add(subtask)
            
            # 记录处理日志
            log = ProcessingLog(
                id=str(uuid.uuid4()),
                task_id=task_id,
                message=f"Task status changed to {status}",
                level="info",
                created_at=datetime.now()
            )
            db.add(log)
            db.commit()


def complete_task(task_id: str, document_id: str):
    """完成任务"""
    with get_db_context() as db:
        # 更新任务状态
        task = db.query(ProcessingTask).filter(ProcessingTask.id == task_id).first()
        if task:
            task.status = "completed"
            task.completed_at = datetime.now()
            task.updated_at = datetime.now()
            db.commit()
        
        # 更新文档状态
        document = db.query(Document).filter(Document.id == document_id).first()
        if document:
            document.status = "completed"
            document.updated_at = datetime.now()
            db.commit()
        
        # 更新队列状态
        queue_item = db.query(TaskQueue).filter(TaskQueue.task_id == task_id).first()
        if queue_item:
            queue_item.status = "completed"
            queue_item.processed_at = datetime.now()
            db.commit()


def fail_task(task_id: str, error_message: str):
    """标记任务失败"""
    with get_db_context() as db:
        # 更新任务状态
        task = db.query(ProcessingTask).filter(ProcessingTask.id == task_id).first()
        if task:
            task.status = "failed"
            task.error_message = error_message
            task.updated_at = datetime.now()
            db.commit()
        
        # 创建失败子任务
        subtask = SubTask(
            id=str(uuid.uuid4()),
            task_id=task_id,
            name="处理失败",
            status="failed",
            error_message=error_message,
            created_at=datetime.now(),
            completed_at=datetime.now()
        )
        db.add(subtask)
        
        # 记录错误日志
        log = ProcessingLog(
            id=str(uuid.uuid4()),
            task_id=task_id,
            message=error_message,
            level="error",
            created_at=datetime.now()
        )
        db.add(log)
        db.commit()