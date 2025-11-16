import logging
import numpy as np
from typing import List, Dict, Any, Optional
import requests
import json
import time
from datetime import datetime

from config.config import settings

logger = logging.getLogger(__name__)


class EmbeddingGenerator:
    """嵌入向量生成工具类"""
    
    def __init__(self):
        """初始化嵌入生成器"""
        self.embedding_model = settings.embedding_model
        self.embedding_api_url = settings.embedding_api_url
        self.embedding_api_key = settings.embedding_api_key
        self.embedding_dimension = settings.embedding_dimension
        self.max_batch_size = settings.embedding_max_batch_size
        self.use_local_model = settings.use_local_embedding_model
        
        # 如果使用本地模型，加载模型
        if self.use_local_model:
            try:
                self._load_local_model()
                logger.info("成功加载本地嵌入模型")
            except Exception as e:
                logger.error(f"加载本地嵌入模型失败: {str(e)}")
                raise
    
    def _load_local_model(self):
        """加载本地嵌入模型"""
        try:
            from sentence_transformers import SentenceTransformer
            self.model = SentenceTransformer(self.embedding_model)
            logger.info(f"本地模型 {self.embedding_model} 加载成功")
        except ImportError:
            logger.error("sentence-transformers 库未安装")
            raise ImportError("请安装 sentence-transformers: pip install sentence-transformers")
        except Exception as e:
            logger.error(f"加载本地模型时出错: {str(e)}")
            raise
    
    def generate_embeddings(self, texts: List[str]) -> List[Dict[str, Any]]:
        """为文本列表生成嵌入向量"""
        if not texts:
            return []
        
        results = []
        
        # 分批处理文本
        for i in range(0, len(texts), self.max_batch_size):
            batch = texts[i:i + self.max_batch_size]
            try:
                if self.use_local_model:
                    batch_embeddings = self._generate_local_embeddings(batch)
                else:
                    batch_embeddings = self._generate_api_embeddings(batch)
                
                results.extend(batch_embeddings)
                
                # 添加速率限制延迟
                if len(texts) > self.max_batch_size:
                    time.sleep(settings.embedding_rate_limit_delay)
                    
            except Exception as e:
                logger.error(f"生成批次嵌入向量失败 (批次 {i//self.max_batch_size + 1}): {str(e)}")
                # 对每个失败的文本生成空结果
                for text in batch:
                    results.append({
                        "text": text,
                        "embedding": None,
                        "error": str(e),
                        "success": False,
                        "timestamp": datetime.now().isoformat()
                    })
        
        return results
    
    def _generate_local_embeddings(self, texts: List[str]) -> List[Dict[str, Any]]:
        """使用本地模型生成嵌入向量"""
        results = []
        
        try:
            # 生成嵌入向量
            embeddings = self.model.encode(texts, convert_to_numpy=True)
            
            # 处理结果
            for i, (text, embedding) in enumerate(zip(texts, embeddings)):
                results.append({
                    "text": text,
                    "embedding": embedding.tolist(),  # 转换为列表格式
                    "model": self.embedding_model,
                    "dimension": len(embedding),
                    "success": True,
                    "timestamp": datetime.now().isoformat()
                })
                
        except Exception as e:
            logger.error(f"使用本地模型生成嵌入时出错: {str(e)}")
            raise
        
        return results
    
    def _generate_api_embeddings(self, texts: List[str]) -> List[Dict[str, Any]]:
        """使用API生成嵌入向量"""
        headers = {
            "Content-Type": "application/json"
        }
        
        # 如果有API密钥，添加到头部
        if self.embedding_api_key:
            headers["Authorization"] = f"Bearer {self.embedding_api_key}"
        
        # 准备请求数据
        payload = {
            "model": self.embedding_model,
            "input": texts
        }
        
        try:
            # 发送请求
            response = requests.post(
                self.embedding_api_url,
                headers=headers,
                data=json.dumps(payload, ensure_ascii=False)
            )
            
            # 检查响应状态
            response.raise_for_status()
            
            # 解析响应
            data = response.json()
            
            # 处理结果
            results = []
            for i, text in enumerate(texts):
                embedding = data.get("data", [{}])[i].get("embedding")
                if embedding:
                    results.append({
                        "text": text,
                        "embedding": embedding,
                        "model": self.embedding_model,
                        "dimension": len(embedding),
                        "success": True,
                        "timestamp": datetime.now().isoformat()
                    })
                else:
                    results.append({
                        "text": text,
                        "embedding": None,
                        "error": "API未返回嵌入向量",
                        "success": False,
                        "timestamp": datetime.now().isoformat()
                    })
            
            return results
            
        except requests.RequestException as e:
            logger.error(f"API请求失败: {str(e)}")
            raise
        except (KeyError, IndexError) as e:
            logger.error(f"API响应解析失败: {str(e)}")
            raise ValueError(f"API响应格式不正确: {str(e)}")
    
    def validate_embedding(self, embedding: List[float]) -> bool:
        """验证嵌入向量是否有效"""
        if not embedding:
            return False
        
        if not isinstance(embedding, list):
            return False
        
        # 检查嵌入向量维度
        if self.embedding_dimension > 0 and len(embedding) != self.embedding_dimension:
            logger.warning(f"嵌入向量维度不匹配: 期望 {self.embedding_dimension}, 实际 {len(embedding)}")
            return False
        
        # 检查向量中是否有有效数值
        try:
            embedding_array = np.array(embedding)
            if np.isnan(embedding_array).any() or np.isinf(embedding_array).any():
                return False
            return True
        except Exception:
            return False
    
    def normalize_embeddings(self, embeddings: List[List[float]]) -> List[List[float]]:
        """归一化嵌入向量"""
        normalized_embeddings = []
        
        for embedding in embeddings:
            if not embedding:
                normalized_embeddings.append([])
                continue
            
            try:
                embedding_array = np.array(embedding)
                norm = np.linalg.norm(embedding_array)
                
                if norm > 0:
                    normalized = (embedding_array / norm).tolist()
                    normalized_embeddings.append(normalized)
                else:
                    normalized_embeddings.append(embedding)
            except Exception:
                normalized_embeddings.append(embedding)
        
        return normalized_embeddings
    
    def batch_generate_embeddings(self, chunks: List[Dict[str, Any]]) -> List[Dict[str, Any]]:
        """为文档分块批量生成嵌入向量"""
        if not chunks:
            return []
        
        # 提取文本
        texts = [chunk.get("content", "") for chunk in chunks]
        
        # 生成嵌入向量
        embedding_results = self.generate_embeddings(texts)
        
        # 合并结果
        enhanced_chunks = []
        for i, chunk in enumerate(chunks):
            if i < len(embedding_results):
                embedding_result = embedding_results[i]
                chunk["embedding"] = embedding_result.get("embedding")
                chunk["embedding_model"] = embedding_result.get("model")
                chunk["embedding_dimension"] = embedding_result.get("dimension")
                chunk["embedding_success"] = embedding_result.get("success", False)
                chunk["embedding_timestamp"] = embedding_result.get("timestamp")
            
            enhanced_chunks.append(chunk)
        
        return enhanced_chunks


# 导出实例
embedding_generator = EmbeddingGenerator()