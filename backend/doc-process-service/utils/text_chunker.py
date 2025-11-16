import logging
import re
from typing import List, Dict, Any, Tuple

from config.config import settings

logger = logging.getLogger(__name__)


class TextChunker:
    """文本分块工具类"""
    
    def __init__(self, chunk_size: int = None, overlap_size: int = None):
        """初始化文本分块器"""
        self.chunk_size = chunk_size or settings.chunk_size
        self.overlap_size = overlap_size or settings.overlap_size
        
        if self.overlap_size >= self.chunk_size:
            logger.warning("Overlap size should be less than chunk size. Adjusting overlap.")
            self.overlap_size = max(0, self.chunk_size // 10)
    
    def chunk_text(self, text: str, page_number: int = 1) -> List[Dict[str, Any]]:
        """将文本分块"""
        if not text:
            return []
        
        chunks = []
        start_pos = 0
        text_length = len(text)
        chunk_id = 0
        
        # 使用段落级别分块优先
        paragraphs = self._split_into_paragraphs(text)
        
        # 如果段落分块效果不好，使用滑动窗口分块
        if len(paragraphs) == 1 or any(len(p) > self.chunk_size * 1.5 for p in paragraphs):
            return self._sliding_window_chunk(text, page_number)
        
        # 段落合并分块
        current_chunk = ""
        current_start = 0
        
        for i, paragraph in enumerate(paragraphs):
            # 检查添加当前段落是否会超过块大小
            if len(current_chunk) + len(paragraph) > self.chunk_size:
                # 如果当前块不为空，保存
                if current_chunk:
                    chunks.append({
                        "content": current_chunk.strip(),
                        "start_pos": current_start,
                        "end_pos": current_start + len(current_chunk),
                        "page_number": page_number,
                        "chunk_id": chunk_id
                    })
                    chunk_id += 1
                
                # 如果当前段落本身就很大，需要进一步分块
                if len(paragraph) > self.chunk_size:
                    sub_chunks = self._split_large_paragraph(paragraph, current_start, page_number)
                    chunks.extend(sub_chunks)
                    chunk_id += len(sub_chunks)
                    current_chunk = ""
                    current_start = current_start + len(paragraph)
                else:
                    # 开始新块
                    current_chunk = paragraph
                    current_start = current_start + len(text[:current_start].count('\n'))
            else:
                # 添加到当前块
                current_chunk += paragraph
        
        # 保存最后一个块
        if current_chunk:
            chunks.append({
                "content": current_chunk.strip(),
                "start_pos": current_start,
                "end_pos": current_start + len(current_chunk),
                "page_number": page_number,
                "chunk_id": chunk_id
            })
        
        return chunks
    
    def _split_into_paragraphs(self, text: str) -> List[str]:
        """将文本分割成段落"""
        # 使用两个或更多换行符作为段落分隔符
        paragraphs = re.split(r'\n\s*\n', text)
        # 过滤空段落
        return [p.strip() + '\n' for p in paragraphs if p.strip()]
    
    def _sliding_window_chunk(self, text: str, page_number: int) -> List[Dict[str, Any]]:
        """使用滑动窗口分块"""
        chunks = []
        text_length = len(text)
        chunk_id = 0
        
        for i in range(0, text_length, self.chunk_size - self.overlap_size):
            # 计算结束位置
            end = min(i + self.chunk_size, text_length)
            
            # 尝试在句子边界处分割
            if end < text_length:
                end = self._find_sentence_boundary(text, end)
            
            chunk_text = text[i:end]
            
            # 只添加非空块
            if chunk_text.strip():
                chunks.append({
                    "content": chunk_text.strip(),
                    "start_pos": i,
                    "end_pos": end,
                    "page_number": page_number,
                    "chunk_id": chunk_id
                })
                chunk_id += 1
        
        return chunks
    
    def _split_large_paragraph(self, paragraph: str, start_pos: int, page_number: int) -> List[Dict[str, Any]]:
        """分割大型段落"""
        chunks = []
        para_length = len(paragraph)
        chunk_id = 0
        
        for i in range(0, para_length, self.chunk_size - self.overlap_size):
            end = min(i + self.chunk_size, para_length)
            end = self._find_sentence_boundary(paragraph, end)
            
            chunk_text = paragraph[i:end]
            if chunk_text.strip():
                chunks.append({
                    "content": chunk_text.strip(),
                    "start_pos": start_pos + i,
                    "end_pos": start_pos + end,
                    "page_number": page_number,
                    "chunk_id": chunk_id
                })
                chunk_id += 1
        
        return chunks
    
    def _find_sentence_boundary(self, text: str, position: int) -> int:
        """找到句子边界"""
        # 标点符号列表
        sentence_endings = ['.', '!', '?', '。', '！', '？']
        
        # 向前搜索最近的句子结束符
        for i in range(min(position + 50, len(text)) - 1, max(position - 50, 0), -1):
            if text[i] in sentence_endings:
                # 确保后面有空格或换行（通常句子结束后会有）
                if i + 1 < len(text) and (text[i + 1].isspace() or text[i + 1].isupper()):
                    return i + 1
        
        # 如果没有找到合适的边界，返回原始位置
        return position
    
    def calculate_chunk_stats(self, chunk: str) -> Dict[str, int]:
        """计算分块的统计信息"""
        # 计算单词数（简单实现）
        words = re.findall(r'\b\w+\b', chunk)
        word_count = len(words)
        
        # 估算token数（粗略估计，1个token≈0.75个单词）
        token_count = int(word_count * 1.33)
        
        # 字符数
        char_count = len(chunk)
        
        return {
            "word_count": word_count,
            "token_count": token_count,
            "char_count": char_count
        }
    
    def get_optimal_chunk_size(self, document_type: str) -> int:
        """根据文档类型获取最佳分块大小"""
        chunk_sizes = {
            "pdf": self.chunk_size,         # 默认PDF分块大小
            "doc": self.chunk_size,         # Word文档
            "docx": self.chunk_size,        # Word文档
            "txt": self.chunk_size,         # 纯文本
            "md": self.chunk_size * 0.8,    # Markdown略小
            "html": self.chunk_size * 1.2,  # HTML可能需要更大块
            "xml": self.chunk_size * 1.2     # XML可能需要更大块
        }
        
        return int(chunk_sizes.get(document_type.lower(), self.chunk_size))


# 导出实例
text_chunker = TextChunker()