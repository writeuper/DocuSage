from sqlalchemy import Column, Integer, String, Text, DateTime, ForeignKey, JSON, Float
from sqlalchemy.orm import relationship
from sqlalchemy.sql import func
from datetime import datetime

from models.db import Base


class ProcessingTask(Base):
    """处理任务模型"""
    __tablename__ = "processing_tasks"
    
    id = Column(Integer, primary_key=True, index=True)
    task_id = Column(String(255), unique=True, nullable=False, index=True)
    document_id = Column(Integer, ForeignKey("documents.id"), nullable=True, index=True)
    user_id = Column(Integer, nullable=False, index=True)
    
    # 任务信息
    task_type = Column(String(100), nullable=False)  # upload, parse, chunk, embed, index
    status = Column(String(50), default="pending", nullable=False, index=True)
    # pending, processing, completed, failed, cancelled
    
    # 进度信息
    progress = Column(Float, default=0.0)  # 0.0-100.0
    total_steps = Column(Integer, default=1)
    completed_steps = Column(Integer, default=0)
    
    # 错误信息
    error_message = Column(Text)
    retry_count = Column(Integer, default=0)
    
    # 配置参数
    params = Column(JSON, default={})
    
    # 时间戳
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    updated_at = Column(DateTime(timezone=True), onupdate=func.now())
    started_at = Column(DateTime(timezone=True), nullable=True)
    completed_at = Column(DateTime(timezone=True), nullable=True)
    
    # 性能指标
    execution_time = Column(Float, default=0.0)  # 秒
    resource_usage = Column(JSON, default={})
    
    # 关联
    document = relationship("Document", back_populates="tasks")
    subtasks = relationship("SubTask", back_populates="parent_task", cascade="all, delete-orphan")
    
    def __repr__(self):
        return f"<ProcessingTask(id={self.id}, task_id={self.task_id}, type={self.task_type}, status={self.status})>"


class SubTask(Base):
    """子任务模型"""
    __tablename__ = "subtasks"
    
    id = Column(Integer, primary_key=True, index=True)
    subtask_id = Column(String(255), unique=True, nullable=False, index=True)
    task_id = Column(Integer, ForeignKey("processing_tasks.id"), nullable=False, index=True)
    
    # 子任务信息
    subtask_type = Column(String(100), nullable=False)  # 更具体的任务类型
    target_id = Column(String(255), nullable=True)  # 目标ID，如chunk_id
    
    # 状态信息
    status = Column(String(50), default="pending", nullable=False, index=True)
    # pending, processing, completed, failed
    
    # 错误信息
    error_message = Column(Text)
    
    # 进度信息
    progress = Column(Float, default=0.0)
    
    # 配置参数
    params = Column(JSON, default={})
    
    # 时间戳
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    updated_at = Column(DateTime(timezone=True), onupdate=func.now())
    started_at = Column(DateTime(timezone=True), nullable=True)
    completed_at = Column(DateTime(timezone=True), nullable=True)
    
    # 执行信息
    worker_id = Column(String(255), nullable=True)
    execution_time = Column(Float, default=0.0)
    
    # 关联
    parent_task = relationship("ProcessingTask", back_populates="subtasks")
    
    def __repr__(self):
        return f"<SubTask(id={self.id}, subtask_id={self.subtask_id}, type={self.subtask_type})>"


class TaskQueue(Base):
    """任务队列模型"""
    __tablename__ = "task_queue"
    
    id = Column(Integer, primary_key=True, index=True)
    task_id = Column(String(255), nullable=False, unique=True, index=True)
    priority = Column(Integer, default=0)  # 0为最高优先级
    
    # 队列信息
    queue_name = Column(String(100), default="default", nullable=False, index=True)
    scheduled_at = Column(DateTime(timezone=True), server_default=func.now())
    retries_remaining = Column(Integer, default=3)
    
    # 时间戳
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    updated_at = Column(DateTime(timezone=True), onupdate=func.now())
    
    def __repr__(self):
        return f"<TaskQueue(id={self.id}, task_id={self.task_id}, queue={self.queue_name})>"


class ProcessingLog(Base):
    """处理日志模型"""
    __tablename__ = "processing_logs"
    
    id = Column(Integer, primary_key=True, index=True)
    task_id = Column(String(255), nullable=True, index=True)
    document_id = Column(Integer, nullable=True, index=True)
    user_id = Column(Integer, nullable=True, index=True)
    
    # 日志信息
    level = Column(String(20), nullable=False)  # INFO, WARNING, ERROR, DEBUG
    message = Column(Text, nullable=False)
    details = Column(JSON, default={})
    
    # 时间戳
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    
    def __repr__(self):
        return f"<ProcessingLog(id={self.id}, level={self.level}, message={self.message[:50]}...)>"