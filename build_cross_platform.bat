@echo off
chcp 65001 >nul
echo ========================================
echo 数据清洗服务交叉编译脚本
echo ========================================
echo.

if not exist "dist" mkdir dist

echo 1. 编译 Windows amd64...
set GOOS=windows
set GOARCH=amd64
go build -o dist/data-cleaning-service-windows-amd64.exe cmd/service/main.go
if %ERRORLEVEL% NEQ 0 goto error

echo 2. 编译 Windows 386...
set GOOS=windows
set GOARCH=386
go build -o dist/data-cleaning-service-windows-386.exe cmd/service/main.go
if %ERRORLEVEL% NEQ 0 goto error

echo 3. 编译 Linux amd64...
set GOOS=linux
set GOARCH=amd64
go build -o dist/data-cleaning-service-linux-amd64 cmd/service/main.go
if %ERRORLEVEL% NEQ 0 goto error

echo 4. 编译 Linux arm64...
set GOOS=linux
set GOARCH=arm64
go build -o dist/data-cleaning-service-linux-arm64 cmd/service/main.go
if %ERRORLEVEL% NEQ 0 goto error

echo 5. 编译 macOS amd64...
set GOOS=darwin
set GOARCH=amd64
go build -o dist/data-cleaning-service-darwin-amd64 cmd/service/main.go
if %ERRORLEVEL% NEQ 0 goto error

echo 6. 编译 macOS arm64 (Apple Silicon)...
set GOOS=darwin
set GOARCH=arm64
go build -o dist/data-cleaning-service-darwin-arm64 cmd/service/main.go
if %ERRORLEVEL% NEQ 0 goto error

echo.
echo [成功] 所有平台编译完成！
echo.
echo 输出目录: dist\
echo.
dir /b dist
echo.
echo ========================================
goto end

:error
echo.
echo [错误] 编译失败！
exit /b 1

:end
pause
