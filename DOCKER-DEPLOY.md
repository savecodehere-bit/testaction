# Docker部署指南

## 架构说明

本项目采用**混合部署**方案：
- **数据库**：Docker容器（docker-compose.yml）
- **应用服务**：Docker容器（docker-compose.services.yml）
- **服务发现**：由微服务系统自身实现（center-service）

## 服务发现机制

### 工作原理

1. **服务注册**：每个服务启动时，向注册中心注册自己的信息
   - 服务名（name）
   - 容器内地址（address）：使用Docker服务名
   - 端口（port）

2. **服务发现**：服务通过注册中心发现其他服务
   - 网关服务自动发现用户服务、订单服务
   - 订单服务自动发现用户服务

3. **Docker网络**：所有服务在同一Docker网络中，可以通过服务名互相访问

### 关键配置

```go
// 服务注册时使用Docker服务名作为address
serviceInfo := map[string]interface{}{
    "name":    "user-service",
    "address": "user-service",  // Docker服务名，不是localhost
    "port":    8081,
}
```

## 快速开始

### 1. 启动数据库

```bash
docker-compose up -d
```

### 2. 启动应用服务

```bash
# 开发环境（默认）
docker-compose -f docker-compose.yml -f docker-compose.services.yml up -d

# 或使用便捷脚本
./start-docker.sh
```

### 3. 查看服务状态

```bash
docker-compose ps
```

### 4. 查看日志

```bash
# 查看所有服务日志
docker-compose logs -f

# 查看特定服务日志
docker-compose logs -f user-service
```

## 环境隔离

### 开发环境（dev）

```bash
# 使用默认配置
docker-compose -f docker-compose.yml -f docker-compose.services.yml up -d

# 或使用override文件
cp docker-compose.override.dev.yml.example docker-compose.override.yml
docker-compose up -d
```

特点：
- 端口映射到主机（8080-8083）
- 日志级别：debug
- 支持代码热重载（可选）

### 测试环境（test）

```bash
docker-compose -f docker-compose.yml -f docker-compose.services.yml \
  -f docker-compose.test.yml up -d
```

### 生产环境（prod）

```bash
docker-compose -f docker-compose.yml -f docker-compose.services.yml \
  -f docker-compose.prod.yml up -d
```

特点：
- 资源限制
- 日志级别：info
- 可配置多实例（replicas）

## 服务发现配置

### 环境变量

每个服务通过环境变量配置注册中心地址：

```bash
REGISTRY_URL=http://center-service:8080  # Docker服务名
```

### 服务注册地址

服务注册时使用Docker服务名：

| 服务 | Docker服务名 | 注册地址 |
|------|-------------|---------|
| 注册中心 | `center-service` | `http://center-service:8080` |
| 用户服务 | `user-service` | `http://user-service:8081` |
| 订单服务 | `order-service` | `http://order-service:8082` |
| 网关服务 | `gateway-service` | `http://gateway-service:8083` |

### 外部访问

虽然容器内使用服务名，但外部访问仍使用localhost：

```bash
# 外部访问
curl http://localhost:8080/services  # 注册中心
curl http://localhost:8081/user       # 用户服务
curl http://localhost:8082/order      # 订单服务
curl http://localhost:8083/health     # 网关
```

## Docker网络

所有服务在同一个Docker网络中（`microservice-network`），可以：

1. **通过服务名访问**：`http://user-service:8081`
2. **通过容器名访问**：`http://microservice-user-service:8081`
3. **端口映射**：外部通过 `localhost:8081` 访问

## 多实例部署

### 启动多个用户服务实例

```bash
# 启动第二个用户服务实例
docker-compose -f docker-compose.services.yml up -d --scale user-service=2

# 查看实例
docker-compose ps | grep user-service
```

注意：需要修改服务注册逻辑支持多实例（当前版本只支持单实例）。

## 构建镜像

### 构建单个服务

```bash
docker build -f Dockerfile --build-arg SERVICE=user_service -t user-service:latest .
```

### 构建所有服务

```bash
# 通过docker-compose构建
docker-compose -f docker-compose.services.yml build
```

## 常用命令

```bash
# 启动所有服务
docker-compose -f docker-compose.yml -f docker-compose.services.yml up -d

# 停止所有服务
docker-compose -f docker-compose.yml -f docker-compose.services.yml down

# 重启服务
docker-compose -f docker-compose.services.yml restart user-service

# 查看日志
docker-compose -f docker-compose.services.yml logs -f user-service

# 进入容器
docker exec -it microservice-user-service sh

# 查看网络
docker network inspect microservice_microservice-network
```

## 故障排查

### 服务无法注册到注册中心

1. 检查注册中心是否运行：
   ```bash
   docker-compose ps center-service
   ```

2. 检查网络连接：
   ```bash
   docker exec -it microservice-user-service wget -O- http://center-service:8080/services
   ```

3. 查看服务日志：
   ```bash
   docker-compose logs user-service
   ```

### 服务无法发现其他服务

1. 检查服务是否已注册：
   ```bash
   curl http://localhost:8080/services
   ```

2. 检查服务名是否正确：
   - 注册时使用的服务名
   - 发现时使用的服务名
   - 必须完全一致

## 与直接运行对比

| 特性 | 直接运行 | Docker部署 |
|------|---------|-----------|
| 启动速度 | 秒级 | 分钟级（首次） |
| 环境隔离 | ❌ | ✅ |
| 资源控制 | ❌ | ✅ |
| 服务发现 | ✅ | ✅（需配置） |
| 调试难度 | 简单 | 中等 |
| 部署一致性 | ⚠️ | ✅ |

## 推荐使用场景

- **开发环境**：直接运行（`./start.sh`）
- **测试环境**：Docker部署（环境隔离）
- **生产环境**：Docker部署（资源控制、一致性）

