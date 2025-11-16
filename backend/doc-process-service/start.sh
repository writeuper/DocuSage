#!/bin/bash

# 文档处理服务启动脚本
# 使用方法: ./start.sh [api|worker|all]

# 错误时退出
set -e

echo "=== 文档处理服务启动脚本 ==="

# 检查Python环境
check_python() {
    echo "检查Python环境..."
    if ! command -v python3 &> /dev/null; then
        echo "错误: 未找到Python3，请先安装Python 3.8+"
        exit 1
    fi
    
    PYTHON_VERSION=$(python3 --version 2>&1 | awk '{print $2}')
    echo "找到Python版本: $PYTHON_VERSION"
}

# 创建虚拟环境
create_venv() {
    if [ ! -d "venv" ]; then
        echo "创建虚拟环境..."
        python3 -m venv venv
    else
        echo "虚拟环境已存在，跳过创建"
    fi
}

# 激活虚拟环境
activate_venv() {
    echo "激活虚拟环境..."
    source venv/bin/activate
}

# 安装依赖
install_deps() {
    echo "安装依赖..."
    pip install --upgrade pip
    pip install -r requirements.txt
    
    # 安装额外的依赖用于PDF和Office文档处理
    pip install PyPDF2 python-docx openpyxl pandas
    
    # 安装FastAPI和ASGI服务器
    pip install fastapi uvicorn
    
    # 安装Redis客户端
    pip install redis
    
    echo "依赖安装完成"
}

# 创建必要的目录
create_directories() {
    echo "创建必要的目录..."
    mkdir -p storage/uploads
    mkdir -p storage/processed
    mkdir -p logs
    
    # 确保目录权限正确
    chmod -R 755 storage
    chmod -R 755 logs
    
    echo "目录创建完成"
}

# 检查环境变量文件
check_env_file() {
    if [ ! -f ".env" ]; then
        echo "警告: .env 文件不存在，请确保已创建环境变量配置文件"
        echo "可以复制 .env.example 并修改相关配置"
    else
        echo "环境变量文件存在"
    fi
}

# 启动API服务
start_api() {
    echo "启动API服务..."
    echo "服务将在 http://0.0.0.0:8001 启动"
    echo "API文档地址: http://0.0.0.0:8001/docs"
    
    # 以开发模式启动API服务
    uvicorn main:app --host 0.0.0.0 --port 8001 --reload
}

# 启动Celery Worker
start_worker() {
    echo "启动Celery Worker..."
    
    # 激活虚拟环境
    activate_venv
    
    # 启动Celery Worker，设置日志级别和并发数
    celery -A tasks.document_tasks worker --loglevel=info --concurrency=4
}

# 启动Celery Beat（用于定时任务）
start_beat() {
    echo "启动Celery Beat..."
    
    # 激活虚拟环境
    activate_venv
    
    # 启动Celery Beat
    celery -A tasks.document_tasks beat --loglevel=info
}

# 启动所有服务
start_all() {
    echo "启动所有服务..."
    echo "注意: 此模式将在前台启动API服务，请在不同终端分别启动Worker和Beat"
    echo ""
    echo "在其他终端执行以下命令启动Worker:"
    echo "  source venv/bin/activate"
    echo "  celery -A tasks.document_tasks worker --loglevel=info --concurrency=4"
    echo ""
    echo "在另一个终端执行以下命令启动Beat:"
    echo "  source venv/bin/activate"
    echo "  celery -A tasks.document_tasks beat --loglevel=info"
    echo ""
    
    # 启动API服务
    start_api
}

# 显示帮助信息
show_help() {
    echo "用法: ./start.sh [command]"
    echo ""
    echo "命令:"
    echo "  setup      - 仅设置环境（创建虚拟环境、安装依赖）"
    echo "  api        - 启动API服务"
    echo "  worker     - 启动Celery Worker"
    echo "  beat       - 启动Celery Beat（定时任务）"
    echo "  all        - 启动所有服务（API在前台，Worker和Beat需在其他终端启动）"
    echo "  help       - 显示此帮助信息"
    echo ""
    echo "示例:"
    echo "  ./start.sh setup      # 设置环境"
    echo "  ./start.sh api        # 启动API服务"
    echo "  ./start.sh worker     # 启动Worker"
    echo ""
}

# 主函数
main() {
    # 检查Python环境
    check_python
    
    # 根据参数执行不同的操作
    case "$1" in
        setup)
            create_venv
            activate_venv
            install_deps
            create_directories
            check_env_file
            echo ""
            echo "环境设置完成！可以使用以下命令启动服务:"
            echo "  ./start.sh api        # 启动API服务"
            echo "  ./start.sh worker     # 启动Celery Worker"
            echo "  ./start.sh beat       # 启动Celery Beat"
            ;;
        api)
            create_venv
            activate_venv
            create_directories
            check_env_file
            start_api
            ;;
        worker)
            create_venv
            activate_venv
            create_directories
            start_worker
            ;;
        beat)
            create_venv
            activate_venv
            start_beat
            ;;
        all)
            create_venv
            activate_venv
            install_deps
            create_directories
            check_env_file
            start_all
            ;;
        help|
        *)
            show_help
            ;;
    esac
}

# 执行主函数
main "$1"