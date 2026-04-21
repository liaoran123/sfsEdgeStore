#!/bin/bash

echo "========================================"
echo "   sfsEdgeStore - 工业物联网边缘存储"
echo "========================================"
echo ""
echo "正在启动 sfsEdgeStore..."
echo ""
echo "Web 监控界面: http://localhost:8081"
echo "停止方法: Ctrl+C"
echo ""

chmod +x sfsedgestore
./sfsedgestore

if [ $? -ne 0 ]; then
    echo ""
    echo "[错误] 程序异常退出！"
    read -p "按 Enter 键继续..."
fi
