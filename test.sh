#!/bin/bash

echo "=== 微服务测试脚本 ==="
echo ""

# 等待服务启动
sleep 3

echo "0. 测试服务注册中心 - 查看所有已注册的服务"
curl -s http://localhost:8080/services | python3 -m json.tool
echo ""
echo ""

echo "1. 测试服务注册中心 - 发现用户服务"
curl -s http://localhost:8080/discover?name=user-service | python3 -m json.tool
echo ""
echo ""

echo "2. 测试用户服务 - 获取用户信息"
curl -s http://localhost:8081/user?id=1 | python3 -m json.tool
echo ""
echo ""

echo "3. 测试用户服务 - 列出所有用户"
curl -s http://localhost:8081/user | python3 -m json.tool
echo ""
echo ""

echo "4. 测试订单服务 - 获取订单信息"
curl -s http://localhost:8082/order?id=1 | python3 -m json.tool
echo ""
echo ""

echo "5. 测试订单服务 - 获取用户的订单列表"
curl -s http://localhost:8082/order?user_id=1 | python3 -m json.tool
echo ""
echo ""

echo "6. 测试服务间调用 - 获取订单（包含用户信息，通过服务发现）"
curl -s http://localhost:8082/order/with-user?id=1 | python3 -m json.tool
echo ""
echo ""

echo "7. 测试创建订单（会通过服务发现调用用户服务验证用户）"
curl -s -X POST http://localhost:8082/order \
  -H "Content-Type: application/json" \
  -d '{"user_id":1,"amount":399.99,"items":["新商品A","新商品B"]}' | python3 -m json.tool
echo ""
echo ""

echo "=== 测试完成 ==="

