import logging
import os
import uuid
from typing import Dict, List, Any, Optional, Tuple
from datetime import datetime

from models.document import Document, DocumentChunk, ChunkMetadata, DocumentStats
from models.processing_task import ProcessingTask, SubTask, TaskQueue, ProcessingLog
from models.db import get_db_context
from utils.file_handler import file_handler
from utils.document_parser import document_parser
from utils.text_chunker import text_chunker
from utils.embedding_generator import embedding_generator
from utils.redis_client import redis_client
from utils.milvus_client import milvus_client
from config.config import settings

logger = logging.getLogger(__name__)


class DocumentProcessingService:
    """文档处理服务类"""
    
    def __init__(self):
        """初始化文档处理服务"""
        self.max_retries = settings.max_retries
        self.document_storage_path = settings.document_storage_path
        
        # 确保文档存储目录存在
        os.makedirs(self.document_storage_path, exist_ok=True)
    
    async def process_document(self, file_path: str, file_name: str, user_id: str) -> Dict[str, Any]:
        """处理文档的主入口"""
        document_id = str(uuid.uuid4())
        
        try:
            # 1. 创建处理任务
            task_id = await self._create_processing_task(document_id, file_name, user_id)
            
            # 2. 保存原始文件
            saved_file_path = await self._save_document_file(file_path, file_name, document_id)
            
            # 3. 更新任务状态
            await self._update_task_status(task_id, "saved", {"file_path": saved_file_path})
            
            # 4. 解析文档
            parsing_result = await self._parse_document(saved_file_path, file_name)
            
            # 5. 更新任务状态
            await self._update_task_status(task_id, "parsed", parsing_result["metadata"])
            
            # 6. 创建文档记录
            document = await self._create_document_record(
                document_id, file_name, parsing_result, user_id
            )
            
            # 7. 文本分块
            chunks = await self._chunk_document(parsing_result, document_id)
            
            # 8. 更新任务状态
            await self._update_task_status(task_id, "chunked", {"chunk_count": len(chunks)})
            
            # 9. 生成嵌入向量
            chunks_with_embeddings = await self._generate_embeddings(chunks)
            
            # 10. 更新任务状态
            await self._update_task_status(task_id, "embedded", {"embedding_count": len(chunks_with_embeddings)})
            
            # 11. 保存文档分块和嵌入向量
            await self._save_document_chunks(chunks_with_embeddings, document_id)
            
            # 12. 更新文档统计信息
            await self._update_document_stats(document_id, chunks_with_embeddings)
            
            # 13. 完成任务
            await self._complete_task(task_id, document_id)
            
            # 14. 缓存文档状态
            redis_client.cache_document_status(document_id, "completed", {
                "chunk_count": len(chunks_with_embeddings),
                "file_name": file_name,
                "user_id": user_id
            })
            
            return {
                "success": True,
                "document_id": document_id,
                "task_id": task_id,
                "file_name": file_name,
                "status": "completed",
                "chunk_count": len(chunks_with_embeddings)
            }
            
        except Exception as e:
            logger.error(f"处理文档 {file_name} 时出错: {str(e)}")
            
            # 记录错误并更新任务状态
            if 'task_id' in locals():
                await self._fail_task(task_id, str(e))
            
            # 更新缓存中的状态
            redis_client.cache_document_status(document_id, "failed", {
                "error": str(e),
                "file_name": file_name
            })
            
            return {
                "success": False,
                "document_id": document_id,
                "error": str(e),
                "status": "failed"
            }
    
    async def _create_processing_task(self, document_id: str, file_name: str, user_id: str) -> str:
        """创建处理任务"""
        task_id = str(uuid.uuid4())
        
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
            db.commit()
            
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
            db.commit()
            
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
            
            logger.info(f"创建处理任务: {task_id} 用于文档: {document_id}")
            return task_id
    
    async def _save_document_file(self, file_path: str, file_name: str, document_id: str) -> str:
        """保存文档文件"""
        # 生成保存路径
        user_dir = os.path.join(self.document_storage_path, "uploads")
        saved_path = file_handler.save_file(file_path, file_name, user_dir)
        logger.info(f"文档已保存: {saved_path}")
        return saved_path
    
    async def _parse_document(self, file_path: str, file_name: str) -> Dict[str, Any]:
        """解析文档内容"""
        return document_parser.parse_document(file_path, file_name)
    
    async def _create_document_record(self, document_id: str, file_name: str, 
                                     parsing_result: Dict[str, Any], user_id: str) -> Document:
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
            return document
    
    async def _chunk_document(self, parsing_result: Dict[str, Any], document_id: str) -> List[Dict[str, Any]]:
        """将文档内容分块"""
        chunks = []
        
        for page in parsing_result["content"]:
            page_chunks = text_chunker.chunk_text(
                page["text"], 
                page.get("page_number", 1)
            )
            
            # 添加文档ID到每个分块
            for chunk in page_chunks:
                chunk["document_id"] = document_id
            
            chunks.extend(page_chunks)
        
        logger.info(f"文档分块完成，生成 {len(chunks)} 个分块")
        return chunks
    
    async def _generate_embeddings(self, chunks: List[Dict[str, Any]]) -> List[Dict[str, Any]]:
        """为文档分块生成嵌入向量，并进行向量验证和归一化"""
        # 生成嵌入向量
        chunks_with_embeddings = embedding_generator.batch_generate_embeddings(chunks)
        
        # 对成功生成的嵌入向量进行归一化处理
        normalized_chunks = []
        embeddings_to_normalize = []
        indices_to_update = []
        
        # 收集需要归一化的嵌入向量
        for i, chunk in enumerate(chunks_with_embeddings):
            if chunk.get("embedding_success", False) and chunk.get("embedding"):
                # 验证嵌入向量
                if not embedding_generator.validate_embedding(chunk["embedding"]):
                    logger.warning(f"嵌入向量验证失败，跳过归一化")
                    chunk["embedding_success"] = False
                    chunk["embedding_error"] = "嵌入向量验证失败"
                else:
                    embeddings_to_normalize.append(chunk["embedding"])
                    indices_to_update.append(i)
            normalized_chunks.append(chunk)
        
        # 批量归一化嵌入向量
        if embeddings_to_normalize:
            normalized_embeddings = embedding_generator.normalize_embeddings(embeddings_to_normalize)
            
            # 更新归一化后的嵌入向量
            for idx, original_idx in enumerate(indices_to_update):
                normalized_chunks[original_idx]["embedding"] = normalized_embeddings[idx]
                normalized_chunks[original_idx]["embedding_normalized"] = True
        
        logger.info(f"嵌入向量生成和归一化完成，成功 {sum(1 for chunk in normalized_chunks if chunk.get('embedding_success', False))} 个，失败 {sum(1 for chunk in normalized_chunks if not chunk.get('embedding_success', False))} 个")
        return normalized_chunks
    
    async def _save_document_chunks(self, chunks: List[Dict[str, Any]], document_id: str):
        """保存文档分块和嵌入向量到数据库和向量数据库"""
        # 准备保存到Milvus的向量数据
        milvus_data = []
        successful_chunks = 0
        
        with get_db_context() as db:
            for chunk in chunks:
                # 只有成功生成嵌入向量的分块才保存
                if chunk.get("embedding_success", False) and chunk.get("embedding"):
                    chunk_id = str(uuid.uuid4())
                    
                    # 创建分块元数据
                    metadata = ChunkMetadata(
                        page_number=chunk.get("page_number", 1),
                        start_pos=chunk.get("start_pos", 0),
                        end_pos=chunk.get("end_pos", 0),
                        word_count=chunk.get("word_count", 0),
                        token_count=chunk.get("token_count", 0)
                    )
                    db.add(metadata)
                    db.flush()  # 获取元数据ID
                    
                    # 创建文档分块
                    document_chunk = DocumentChunk(
                        id=chunk_id,
                        document_id=document_id,
                        content=chunk["content"],
                        embedding=chunk["embedding"],
                        embedding_model=chunk.get("embedding_model", ""),
                        metadata_id=metadata.id,
                        created_at=datetime.now()
                    )
                    db.add(document_chunk)
                    
                    # 准备Milvus数据
                    milvus_item = {
                        "chunk_id": chunk_id,
                        "document_id": document_id,
                        "content": chunk["content"],
                        "embedding": chunk["embedding"],
                        "created_at": int(datetime.now().timestamp() * 1000),  # 毫秒时间戳
                        "metadata": {
                            "page_number": chunk.get("page_number", 1),
                            "word_count": chunk.get("word_count", 0),
                            "token_count": chunk.get("token_count", 0),
                            "embedding_model": chunk.get("embedding_model", "")
                        }
                    }
                    milvus_data.append(milvus_item)
                    successful_chunks += 1
            
            db.commit()
            logger.info(f"已保存 {successful_chunks} 个文档分块到数据库")
        
        # 保存向量数据到Milvus
        if milvus_data:
            try:
                # 检查Milvus连接健康状态
                if not milvus_client.check_health():
                    logger.error("Milvus连接不健康，无法保存向量数据")
                    return
                
                # 批量插入向量数据
                success = milvus_client.insert_embeddings(milvus_data)
                if success:
                    logger.info(f"成功将 {len(milvus_data)} 个向量保存到Milvus")
                else:
                    logger.error(f"保存向量到Milvus失败")
            except Exception as e:
                logger.error(f"保存向量到Milvus时发生异常: {str(e)}")
                # 记录错误但不中断流程，确保文档分块已保存到关系数据库
    
    async def _update_document_stats(self, document_id: str, chunks: List[Dict[str, Any]]):
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
    
    async def _update_task_status(self, task_id: str, status: str, metadata: Optional[Dict[str, Any]] = None):
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
                db.commit()
                
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
    
    async def _complete_task(self, task_id: str, document_id: str):
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
    
    async def _fail_task(self, task_id: str, error_message: str):
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
            db.commit()
            
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
    
    async def get_task_status(self, task_id: str) -> Optional[Dict[str, Any]]:
        """获取任务状态"""
        # 先尝试从缓存获取
        cached_status = redis_client.get_processing_task(task_id)
        if cached_status:
            return cached_status
        
        # 从数据库获取
        with get_db_context() as db:
            task = db.query(ProcessingTask).filter(ProcessingTask.id == task_id).first()
            if not task:
                return None
            
            status = {
                "task_id": task.id,
                "document_id": task.document_id,
                "file_name": task.file_name,
                "status": task.status,
                "created_at": task.created_at.isoformat(),
                "updated_at": task.updated_at.isoformat(),
                "completed_at": task.completed_at.isoformat() if task.completed_at else None,
                "error_message": task.error_message
            }
            
            # 缓存状态
            redis_client.cache_processing_task(task_id, status)
            return status
    
    async def cancel_task(self, task_id: str) -> bool:
        """取消任务"""
        try:
            with get_db_context() as db:
                task = db.query(ProcessingTask).filter(ProcessingTask.id == task_id).first()
                if not task or task.status in ["completed", "failed", "canceled"]:
                    return False
                
                task.status = "canceled"
                task.updated_at = datetime.now()
                db.commit()
                
                # 更新队列状态
                queue_item = db.query(TaskQueue).filter(TaskQueue.task_id == task_id).first()
                if queue_item:
                    queue_item.status = "canceled"
                    db.commit()
                
                logger.info(f"任务已取消: {task_id}")
                return True
        except Exception as e:
            logger.error(f"取消任务 {task_id} 时出错: {str(e)}")
            return False


    async def delete_document(self, document_id: str, user_id: str) -> Dict[str, Any]:
        """删除文档及其所有相关数据
        
        Args:
            document_id: 文档ID
            user_id: 用户ID（用于权限验证）
            
        Returns:
            Dict: 删除结果
        """
        try:
            # 验证文档所有权
            with get_db_context() as db:
                document = db.query(Document).filter(
                    Document.id == document_id,
                    Document.user_id == user_id
                ).first()
                
                if not document:
                    return {
                        "success": False,
                        "error": "文档不存在或无权限删除",
                        "code": "document_not_found"
                    }
                
                # 先从Milvus删除向量数据
                try:
                    if milvus_client.check_health():
                        success = milvus_client.delete_by_document_id(document_id)
                        if success:
                            logger.info(f"从Milvus成功删除文档 {document_id} 的向量数据")
                        else:
                            logger.warning(f"从Milvus删除向量数据失败，但继续删除数据库记录")
                    else:
                        logger.warning("Milvus连接不健康，跳过向量数据删除")
                except Exception as e:
                    logger.error(f"从Milvus删除向量数据时发生异常: {str(e)}")
                    # 记录错误但继续删除数据库记录
                
                # 删除相关的文档分块
                chunks = db.query(DocumentChunk).filter(DocumentChunk.document_id == document_id).all()
                chunk_count = len(chunks)
                
                # 删除分块元数据
                metadata_ids = [chunk.metadata_id for chunk in chunks if chunk.metadata_id]
                if metadata_ids:
                    db.query(ChunkMetadata).filter(ChunkMetadata.id.in_(metadata_ids)).delete()
                
                # 删除文档分块
                db.query(DocumentChunk).filter(DocumentChunk.document_id == document_id).delete()
                
                # 删除文档统计信息
                db.query(DocumentStats).filter(DocumentStats.document_id == document_id).delete()
                
                # 删除相关任务
                tasks = db.query(ProcessingTask).filter(ProcessingTask.document_id == document_id).all()
                task_ids = [task.id for task in tasks]
                
                if task_ids:
                    # 删除子任务
                    db.query(SubTask).filter(SubTask.task_id.in_(task_ids)).delete()
                    # 删除处理日志
                    db.query(ProcessingLog).filter(ProcessingLog.task_id.in_(task_ids)).delete()
                    # 删除队列项
                    db.query(TaskQueue).filter(TaskQueue.task_id.in_(task_ids)).delete()
                    # 删除任务
                    db.query(ProcessingTask).filter(ProcessingTask.document_id == document_id).delete()
                
                # 删除文档记录
                db.delete(document)
                db.commit()
                
                # 清理缓存
                redis_client.delete_document_cache(document_id)
                
                logger.info(f"成功删除文档 {document_id} 及其 {chunk_count} 个分块")
                return {
                    "success": True,
                    "message": "文档删除成功",
                    "document_id": document_id,
                    "chunk_count": chunk_count
                }
                
        except Exception as e:
            logger.error(f"删除文档 {document_id} 时发生异常: {str(e)}")
            return {
                "success": False,
                "error": f"删除文档失败: {str(e)}",
                "code": "delete_failed"
            }

    async def search_similar_chunks(self, query_text: str, user_id: str, top_k: int = 10) -> Dict[str, Any]:
        """搜索相似文档分块
        
        Args:
            query_text: 查询文本
            user_id: 用户ID
            top_k: 返回结果数量
            
        Returns:
            Dict: 搜索结果
        """
        try:
            # 生成查询嵌入向量
            embedding_results = embedding_generator.generate_embeddings([query_text])
            if not embedding_results or not embedding_results[0].get("embedding"):
                return {
                    "success": False,
                    "error": "无法生成查询嵌入向量",
                    "code": "embedding_failed"
                }
            
            query_embedding = embedding_results[0]["embedding"]
            
            # 验证嵌入向量
            if not embedding_generator.validate_embedding(query_embedding):
                return {
                    "success": False,
                    "error": "查询嵌入向量无效",
                    "code": "invalid_embedding"
                }
            
            # 归一化查询嵌入向量
            normalized_query = embedding_generator.normalize_embeddings([query_embedding])[0]
            
            # 构建过滤表达式，只搜索用户自己的文档
            with get_db_context() as db:
                # 获取用户的所有文档ID
                user_docs = db.query(Document.id).filter(Document.user_id == user_id).all()
                if not user_docs:
                    return {
                        "success": True,
                        "results": [],
                        "message": "没有找到相关文档"
                    }
                
                doc_ids = [doc.id for doc in user_docs]
                # 构建过滤表达式
                filter_expr = f"document_id in {tuple(doc_ids)}"
                
                # 从Milvus搜索相似向量
                if milvus_client.check_health():
                    search_results = milvus_client.search_similar(
                        normalized_query,
                        top_k=top_k,
                        filter_expr=filter_expr
                    )
                    
                    # 丰富结果信息，从数据库获取文档标题等信息
                    for result in search_results:
                        doc = db.query(Document).filter(Document.id == result["document_id"]).first()
                        if doc:
                            result["document_title"] = doc.file_name
                            result["document_status"] = doc.status
                else:
                    logger.error("Milvus连接不健康，无法执行相似搜索")
                    return {
                        "success": False,
                        "error": "向量搜索服务不可用",
                        "code": "search_service_unavailable"
                    }
            
            return {
                "success": True,
                "results": search_results,
                "query_text": query_text,
                "result_count": len(search_results)
            }
            
        except Exception as e:
            logger.error(f"执行相似搜索时发生异常: {str(e)}")
            return {
                "success": False,
                "error": f"搜索失败: {str(e)}",
                "code": "search_failed"
            }


# 导出实例
document_processing_service = DocumentProcessingService()