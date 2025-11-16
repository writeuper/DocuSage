import logging
import os
from typing import Optional, List, Dict, Any, BinaryIO
from minio import Minio
from minio.error import S3Error
from datetime import timedelta
from config.config import settings

logger = logging.getLogger(__name__)

class MinioClient:
    """MinIO对象存储客户端工具类"""
    
    def __init__(self):
        """初始化MinIO连接"""
        self.endpoint = settings.minio_endpoint
        self.access_key = settings.minio_access_key
        self.secret_key = settings.minio_secret_key
        self.secure = settings.minio_secure
        self.bucket_name = settings.minio_bucket_name
        self.client = None
        self._connect()
    
    def _connect(self):
        """建立MinIO连接"""
        try:
            # 创建MinIO客户端
            self.client = Minio(
                endpoint=self.endpoint,
                access_key=self.access_key,
                secret_key=self.secret_key,
                secure=self.secure,
                region=settings.minio_region,
                http_client=settings.minio_http_client if hasattr(settings, 'minio_http_client') else None
            )
            
            logger.info(f"成功连接到MinIO服务: {self.endpoint}")
            
            # 检查并创建存储桶
            self._ensure_bucket_exists()
        except Exception as e:
            logger.error(f"初始化MinIO客户端失败: {str(e)}")
            raise
    
    def _ensure_bucket_exists(self):
        """确保存储桶存在，如果不存在则创建"""
        try:
            if not self.client.bucket_exists(self.bucket_name):
                logger.info(f"MinIO存储桶 {self.bucket_name} 不存在，正在创建...")
                self.client.make_bucket(
                    bucket_name=self.bucket_name,
                    location=settings.minio_region
                )
                logger.info(f"MinIO存储桶 {self.bucket_name} 创建成功")
                
                # 设置存储桶策略（如果配置了）
                if settings.minio_bucket_policy:
                    self._set_bucket_policy()
            else:
                logger.info(f"MinIO存储桶 {self.bucket_name} 已存在")
        except Exception as e:
            logger.error(f"确保MinIO存储桶存在失败: {str(e)}")
            raise
    
    def _set_bucket_policy(self):
        """设置存储桶策略"""
        try:
            policy = settings.minio_bucket_policy
            self.client.set_bucket_policy(self.bucket_name, policy)
            logger.info(f"MinIO存储桶 {self.bucket_name} 策略设置成功")
        except Exception as e:
            logger.error(f"设置MinIO存储桶策略失败: {str(e)}")
    
    def upload_file(self, file_path: str, object_name: Optional[str] = None, 
                   content_type: Optional[str] = None) -> bool:
        """上传文件到MinIO
        
        Args:
            file_path: 本地文件路径
            object_name: 对象名称，如果不提供则使用文件名
            content_type: 内容类型
            
        Returns:
            bool: 是否上传成功
        """
        try:
            if not object_name:
                object_name = os.path.basename(file_path)
            
            # 获取文件大小
            file_size = os.path.getsize(file_path)
            
            # 上传文件
            self.client.fput_object(
                bucket_name=self.bucket_name,
                object_name=object_name,
                file_path=file_path,
                content_type=content_type,
                part_size=settings.minio_part_size
            )
            
            logger.info(f"文件 {file_path} 上传成功，对象名: {object_name}，大小: {file_size} 字节")
            return True
        except S3Error as e:
            logger.error(f"MinIO上传文件失败: {str(e)}")
            return False
        except Exception as e:
            logger.error(f"上传文件失败: {str(e)}")
            return False
    
    def upload_fileobj(self, file_data: BinaryIO, object_name: str, 
                      content_type: Optional[str] = None, 
                      length: Optional[int] = None) -> bool:
        """上传文件对象到MinIO
        
        Args:
            file_data: 文件对象（如BytesIO）
            object_name: 对象名称
            content_type: 内容类型
            length: 文件长度
            
        Returns:
            bool: 是否上传成功
        """
        try:
            # 上传文件对象
            self.client.put_object(
                bucket_name=self.bucket_name,
                object_name=object_name,
                data=file_data,
                length=length,
                content_type=content_type,
                part_size=settings.minio_part_size
            )
            
            logger.info(f"文件对象上传成功，对象名: {object_name}")
            return True
        except S3Error as e:
            logger.error(f"MinIO上传文件对象失败: {str(e)}")
            return False
        except Exception as e:
            logger.error(f"上传文件对象失败: {str(e)}")
            return False
    
    def download_file(self, object_name: str, file_path: str) -> bool:
        """从MinIO下载文件
        
        Args:
            object_name: 对象名称
            file_path: 本地文件路径
            
        Returns:
            bool: 是否下载成功
        """
        try:
            # 下载文件
            self.client.fget_object(
                bucket_name=self.bucket_name,
                object_name=object_name,
                file_path=file_path
            )
            
            logger.info(f"对象 {object_name} 下载成功，保存为: {file_path}")
            return True
        except S3Error as e:
            logger.error(f"MinIO下载文件失败: {str(e)}")
            return False
        except Exception as e:
            logger.error(f"下载文件失败: {str(e)}")
            return False
    
    def get_object(self, object_name: str) -> Optional[BinaryIO]:
        """获取对象内容
        
        Args:
            object_name: 对象名称
            
        Returns:
            BinaryIO: 对象内容的文件对象
        """
        try:
            # 获取对象
            response = self.client.get_object(
                bucket_name=self.bucket_name,
                object_name=object_name
            )
            
            logger.info(f"成功获取对象: {object_name}")
            return response
        except S3Error as e:
            logger.error(f"MinIO获取对象失败: {str(e)}")
            return None
        except Exception as e:
            logger.error(f"获取对象失败: {str(e)}")
            return None
    
    def remove_object(self, object_name: str) -> bool:
        """删除对象
        
        Args:
            object_name: 对象名称
            
        Returns:
            bool: 是否删除成功
        """
        try:
            # 删除对象
            self.client.remove_object(
                bucket_name=self.bucket_name,
                object_name=object_name
            )
            
            logger.info(f"对象 {object_name} 删除成功")
            return True
        except S3Error as e:
            logger.error(f"MinIO删除对象失败: {str(e)}")
            return False
        except Exception as e:
            logger.error(f"删除对象失败: {str(e)}")
            return False
    
    def list_objects(self, prefix: Optional[str] = None, 
                    recursive: bool = False) -> List[Dict[str, Any]]:
        """列出存储桶中的对象
        
        Args:
            prefix: 前缀过滤
            recursive: 是否递归列出
            
        Returns:
            List[Dict]: 对象列表
        """
        try:
            # 列出对象
            objects = self.client.list_objects(
                bucket_name=self.bucket_name,
                prefix=prefix,
                recursive=recursive
            )
            
            # 转换为列表
            object_list = []
            for obj in objects:
                object_list.append({
                    "name": obj.object_name,
                    "size": obj.size,
                    "last_modified": obj.last_modified,
                    "etag": obj.etag,
                    "content_type": obj.content_type
                })
            
            logger.info(f"成功列出 {len(object_list)} 个对象")
            return object_list
        except S3Error as e:
            logger.error(f"MinIO列出对象失败: {str(e)}")
            return []
        except Exception as e:
            logger.error(f"列出对象失败: {str(e)}")
            return []
    
    def get_presigned_url(self, object_name: str, expires: int = 3600) -> Optional[str]:
        """生成预签名URL
        
        Args:
            object_name: 对象名称
            expires: 过期时间（秒）
            
        Returns:
            str: 预签名URL
        """
        try:
            # 生成预签名URL
            url = self.client.presigned_get_object(
                bucket_name=self.bucket_name,
                object_name=object_name,
                expires=timedelta(seconds=expires)
            )
            
            logger.info(f"生成对象 {object_name} 的预签名URL成功，有效期: {expires} 秒")
            return url
        except S3Error as e:
            logger.error(f"MinIO生成预签名URL失败: {str(e)}")
            return None
        except Exception as e:
            logger.error(f"生成预签名URL失败: {str(e)}")
            return None
    
    def get_object_info(self, object_name: str) -> Optional[Dict[str, Any]]:
        """获取对象信息
        
        Args:
            object_name: 对象名称
            
        Returns:
            Dict: 对象信息
        """
        try:
            # 获取对象信息
            stat = self.client.stat_object(
                bucket_name=self.bucket_name,
                object_name=object_name
            )
            
            # 转换为字典
            info = {
                "name": stat.object_name,
                "size": stat.size,
                "last_modified": stat.last_modified,
                "etag": stat.etag,
                "content_type": stat.content_type,
                "metadata": dict(stat.metadata)
            }
            
            logger.info(f"获取对象 {object_name} 信息成功")
            return info
        except S3Error as e:
            logger.error(f"MinIO获取对象信息失败: {str(e)}")
            return None
        except Exception as e:
            logger.error(f"获取对象信息失败: {str(e)}")
            return None
    
    def check_health(self) -> bool:
        """检查MinIO连接健康状态
        
        Returns:
            bool: 连接是否健康
        """
        try:
            # 检查存储桶是否存在
            exists = self.client.bucket_exists(self.bucket_name)
            if exists:
                logger.info("MinIO连接健康检查通过")
                return True
            else:
                logger.warning(f"MinIO存储桶 {self.bucket_name} 不存在")
                return False
        except Exception as e:
            logger.error(f"MinIO健康检查失败: {str(e)}")
            return False
    
    def close(self):
        """关闭MinIO连接"""
        try:
            # MinIO客户端不需要显式关闭
            logger.info("MinIO客户端已关闭")
        except Exception as e:
            logger.error(f"关闭MinIO客户端时出错: {str(e)}")

# 导出全局实例
minio_client = MinioClient()