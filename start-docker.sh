#!/bin/bash

echo "启动Docker微服务..."

# 检查Docker是否运行
if ! docker info > /dev/null 2>&1; then
    echo "错误: Docker未运行，请先启动Docker"
    exit 1
fi

# 检查网络是否存在
if ! docker network inspect microservice_microservice-network > /dev/null 2>&1; then
    echo "创建Docker网络..."
    docker network create microservice_microservice-network
fi

# 启动数据库（如果还没启动）
echo "检查数据库状态..."
if ! docker-compose ps | grep -q "user-db.*Up"; then
    echo "启动数据库..."
    docker-compose up -d
    echo "等待数据库启动..."
    sleep 5
fi

# 启动应用服务
echo "启动应用服务..."
docker-compose -f docker-compose.yml -f docker-compose.services.yml up -d

echo ""
echo "等待服务启动..."
sleep 5

# 检查服务状态
echo ""
echo "服务状态:"
docker-compose -f docker-compose.yml -f docker-compose.services.yml ps

echo ""
echo "服务地址:"
echo "  注册中心: http://localhost:8080"
echo "  用户服务: http://localhost:8081"
echo "  订单服务: http://localhost:8082"
echo "  API网关:  http://localhost:8083"
echo ""
echo "查看服务注册情况:"
echo "  curl http://localhost:8080/services"
echo ""
echo "查看日志:"
echo "  docker-compose -f docker-compose.services.yml logs -f"
echo ""
echo "停止服务:"
echo "  docker-compose -f docker-compose.yml -f docker-compose.services.yml down"

