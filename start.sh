#!/bin/bash

# 创建bin目录
mkdir -p bin

# 检查是否已编译
if [ ! -f "bin/center_service" ] || [ ! -f "bin/user_service" ] || [ ! -f "bin/order_service" ] || [ ! -f "bin/gateway_service" ]; then
    echo "检测到可执行文件不存在，开始编译..."
    ./build.sh
fi

echo ""
echo "启动微服务（按顺序启动：注册中心 -> 用户服务 -> 订单服务 -> 网关服务）..."
echo ""

# 启动服务注册中心（后台运行）
echo "1. 启动服务注册中心（端口8080）..."
./bin/center_service &
CENTER_PID=$!
echo "   服务注册中心 PID: $CENTER_PID"
sleep 2

# 启动用户服务（后台运行）
echo "2. 启动用户服务（端口8081）..."
./bin/user_service &
USER_PID=$!
echo "   用户服务 PID: $USER_PID"
sleep 2

# 启动订单服务（后台运行）
echo "3. 启动订单服务（端口8082）..."
./bin/order_service &
ORDER_PID=$!
echo "   订单服务 PID: $ORDER_PID"
sleep 2

# 启动网关服务（后台运行）
echo "4. 启动API网关服务（端口8083）..."
./bin/gateway_service &
GATEWAY_PID=$!
echo "   API网关服务 PID: $GATEWAY_PID"

sleep 1

echo ""
echo "所有服务已启动！"
echo ""
echo "服务地址:"
echo "  服务注册中心: http://localhost:8080"
echo "  用户服务:     http://localhost:8081"
echo "  订单服务:     http://localhost:8082"
echo "  API网关:      http://localhost:8083"
echo ""
echo "查看所有已注册的服务:"
echo "  curl http://localhost:8080/services"
echo ""
echo "按 Ctrl+C 停止所有服务"
echo ""

# 等待中断信号
trap "echo ''; echo '正在停止服务...'; kill $CENTER_PID $USER_PID $ORDER_PID $GATEWAY_PID 2>/dev/null; exit" INT TERM

# 保持脚本运行
wait

