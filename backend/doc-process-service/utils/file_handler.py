import os
import uuid
import hashlib
import logging
from typing import Dict, Any, BinaryIO, Optional
from pathlib import Path

from config.config import settings

logger = logging.getLogger(__name__)


class FileHandler:
    """文件处理工具类"""
    
    @staticmethod
    def generate_unique_filename(original_filename: str) -> str:
        """生成唯一文件名"""
        ext = os.path.splitext(original_filename)[1]
        unique_id = str(uuid.uuid4())
        return f"{unique_id}{ext}"
    
    @staticmethod
    def calculate_file_hash(file_path: str, chunk_size: int = 8192) -> str:
        """计算文件哈希值"""
        sha256 = hashlib.sha256()
        try:
            with open(file_path, 'rb') as f:
                while True:
                    data = f.read(chunk_size)
                    if not data:
                        break
                    sha256.update(data)
            return sha256.hexdigest()
        except Exception as e:
            logger.error(f"Failed to calculate file hash for {file_path}: {e}")
            raise
    
    @staticmethod
    def save_uploaded_file(file: BinaryIO, original_filename: str) -> Dict[str, Any]:
        """保存上传的文件"""
        try:
            # 检查文件扩展名
            ext = os.path.splitext(original_filename)[1].lower()
            if ext not in settings.allowed_extensions:
                raise ValueError(f"File type {ext} not allowed")
            
            # 生成唯一文件名
            unique_filename = FileHandler.generate_unique_filename(original_filename)
            file_path = os.path.join(settings.upload_dir, unique_filename)
            
            # 保存文件
            with open(file_path, 'wb') as f:
                file.seek(0)
                f.write(file.read())
            
            # 获取文件信息
            file_size = os.path.getsize(file_path)
            content_hash = FileHandler.calculate_file_hash(file_path)
            
            # 验证文件大小
            if file_size > settings.max_file_size_mb * 1024 * 1024:
                os.remove(file_path)  # 删除过大的文件
                raise ValueError(f"File size exceeds {settings.max_file_size_mb}MB limit")
            
            return {
                "file_path": file_path,
                "filename": unique_filename,
                "original_filename": original_filename,
                "file_size": file_size,
                "content_hash": content_hash,
                "file_extension": ext
            }
            
        except Exception as e:
            logger.error(f"Failed to save uploaded file {original_filename}: {e}")
            raise
    
    @staticmethod
    def delete_file(file_path: str) -> bool:
        """删除文件"""
        try:
            if os.path.exists(file_path):
                os.remove(file_path)
                logger.info(f"File deleted: {file_path}")
                return True
            return False
        except Exception as e:
            logger.error(f"Failed to delete file {file_path}: {e}")
            return False
    
    @staticmethod
    def get_file_info(file_path: str) -> Optional[Dict[str, Any]]:
        """获取文件信息"""
        try:
            if not os.path.exists(file_path):
                return None
            
            file_stat = os.stat(file_path)
            return {
                "size": file_stat.st_size,
                "created_at": file_stat.st_ctime,
                "modified_at": file_stat.st_mtime,
                "is_file": os.path.isfile(file_path)
            }
        except Exception as e:
            logger.error(f"Failed to get file info for {file_path}: {e}")
            return None
    
    @staticmethod
    def validate_file(file_path: str) -> bool:
        """验证文件是否有效"""
        try:
            # 检查文件是否存在
            if not os.path.exists(file_path):
                logger.error(f"File does not exist: {file_path}")
                return False
            
            # 检查文件大小
            file_size = os.path.getsize(file_path)
            if file_size > settings.max_file_size_mb * 1024 * 1024:
                logger.error(f"File size exceeds limit: {file_path}")
                return False
            
            # 检查文件扩展名
            ext = os.path.splitext(file_path)[1].lower()
            if ext not in settings.allowed_extensions:
                logger.error(f"File type not allowed: {file_path}")
                return False
            
            return True
        except Exception as e:
            logger.error(f"Failed to validate file {file_path}: {e}")
            return False
    
    @staticmethod
    def create_directory(path: str) -> bool:
        """创建目录"""
        try:
            Path(path).mkdir(parents=True, exist_ok=True)
            logger.info(f"Directory created or exists: {path}")
            return True
        except Exception as e:
            logger.error(f"Failed to create directory {path}: {e}")
            return False
    
    @staticmethod
    def get_directory_size(path: str) -> int:
        """获取目录大小（字节）"""
        total_size = 0
        try:
            for dirpath, dirnames, filenames in os.walk(path):
                for filename in filenames:
                    filepath = os.path.join(dirpath, filename)
                    if os.path.exists(filepath):
                        total_size += os.path.getsize(filepath)
        except Exception as e:
            logger.error(f"Failed to get directory size for {path}: {e}")
        return total_size
    
    @staticmethod
    def clean_old_files(directory: str, days_old: int = 30) -> int:
        """清理指定天数前的文件"""
        import time
        deleted_count = 0
        current_time = time.time()
        cutoff_time = current_time - (days_old * 86400)  # 86400 seconds in a day
        
        try:
            for filename in os.listdir(directory):
                filepath = os.path.join(directory, filename)
                if os.path.isfile(filepath):
                    file_modified = os.path.getmtime(filepath)
                    if file_modified < cutoff_time:
                        os.remove(filepath)
                        deleted_count += 1
                        logger.info(f"Deleted old file: {filepath}")
        except Exception as e:
            logger.error(f"Failed to clean old files in {directory}: {e}")
        
        logger.info(f"Cleaned {deleted_count} old files from {directory}")
        return deleted_count


# 导出实例
file_handler = FileHandler()