import logging
import json
from typing import Any, Optional, Dict, List
import redis
from datetime import datetime, timedelta

from config.config import settings

logger = logging.getLogger(__name__)


class RedisClient:
    """Redis客户端工具类"""
    
    def __init__(self):
        """初始化Redis连接"""
        self.redis_url = settings.redis_url
        self.redis_client = None
        self._connect()
        logger.info(f"Redis连接池配置: 最大连接数={settings.redis_max_connections}, 最小空闲连接={settings.redis_min_idle_connections}")
    
    def _connect(self):
        """建立Redis连接"""
        try:
            # 构建Redis配置
            redis_config = {
                'decode_responses': True,
                'socket_connect_timeout': settings.redis_connect_timeout,
                'socket_timeout': settings.redis_socket_timeout,
                'socket_keepalive': True,
                'health_check_interval': settings.redis_health_check_interval,
                'max_connections': settings.redis_max_connections,
                'retry_on_timeout': True,
                'retry_on_error': [redis.ConnectionError, redis.TimeoutError],
                'backoff_factor': 0.5
            }
            
            # 添加TLS配置（如果启用）
            if settings.redis_tls_enabled:
                redis_config['ssl'] = True
                redis_config['ssl_cert_reqs'] = settings.redis_tls_verify_mode
                logger.info("Redis TLS连接已启用")
            
            self.redis_client = redis.from_url(
                self.redis_url,
                **redis_config
            )
            
            # 测试连接
            self.redis_client.ping()
            logger.info("成功连接到Redis")
            logger.info(f"Redis连接池状态: 最大连接数={settings.redis_max_connections}, 最小空闲连接={settings.redis_min_idle_connections}")
        except redis.ConnectionError as e:
            logger.error(f"Redis连接失败: {str(e)}")
            raise
        except Exception as e:
            logger.error(f"初始化Redis客户端时出错: {str(e)}")
            raise
    
    def get(self, key: str) -> Optional[str]:
        """获取键值"""
        try:
            return self.redis_client.get(key)
        except Exception as e:
            logger.error(f"获取Redis键 {key} 失败: {str(e)}")
            return None
    
    def set(self, key: str, value: Any, expire_seconds: Optional[int] = None) -> bool:
        """设置键值"""
        try:
            if expire_seconds:
                return self.redis_client.setex(key, expire_seconds, value)
            else:
                return self.redis_client.set(key, value)
        except Exception as e:
            logger.error(f"设置Redis键 {key} 失败: {str(e)}")
            return False
    
    def delete(self, key: str) -> bool:
        """删除键"""
        try:
            return bool(self.redis_client.delete(key))
        except Exception as e:
            logger.error(f"删除Redis键 {key} 失败: {str(e)}")
            return False
    
    def exists(self, key: str) -> bool:
        """检查键是否存在"""
        try:
            return bool(self.redis_client.exists(key))
        except Exception as e:
            logger.error(f"检查Redis键 {key} 失败: {str(e)}")
            return False
    
    def expire(self, key: str, seconds: int) -> bool:
        """设置键过期时间"""
        try:
            return bool(self.redis_client.expire(key, seconds))
        except Exception as e:
            logger.error(f"设置Redis键 {key} 过期时间失败: {str(e)}")
            return False
    
    def get_ttl(self, key: str) -> int:
        """获取键的剩余生存时间"""
        try:
            return self.redis_client.ttl(key)
        except Exception as e:
            logger.error(f"获取Redis键 {key} 过期时间失败: {str(e)}")
            return -1
    
    # JSON相关操作
    def get_json(self, key: str) -> Optional[Dict[str, Any]]:
        """获取JSON格式的键值"""
        try:
            data = self.redis_client.get(key)
            if data:
                return json.loads(data)
            return None
        except json.JSONDecodeError as e:
            logger.error(f"解析Redis键 {key} 的JSON数据失败: {str(e)}")
            return None
        except Exception as e:
            logger.error(f"获取Redis键 {key} 的JSON数据失败: {str(e)}")
            return None
    
    def set_json(self, key: str, value: Dict[str, Any], expire_seconds: Optional[int] = None) -> bool:
        """设置JSON格式的键值"""
        try:
            json_data = json.dumps(value, ensure_ascii=False, default=str)
            return self.set(key, json_data, expire_seconds)
        except json.JSONDecodeError as e:
            logger.error(f"序列化数据为JSON失败: {str(e)}")
            return False
        except Exception as e:
            logger.error(f"设置Redis键 {key} 的JSON数据失败: {str(e)}")
            return False
    
    # 哈希表操作
    def hget(self, name: str, key: str) -> Optional[str]:
        """获取哈希表字段值"""
        try:
            return self.redis_client.hget(name, key)
        except Exception as e:
            logger.error(f"获取Redis哈希表 {name} 的字段 {key} 失败: {str(e)}")
            return None
    
    def hset(self, name: str, key: str, value: Any) -> bool:
        """设置哈希表字段值"""
        try:
            return bool(self.redis_client.hset(name, key, value))
        except Exception as e:
            logger.error(f"设置Redis哈希表 {name} 的字段 {key} 失败: {str(e)}")
            return False
    
    def hgetall(self, name: str) -> Dict[str, str]:
        """获取哈希表所有字段和值"""
        try:
            return self.redis_client.hgetall(name)
        except Exception as e:
            logger.error(f"获取Redis哈希表 {name} 所有字段失败: {str(e)}")
            return {}
    
    def hdel(self, name: str, *keys) -> bool:
        """删除哈希表字段"""
        try:
            return bool(self.redis_client.hdel(name, *keys))
        except Exception as e:
            logger.error(f"删除Redis哈希表 {name} 的字段失败: {str(e)}")
            return False
    
    # 列表操作
    def lpush(self, name: str, *values) -> int:
        """向列表左侧推入值"""
        try:
            return self.redis_client.lpush(name, *values)
        except Exception as e:
            logger.error(f"向Redis列表 {name} 左侧推入值失败: {str(e)}")
            return 0
    
    def rpush(self, name: str, *values) -> int:
        """向列表右侧推入值"""
        try:
            return self.redis_client.rpush(name, *values)
        except Exception as e:
            logger.error(f"向Redis列表 {name} 右侧推入值失败: {str(e)}")
            return 0
    
    def lpop(self, name: str) -> Optional[str]:
        """从列表左侧弹出值"""
        try:
            return self.redis_client.lpop(name)
        except Exception as e:
            logger.error(f"从Redis列表 {name} 左侧弹出值失败: {str(e)}")
            return None
    
    def rpop(self, name: str) -> Optional[str]:
        """从列表右侧弹出值"""
        try:
            return self.redis_client.rpop(name)
        except Exception as e:
            logger.error(f"从Redis列表 {name} 右侧弹出值失败: {str(e)}")
            return None
    
    def llen(self, name: str) -> int:
        """获取列表长度"""
        try:
            return self.redis_client.llen(name)
        except Exception as e:
            logger.error(f"获取Redis列表 {name} 长度失败: {str(e)}")
            return 0
    
    # 集合操作
    def sadd(self, name: str, *values) -> int:
        """向集合添加成员"""
        try:
            return self.redis_client.sadd(name, *values)
        except Exception as e:
            logger.error(f"向Redis集合 {name} 添加成员失败: {str(e)}")
            return 0
    
    def srem(self, name: str, *values) -> int:
        """从集合移除成员"""
        try:
            return self.redis_client.srem(name, *values)
        except Exception as e:
            logger.error(f"从Redis集合 {name} 移除成员失败: {str(e)}")
            return 0
    
    def smembers(self, name: str) -> set:
        """获取集合所有成员"""
        try:
            return self.redis_client.smembers(name)
        except Exception as e:
            logger.error(f"获取Redis集合 {name} 所有成员失败: {str(e)}")
            return set()
    
    # 发布订阅
    def publish(self, channel: str, message: Any) -> int:
        """发布消息到频道"""
        try:
            if isinstance(message, (dict, list)):
                message = json.dumps(message, ensure_ascii=False)
            return self.redis_client.publish(channel, message)
        except Exception as e:
            logger.error(f"向Redis频道 {channel} 发布消息失败: {str(e)}")
            return 0
    
    # 文档处理相关的缓存方法
    def cache_document_status(self, document_id: str, status: str, metadata: Optional[Dict[str, Any]] = None) -> bool:
        """缓存文档处理状态"""
        key = f"doc_status:{document_id}"
        data = {
            "status": status,
            "updated_at": datetime.now().isoformat(),
        }
        if metadata:
            data.update(metadata)
        return self.set_json(key, data, expire_seconds=86400)  # 24小时过期
    
    def get_document_status(self, document_id: str) -> Optional[Dict[str, Any]]:
        """获取文档处理状态"""
        key = f"doc_status:{document_id}"
        return self.get_json(key)
    
    def cache_processing_task(self, task_id: str, task_data: Dict[str, Any]) -> bool:
        """缓存处理任务信息"""
        key = f"task:{task_id}"
        task_data["cached_at"] = datetime.now().isoformat()
        return self.set_json(key, task_data, expire_seconds=3600)  # 1小时过期
    
    def get_processing_task(self, task_id: str) -> Optional[Dict[str, Any]]:
        """获取处理任务信息"""
        key = f"task:{task_id}"
        return self.get_json(key)
    
    def add_to_processing_queue(self, document_id: str, priority: int = 0) -> bool:
        """添加文档到处理队列"""
        # 使用优先级分数（分数越小优先级越高）
        try:
            timestamp = datetime.now().timestamp()
            score = timestamp + priority  # 优先级高的文档分数低
            self.redis_client.zadd("doc_processing_queue", {document_id: score})
            logger.info(f"文档 {document_id} 已添加到处理队列，优先级: {priority}")
            return True
        except Exception as e:
            logger.error(f"添加文档 {document_id} 到处理队列失败: {str(e)}")
            return False
    
    def get_next_from_processing_queue(self) -> Optional[str]:
        """获取处理队列中的下一个文档"""
        try:
            # 获取分数最小的元素（优先级最高）
            result = self.redis_client.zpopmin("doc_processing_queue", 1)
            if result:
                document_id = result[0][0]
                logger.info(f"从处理队列获取文档: {document_id}")
                return document_id
            return None
        except Exception as e:
            logger.error(f"从处理队列获取文档失败: {str(e)}")
            return None
    
    def close(self):
        """关闭Redis连接"""
        if self.redis_client:
            try:
                # 获取连接池统计信息
                pool = self.redis_client.connection_pool
                active_connections = len([c for c in pool._available_connections if c is not None]) + len(pool._in_use_connections)
                logger.info(f"关闭Redis连接池，当前活跃连接数: {active_connections}")
                
                self.redis_client.close()
                logger.info("Redis连接已关闭")
            except Exception as e:
                logger.error(f"关闭Redis连接时出错: {str(e)}")
    
    def check_health(self) -> bool:
        """检查Redis连接健康状态"""
        try:
            if not self.redis_client:
                logger.warning("Redis客户端未初始化")
                return False
            
            # 执行ping操作检查连接
            result = self.redis_client.ping()
            if result:
                logger.debug("Redis连接健康检查通过")
                return True
            else:
                logger.warning("Redis ping返回非True值")
                return False
        except Exception as e:
            logger.error(f"Redis健康检查失败: {str(e)}")
            return False
    
    def get_connection_stats(self) -> Dict[str, int]:
        """获取Redis连接池统计信息"""
        try:
            if not self.redis_client:
                return {"error": "Redis客户端未初始化"}
            
            pool = self.redis_client.connection_pool
            stats = {
                "available_connections": len([c for c in pool._available_connections if c is not None]),
                "in_use_connections": len(pool._in_use_connections),
                "created_connections": pool._created_connections if hasattr(pool, '_created_connections') else 0
            }
            stats["total_connections"] = stats["available_connections"] + stats["in_use_connections"]
            return stats
        except Exception as e:
            logger.error(f"获取Redis连接池统计信息失败: {str(e)}")
            return {"error": str(e)}


# 导出全局实例
redis_client = RedisClient()