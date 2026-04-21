@echo off
chcp 65001 > nul
title sfsEdgeStore 启动

echo ========================================
echo    sfsEdgeStore - 工业物联网边缘存储
echo ========================================
echo.
echo 正在启动 sfsEdgeStore...
echo.
echo Web 监控界面: http://localhost:8081
echo 停止方法: 关闭此窗口或按 Ctrl+C
echo.

sfsedgestore.exe

if errorlevel 1 (
    echo.
    echo [错误] 程序异常退出！
    pause
)
