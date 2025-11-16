import logging
from typing import List, Dict, Any, Optional, Tuple
from pymilvus import connections, Collection, FieldSchema, CollectionSchema, DataType, utility
from config.config import settings

logger = logging.getLogger(__name__)

class MilvusClient:
    """Milvus向量数据库客户端工具类"""
    
    def __init__(self):
        """初始化Milvus连接"""
        self.uri = settings.milvus_uri
        self.token = settings.milvus_token
        self.collection_name = settings.milvus_collection_name
        self.embedding_dim = settings.embedding_dimension
        self.connection = None
        self.collection = None
        self._connect()
    
    def _connect(self):
        """建立Milvus连接"""
        try:
            # 连接到Milvus服务
            connections.connect(
                uri=self.uri,
                token=self.token,
                secure=settings.milvus_tls_enabled,
                timeout=settings.milvus_timeout,
                max_retries=settings.milvus_max_retries
            )
            logger.info(f"成功连接到Milvus服务: {self.uri}")
            
            # 检查并创建集合
            self._ensure_collection_exists()
            
            # 加载集合到内存
            self._load_collection()
        except Exception as e:
            logger.error(f"初始化Milvus客户端失败: {str(e)}")
            raise
    
    def _ensure_collection_exists(self):
        """确保集合存在，如果不存在则创建"""
        try:
            if not utility.has_collection(self.collection_name):
                logger.info(f"Milvus集合 {self.collection_name} 不存在，正在创建...")
                
                # 定义字段
                fields = [
                    FieldSchema(name="id", dtype=DataType.INT64, is_primary=True, auto_id=True),
                    FieldSchema(name="chunk_id", dtype=DataType.VARCHAR, max_length=255),
                    FieldSchema(name="document_id", dtype=DataType.VARCHAR, max_length=255),
                    FieldSchema(name="content", dtype=DataType.VARCHAR, max_length=65535),
                    FieldSchema(name="embedding", dtype=DataType.FLOAT_VECTOR, dim=self.embedding_dim),
                    FieldSchema(name="created_at", dtype=DataType.INT64),
                    FieldSchema(name="metadata", dtype=DataType.JSON)
                ]
                
                # 创建集合
                schema = CollectionSchema(fields, description="文档块向量集合")
                Collection(
                    name=self.collection_name,
                    schema=schema,
                    consistency_level="Strong"
                )
                
                # 创建索引
                self._create_index()
                logger.info(f"Milvus集合 {self.collection_name} 创建成功")
            else:
                logger.info(f"Milvus集合 {self.collection_name} 已存在")
        except Exception as e:
            logger.error(f"确保Milvus集合存在失败: {str(e)}")
            raise
    
    def _create_index(self):
        """为向量字段创建索引"""
        try:
            # 获取集合
            collection = Collection(self.collection_name)
            
            # 创建索引
            index_params = {
                "index_type": settings.milvus_index_type,
                "metric_type": settings.milvus_metric_type,
                "params": settings.milvus_index_params
            }
            collection.create_index(
                field_name="embedding",
                index_params=index_params,
                index_name="embedding_index"
            )
            logger.info(f"Milvus集合索引创建成功")
        except Exception as e:
            logger.error(f"创建Milvus索引失败: {str(e)}")
            raise
    
    def _load_collection(self):
        """加载集合到内存"""
        try:
            self.collection = Collection(self.collection_name)
            self.collection.load()
            logger.info(f"Milvus集合 {self.collection_name} 已加载到内存")
        except Exception as e:
            logger.error(f"加载Milvus集合失败: {str(e)}")
            raise
    
    def insert_embeddings(self, data: List[Dict[str, Any]]) -> bool:
        """插入向量数据
        
        Args:
            data: 要插入的数据列表，每个元素包含chunk_id, document_id, content, embedding, created_at, metadata
            
        Returns:
            bool: 是否插入成功
        """
        try:
            if not self.collection:
                self._load_collection()
            
            # 准备插入数据
            entities = [
                [d["chunk_id"] for d in data],  # chunk_id
                [d["document_id"] for d in data],  # document_id
                [d["content"] for d in data],  # content
                [d["embedding"] for d in data],  # embedding
                [d["created_at"] for d in data],  # created_at
                [d["metadata"] for d in data]  # metadata
            ]
            
            # 插入数据
            result = self.collection.insert(entities)
            self.collection.flush()
            logger.info(f"成功插入 {len(data)} 条向量数据，插入ID: {result.primary_keys[:3]}...")
            return True
        except Exception as e:
            logger.error(f"插入向量数据失败: {str(e)}")
            return False
    
    def search_similar(self, query_embedding: List[float], top_k: int = 5, 
                      filter_expr: Optional[str] = None) -> List[Dict[str, Any]]:
        """搜索相似向量
        
        Args:
            query_embedding: 查询向量
            top_k: 返回结果数量
            filter_expr: 过滤表达式
            
        Returns:
            List[Dict]: 搜索结果列表，包含相似度、chunk_id、document_id、content等信息
        """
        try:
            if not self.collection:
                self._load_collection()
            
            # 准备搜索参数
            search_params = {
                "metric_type": settings.milvus_metric_type,
                "params": settings.milvus_search_params
            }
            
            # 执行搜索
            results = self.collection.search(
                data=[query_embedding],
                anns_field="embedding",
                param=search_params,
                limit=top_k,
                expr=filter_expr,
                output_fields=["chunk_id", "document_id", "content", "metadata"]
            )
            
            # 处理搜索结果
            search_results = []
            for hits in results:
                for hit in hits:
                    search_results.append({
                        "similarity": hit.distance,
                        "chunk_id": hit.entity.get("chunk_id"),
                        "document_id": hit.entity.get("document_id"),
                        "content": hit.entity.get("content"),
                        "metadata": hit.entity.get("metadata"),
                        "id": hit.id
                    })
            
            logger.info(f"相似向量搜索完成，返回 {len(search_results)} 条结果")
            return search_results
        except Exception as e:
            logger.error(f"搜索相似向量失败: {str(e)}")
            return []
    
    def delete_by_document_id(self, document_id: str) -> bool:
        """根据文档ID删除相关向量
        
        Args:
            document_id: 文档ID
            
        Returns:
            bool: 是否删除成功
        """
        try:
            if not self.collection:
                self._load_collection()
            
            # 执行删除
            expr = f"document_id == '{document_id}'"
            result = self.collection.delete(expr)
            self.collection.flush()
            logger.info(f"已删除文档 {document_id} 的 {result.delete_count} 条向量数据")
            return True
        except Exception as e:
            logger.error(f"删除文档向量数据失败: {str(e)}")
            return False
    
    def get_collection_stats(self) -> Dict[str, Any]:
        """获取集合统计信息
        
        Returns:
            Dict: 集合统计信息
        """
        try:
            stats = utility.get_collection_stats(self.collection_name)
            logger.info(f"获取Milvus集合统计信息: {stats}")
            return stats
        except Exception as e:
            logger.error(f"获取集合统计信息失败: {str(e)}")
            return {"error": str(e)}
    
    def check_health(self) -> bool:
        """检查Milvus连接健康状态
        
        Returns:
            bool: 连接是否健康
        """
        try:
            # 检查连接
            status = connections.has_connection(alias="default")
            if not status:
                logger.warning("Milvus连接不存在，尝试重新连接...")
                self._connect()
                return True
            
            # 检查集合
            if not utility.has_collection(self.collection_name):
                logger.warning(f"Milvus集合 {self.collection_name} 不存在")
                return False
            
            # 检查集合是否加载
            if self.collection:
                loaded = self.collection.is_loaded
                if not loaded:
                    logger.info(f"Milvus集合 {self.collection_name} 未加载，尝试加载...")
                    self._load_collection()
            
            logger.info("Milvus连接健康检查通过")
            return True
        except Exception as e:
            logger.error(f"Milvus健康检查失败: {str(e)}")
            return False
    
    def close(self):
        """关闭Milvus连接"""
        try:
            if self.collection:
                self.collection.release()
                logger.info("Milvus集合已从内存中释放")
                self.collection = None
            
            connections.disconnect(alias="default")
            logger.info("Milvus连接已关闭")
        except Exception as e:
            logger.error(f"关闭Milvus连接时出错: {str(e)}")

# 导出全局实例
milvus_client = MilvusClient()