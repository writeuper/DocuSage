from sqlalchemy import Column, Integer, String, Text, DateTime, ForeignKey, Boolean, Float, JSON
from sqlalchemy.orm import relationship
from sqlalchemy.sql import func
from datetime import datetime

from models.db import Base


class Document(Base):
    """文档模型"""
    __tablename__ = "documents"
    
    id = Column(Integer, primary_key=True, index=True)
    document_id = Column(String(255), unique=True, nullable=False, index=True)
    user_id = Column(Integer, nullable=False, index=True)
    title = Column(String(500), nullable=False)
    filename = Column(String(500), nullable=False)
    file_path = Column(Text, nullable=False)
    file_size = Column(Integer)  # bytes
    file_type = Column(String(100))
    mime_type = Column(String(100))
    content_hash = Column(String(255), nullable=False, index=True)
    total_pages = Column(Integer, default=0)
    status = Column(String(50), default="uploaded", nullable=False, index=True)
    # uploaded, processing, processed, failed
    
    language = Column(String(50), default="en")
    metadata = Column(JSON, default={})
    error_message = Column(Text)
    
    # 统计信息
    total_chunks = Column(Integer, default=0)
    processed_chunks = Column(Integer, default=0)
    
    # 时间戳
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    updated_at = Column(DateTime(timezone=True), onupdate=func.now())
    processed_at = Column(DateTime(timezone=True), nullable=True)
    
    # 关联
    chunks = relationship("DocumentChunk", back_populates="document", cascade="all, delete-orphan")
    tasks = relationship("ProcessingTask", back_populates="document", cascade="all, delete-orphan")
    
    def __repr__(self):
        return f"<Document(id={self.id}, document_id={self.document_id}, title={self.title})>"


class DocumentChunk(Base):
    """文档分块模型"""
    __tablename__ = "document_chunks"
    
    id = Column(Integer, primary_key=True, index=True)
    chunk_id = Column(String(255), unique=True, nullable=False, index=True)
    document_id = Column(Integer, ForeignKey("documents.id"), nullable=False, index=True)
    content = Column(Text, nullable=False)
    content_type = Column(String(50), default="text")  # text, table, image
    
    # 位置信息
    page_number = Column(Integer, default=1)
    start_pos = Column(Integer, default=0)
    end_pos = Column(Integer, default=0)
    
    # 分块统计
    word_count = Column(Integer, default=0)
    token_count = Column(Integer, default=0)
    
    # 嵌入向量ID（在向量数据库中）
    embedding_id = Column(String(255), nullable=True)
    embedding_status = Column(String(50), default="pending", index=True)
    # pending, embedding, embedded, failed
    
    # 元数据
    metadata = Column(JSON, default={})
    
    # 时间戳
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    updated_at = Column(DateTime(timezone=True), onupdate=func.now())
    
    # 关联
    document = relationship("Document", back_populates="chunks")
    
    def __repr__(self):
        return f"<DocumentChunk(id={self.id}, chunk_id={self.chunk_id}, document_id={self.document_id})>"


class ChunkMetadata(Base):
    """分块元数据扩展"""
    __tablename__ = "chunk_metadata"
    
    id = Column(Integer, primary_key=True, index=True)
    chunk_id = Column(Integer, ForeignKey("document_chunks.id"), nullable=False, unique=True)
    
    # 语义信息
    keywords = Column(JSON, default=[])
    entities = Column(JSON, default={})
    topics = Column(JSON, default=[])
    
    # 格式化信息
    has_tables = Column(Boolean, default=False)
    has_images = Column(Boolean, default=False)
    has_lists = Column(Boolean, default=False)
    
    # 引用信息
    citations = Column(JSON, default=[])
    references = Column(JSON, default=[])
    
    # 时间戳
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    updated_at = Column(DateTime(timezone=True), onupdate=func.now())
    
    def __repr__(self):
        return f"<ChunkMetadata(id={self.id}, chunk_id={self.chunk_id})>"


class DocumentStats(Base):
    """文档统计模型"""
    __tablename__ = "document_stats"
    
    id = Column(Integer, primary_key=True, index=True)
    document_id = Column(Integer, ForeignKey("documents.id"), nullable=False, unique=True)
    
    # 统计信息
    total_words = Column(Integer, default=0)
    total_tokens = Column(Integer, default=0)
    avg_reading_time = Column(Float, default=0.0)  # 分钟
    
    # 处理信息
    processing_time = Column(Float, default=0.0)  # 秒
    embedding_time = Column(Float, default=0.0)  # 秒
    
    # 访问统计
    access_count = Column(Integer, default=0)
    last_accessed_at = Column(DateTime(timezone=True), nullable=True)
    
    # 时间戳
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    updated_at = Column(DateTime(timezone=True), onupdate=func.now())
    
    def __repr__(self):
        return f"<DocumentStats(id={self.id}, document_id={self.document_id})>"