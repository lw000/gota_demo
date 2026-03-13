@echo off
chcp 65001 >nul
echo ========================================
echo 数据清洗服务运行脚本
echo ========================================
echo.

if not exist "data-cleaning-service.exe" (
    echo [错误] 可执行文件不存在，请先运行 build.bat 编译项目
    pause
    exit /b 1
)

echo 启动数据清洗服务...
echo.

.\data-cleaning-service.exe -config .\config.toml

if %ERRORLEVEL% NEQ 0 (
    echo.
    echo [错误] 服务运行异常！退出码: %ERRORLEVEL%
)

echo.
pause
