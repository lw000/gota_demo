#!/bin/bash

echo "========================================"
echo "数据清洗服务编译脚本 (Linux)"
echo "========================================"
echo

export GOOS=linux
export GOARCH=amd64

echo "开始编译..."
go build -o data-cleaning-service cmd/service/main.go

if [ $? -eq 0 ]; then
    echo
    echo "[成功] 编译完成！"
    echo "输出文件: data-cleaning-service"
    echo
    echo "运行服务:"
    echo "  ./data-cleaning-service"
    echo
    echo "或指定配置文件:"
    echo "  ./data-cleaning-service -config ./config.toml"
    echo
    chmod +x data-cleaning-service
else
    echo
    echo "[错误] 编译失败！"
    exit 1
fi

echo "========================================"
