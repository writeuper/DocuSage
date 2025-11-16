import logging
import os
import json
from typing import Dict, List, Any, Optional, Tuple
from datetime import datetime

# 导入必要的库
import PyPDF2  # PDF解析
import docx  # Word文档解析

from config.config import settings

logger = logging.getLogger(__name__)


class DocumentParser:
    """文档解析工具类"""
    
    # 支持的文件类型
    SUPPORTED_FORMATS = {
        '.pdf': 'pdf',
        '.doc': 'word',
        '.docx': 'word',
        '.txt': 'text',
        '.md': 'markdown',
        '.json': 'json',
    }
    
    @classmethod
    def parse_document(cls, file_path: str, file_name: str) -> Dict[str, Any]:
        """解析文档，根据文件类型调用相应的解析方法"""
        if not os.path.exists(file_path):
            raise FileNotFoundError(f"文件不存在: {file_path}")
        
        # 获取文件扩展名
        _, ext = os.path.splitext(file_name)
        ext = ext.lower()
        
        file_type = cls.SUPPORTED_FORMATS.get(ext)
        if not file_type:
            raise ValueError(f"不支持的文件类型: {ext}")
        
        try:
            if file_type == 'pdf':
                return cls._parse_pdf(file_path)
            elif file_type == 'word':
                return cls._parse_word(file_path)
            elif file_type == 'text':
                return cls._parse_text(file_path)
            elif file_type == 'markdown':
                return cls._parse_markdown(file_path)
            elif file_type == 'json':
                return cls._parse_json(file_path)
            else:
                raise ValueError(f"未实现的文件类型解析: {file_type}")
        except Exception as e:
            logger.error(f"解析文档 {file_name} 时出错: {str(e)}")
            raise
    
    @classmethod
    def _parse_pdf(cls, file_path: str) -> Dict[str, Any]:
        """解析PDF文件"""
        result = {
            "content": [],
            "metadata": {
                "format": "pdf",
                "total_pages": 0,
                "parsed_at": datetime.now().isoformat(),
                "has_images": False,  # 简化处理，实际应用中可能需要更复杂的图像检测
            }
        }
        
        try:
            with open(file_path, 'rb') as file:
                reader = PyPDF2.PdfReader(file)
                result["metadata"]["total_pages"] = len(reader.pages)
                
                # 尝试提取元数据
                if reader.metadata:
                    pdf_metadata = {
                        "title": reader.metadata.get('/Title', '').strip(),
                        "author": reader.metadata.get('/Author', '').strip(),
                        "creator": reader.metadata.get('/Creator', '').strip(),
                        "producer": reader.metadata.get('/Producer', '').strip(),
                        "creation_date": reader.metadata.get('/CreationDate', '').strip(),
                        "modification_date": reader.metadata.get('/ModDate', '').strip(),
                    }
                    result["metadata"]["pdf_metadata"] = pdf_metadata
                
                # 提取每页文本
                for page_num in range(len(reader.pages)):
                    page = reader.pages[page_num]
                    text = page.extract_text()
                    
                    if text:
                        result["content"].append({
                            "page_number": page_num + 1,
                            "text": text.strip(),
                            "words_count": len(text.split()),
                            "characters_count": len(text)
                        })
                    else:
                        logger.warning(f"PDF第 {page_num + 1} 页无法提取文本")
            
            logger.info(f"PDF文件解析完成，共 {result['metadata']['total_pages']} 页")
            return result
            
        except PyPDF2.errors.PdfReadError as e:
            logger.error(f"PDF读取错误: {str(e)}")
            raise ValueError(f"无效的PDF文件: {str(e)}")
        except Exception as e:
            logger.error(f"解析PDF时出错: {str(e)}")
            raise
    
    @classmethod
    def _parse_word(cls, file_path: str) -> Dict[str, Any]:
        """解析Word文档（.doc和.docx）"""
        result = {
            "content": [],
            "metadata": {
                "format": "word",
                "parsed_at": datetime.now().isoformat(),
                "has_images": False,  # 简化处理
            }
        }
        
        try:
            # 处理.docx文件
            if file_path.lower().endswith('.docx'):
                doc = docx.Document(file_path)
                
                # 提取元数据
                if hasattr(doc.core_properties, '_core_properties'):
                    word_metadata = {
                        "title": doc.core_properties.title,
                        "author": doc.core_properties.author,
                        "created": doc.core_properties.created.isoformat() if doc.core_properties.created else None,
                        "modified": doc.core_properties.modified.isoformat() if doc.core_properties.modified else None,
                        "subject": doc.core_properties.subject,
                    }
                    result["metadata"]["word_metadata"] = word_metadata
                
                # 提取文本
                all_text = []
                for paragraph in doc.paragraphs:
                    if paragraph.text.strip():
                        all_text.append(paragraph.text)
                
                full_text = '\n'.join(all_text)
                
                result["content"].append({
                    "page_number": 1,  # Word文档不直接提供页码
                    "text": full_text,
                    "words_count": len(full_text.split()),
                    "characters_count": len(full_text)
                })
            else:
                # .doc文件的简化处理（实际可能需要使用python-docx2txt或其他库）
                raise NotImplementedError("完整的.doc文件支持需要额外的库")
            
            logger.info("Word文档解析完成")
            return result
            
        except Exception as e:
            logger.error(f"解析Word文档时出错: {str(e)}")
            raise
    
    @classmethod
    def _parse_text(cls, file_path: str) -> Dict[str, Any]:
        """解析纯文本文件"""
        result = {
            "content": [],
            "metadata": {
                "format": "text",
                "parsed_at": datetime.now().isoformat(),
            }
        }
        
        try:
            # 尝试不同编码读取
            encodings = ['utf-8', 'latin-1', 'cp1252', 'gbk', 'gb2312']
            text_content = None
            
            for encoding in encodings:
                try:
                    with open(file_path, 'r', encoding=encoding) as file:
                        text_content = file.read()
                    break
                except UnicodeDecodeError:
                    continue
            
            if text_content is None:
                raise ValueError("无法解码文本文件，尝试了多种编码")
            
            # 按段落分割
            paragraphs = text_content.split('\n\n')
            
            # 计算大致的页数（假设每页50行）
            line_count = text_content.count('\n')
            page_count = (line_count // 50) + 1
            
            result["metadata"]["total_pages"] = page_count
            
            # 将内容按"页"组织（简化处理）
            lines_per_page = 50
            current_page = 1
            current_page_lines = []
            
            for line in text_content.split('\n'):
                current_page_lines.append(line)
                
                if len(current_page_lines) >= lines_per_page:
                    page_text = '\n'.join(current_page_lines)
                    result["content"].append({
                        "page_number": current_page,
                        "text": page_text,
                        "words_count": len(page_text.split()),
                        "characters_count": len(page_text)
                    })
                    current_page += 1
                    current_page_lines = []
            
            # 添加最后一页
            if current_page_lines:
                page_text = '\n'.join(current_page_lines)
                result["content"].append({
                    "page_number": current_page,
                    "text": page_text,
                    "words_count": len(page_text.split()),
                    "characters_count": len(page_text)
                })
            
            logger.info(f"文本文件解析完成，约 {page_count} 页")
            return result
            
        except Exception as e:
            logger.error(f"解析文本文件时出错: {str(e)}")
            raise
    
    @classmethod
    def _parse_markdown(cls, file_path: str) -> Dict[str, Any]:
        """解析Markdown文件"""
        # Markdown本质上也是文本，所以可以复用文本解析的逻辑
        result = cls._parse_text(file_path)
        result["metadata"]["format"] = "markdown"
        
        # 可以添加额外的Markdown特定处理
        # 例如提取标题结构、代码块等
        
        logger.info("Markdown文件解析完成")
        return result
    
    @classmethod
    def _parse_json(cls, file_path: str) -> Dict[str, Any]:
        """解析JSON文件"""
        result = {
            "content": [],
            "metadata": {
                "format": "json",
                "parsed_at": datetime.now().isoformat(),
            }
        }
        
        try:
            with open(file_path, 'r', encoding='utf-8') as file:
                data = json.load(file)
            
            # 将JSON数据转换为格式化的字符串
            formatted_content = json.dumps(data, ensure_ascii=False, indent=2)
            
            result["content"].append({
                "page_number": 1,
                "text": formatted_content,
                "words_count": len(formatted_content.split()),
                "characters_count": len(formatted_content)
            })
            
            # 记录JSON结构信息
            if isinstance(data, dict):
                result["metadata"]["keys"] = list(data.keys())
                result["metadata"]["structure_type"] = "object"
            elif isinstance(data, list):
                result["metadata"]["item_count"] = len(data)
                result["metadata"]["structure_type"] = "array"
            
            logger.info("JSON文件解析完成")
            return result
            
        except json.JSONDecodeError as e:
            logger.error(f"JSON解析错误: {str(e)}")
            raise ValueError(f"无效的JSON文件: {str(e)}")
        except Exception as e:
            logger.error(f"解析JSON文件时出错: {str(e)}")
            raise
    
    @classmethod
    def get_supported_formats(cls) -> List[str]:
        """获取支持的文件格式列表"""
        return list(cls.SUPPORTED_FORMATS.keys())
    
    @classmethod
    def is_supported(cls, file_extension: str) -> bool:
        """检查文件扩展名是否被支持"""
        return file_extension.lower() in cls.SUPPORTED_FORMATS


# 导出实例
document_parser = DocumentParser()