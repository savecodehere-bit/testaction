# Docker部署代码修改指南

## 概述

为了让服务支持Docker部署并保持服务发现功能，需要修改服务代码以支持：
1. 从环境变量读取配置
2. 在Docker环境中使用Docker服务名注册
3. 环境隔离配置

## 已创建的工具

已创建 `pkg/config/config.go` 提供配置辅助函数。

## 需要修改的服务

### 1. 用户服务 (user_service/main.go)

#### 修改点1：导入config包

```go
import (
    // ... 其他导入
    "ttt/pkg/config"
)
```

#### 修改点2：修改main函数中的配置读取

```go
func main() {
    // ... GUI初始化代码 ...

    // 从环境变量读取配置
    port := config.GetEnvInt("PORT", 8081)
    registryURL := config.GetEnv("REGISTRY_URL", "http://localhost:8080")
    serviceAddress := config.GetServiceAddress("USER_SERVICE", "localhost")
    
    // 如果PORT环境变量设置了，使用环境变量的值
    if portEnv := os.Getenv("PORT"); portEnv != "" {
        if p, err := strconv.Atoi(portEnv); err == nil {
            port = p
        }
    }

    service := NewUserService(port, registryURL, logContainer, logScroll, statusLabel)
    service.serviceAddress = serviceAddress  // 需要添加这个字段
```

#### 修改点3：修改RegisterToRegistry方法

```go
// RegisterToRegistry 注册到服务注册中心
func (us *UserService) RegisterToRegistry() {
    if us.registryURL == "" {
        return
    }

    // 使用配置的服务地址
    address := us.serviceAddress
    if address == "" {
        // 检查是否在Docker环境中
        if os.Getenv("ENV") != "" {
            address = "user-service"  // Docker服务名
        } else {
            address = "localhost"
        }
    }

    serviceInfo := map[string]interface{}{
        "name":    "user-service",
        "address": address,  // 使用配置的地址
        "port":    us.port,
    }

    // ... 其余代码不变 ...
}
```

#### 修改点4：添加serviceAddress字段

```go
type UserService struct {
    users        map[int]*User
    mu           sync.RWMutex
    nextID       int
    port         int
    registryURL  string
    serviceAddress string  // 新增字段
    logContainer *fyne.Container
    logScroll    *container.Scroll
    statusLabel  *widget.Label
}
```

### 2. 订单服务 (order_service/main.go)

同样的修改，但使用：
- `ORDER_SERVICE_ADDRESS` 环境变量
- Docker服务名：`order-service`

### 3. 网关服务 (gateway_service/main.go)

同样的修改，但使用：
- `GATEWAY_SERVICE_ADDRESS` 环境变量
- Docker服务名：`gateway-service`

### 4. 注册中心 (center_service/main.go)

注册中心不需要注册自己，但需要：
- 从环境变量读取端口：`PORT`（默认8080）
- 支持环境变量配置

## 环境变量配置

### Docker Compose中的环境变量

已在 `docker-compose.services.yml` 中配置：

```yaml
environment:
  - PORT=8081
  - REGISTRY_URL=http://center-service:8080
  - ENV=dev
  - USER_SERVICE_ADDRESS=user-service  # Docker服务名
```

### 直接运行时的环境变量（可选）

```bash
export PORT=8081
export REGISTRY_URL=http://localhost:8080
export USER_SERVICE_ADDRESS=localhost
./bin/user_service
```

## 服务发现工作原理

### Docker环境

1. **服务注册**：
   ```go
   {
       "name": "user-service",
       "address": "user-service",  // Docker服务名
       "port": 8081
   }
   ```

2. **服务发现**：
   - 网关服务通过注册中心获取：`http://user-service:8081`
   - Docker网络自动解析服务名到容器IP

3. **外部访问**：
   - 通过端口映射：`http://localhost:8081`

### 直接运行环境

1. **服务注册**：
   ```go
   {
       "name": "user-service",
       "address": "localhost",
       "port": 8081
   }
   ```

2. **服务发现**：
   - 网关服务通过注册中心获取：`http://localhost:8081`

## 快速测试

### 1. 测试Docker部署

```bash
# 启动数据库
docker-compose up -d

# 启动应用服务
docker-compose -f docker-compose.yml -f docker-compose.services.yml up -d

# 检查服务注册
curl http://localhost:8080/services
```

### 2. 测试直接运行（向后兼容）

```bash
# 直接运行，应该使用localhost注册
./start.sh

# 检查服务注册
curl http://localhost:8080/services
```

## 注意事项

1. **向后兼容**：修改后的代码应该同时支持Docker和直接运行
2. **环境检测**：通过环境变量`ENV`或`HOSTNAME`判断是否在Docker中
3. **服务名一致性**：Docker服务名必须与注册时的服务名一致
4. **网络配置**：确保所有服务在同一Docker网络中

## 完整示例代码

参考 `pkg/config/config.go` 中的 `GetServiceAddress` 函数实现。

