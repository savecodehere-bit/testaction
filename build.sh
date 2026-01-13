#!/bin/bash

echo "开始编译微服务..."

# 编译服务注册中心
echo "编译服务注册中心..."
cd center_service
go build -o ../bin/center_service main.go
if [ $? -eq 0 ]; then
    echo "✓ 服务注册中心编译成功: bin/center_service"
else
    echo "✗ 服务注册中心编译失败"
    exit 1
fi

# 编译用户服务
echo "编译用户服务..."
cd ../user_service
go build -o ../bin/user_service main.go
if [ $? -eq 0 ]; then
    echo "✓ 用户服务编译成功: bin/user_service"
else
    echo "✗ 用户服务编译失败"
    exit 1
fi

# 编译订单服务
echo "编译订单服务..."
cd ../order_service
go build -o ../bin/order_service main.go
if [ $? -eq 0 ]; then
    echo "✓ 订单服务编译成功: bin/order_service"
else
    echo "✗ 订单服务编译失败"
    exit 1
fi

# 编译网关服务
echo "编译网关服务..."
cd ../gateway_service
go build -o ../bin/gateway_service main.go
if [ $? -eq 0 ]; then
    echo "✓ 网关服务编译成功: bin/gateway_service"
else
    echo "✗ 网关服务编译失败"
    exit 1
fi

# 编译测试客户端
echo "编译测试客户端..."
cd ../client
go build -o ../bin/client main.go
if [ $? -eq 0 ]; then
    echo "✓ 测试客户端编译成功: bin/client"
else
    echo "✗ 测试客户端编译失败"
    exit 1
fi

cd ..
echo ""
echo "所有服务编译完成！"
echo "可执行文件位置:"
echo "  - bin/center_service    (服务注册中心，端口8080)"
echo "  - bin/user_service      (用户服务，端口8081)"
echo "  - bin/order_service     (订单服务，端口8082)"
echo "  - bin/gateway_service   (API网关服务，端口8083)"
echo "  - bin/client            (API测试客户端)"
echo ""
echo "运行方式:"
echo "  ./bin/center_service    # 启动服务注册中心（端口8080）"
echo "  ./bin/user_service      # 启动用户服务（端口8081）"
echo "  ./bin/order_service     # 启动订单服务（端口8082）"
echo "  ./bin/gateway_service   # 启动API网关服务（端口8083）"
echo "  ./bin/client            # 启动API测试客户端"
echo ""
echo "或使用 ./start.sh 一键启动所有服务"

