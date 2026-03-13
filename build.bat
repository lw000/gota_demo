@echo off
chcp 65001 >nul
echo ========================================
echo 数据清洗服务编译脚本
echo ========================================
echo.

set GOOS=windows
set GOARCH=amd64

echo 开始编译...
go build -o data-cleaning-service.exe cmd/service/main.go

if %ERRORLEVEL% EQU 0 (
    echo.
    echo [成功] 编译完成！
    echo 输出文件: data-cleaning-service.exe
    echo.
    echo 运行服务:
    echo   .\data-cleaning-service.exe
    echo.
    echo 或指定配置文件:
    echo   .\data-cleaning-service.exe -config .\config.toml
) else (
    echo.
    echo [错误] 编译失败！
    exit /b 1
)

echo ========================================
pause
