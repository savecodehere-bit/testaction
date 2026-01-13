# 简单微服务示例

这是一个用 Go 编写的简单微服务示例，展示如何在裸机上直接运行微服务，**无需 Docker 或 Kubernetes**。

## 架构说明

本项目包含四个服务：

1. **服务注册中心 (Service Registry)** - 端口 8080
   - 服务注册与发现
   - 服务心跳检测
   - 服务列表查询

2. **用户服务 (User Service)** - 端口 8081
   - 管理用户信息
   - 提供用户 CRUD 操作
   - 启动时自动注册到注册中心
   - 定期发送心跳保持在线状态

3. **订单服务 (Order Service)** - 端口 8082
   - 管理订单信息
   - **通过服务发现**找到用户服务，而不是硬编码URL
   - 通过 HTTP 调用用户服务验证用户是否存在
   - 演示服务间通信和服务发现

4. **API网关服务 (Gateway Service)** - 端口 8083
   - 统一的API入口，所有外部请求通过网关访问
   - 自动发现后端服务（用户服务、订单服务）
   - 请求路由和转发
   - 健康检查接口
   - 演示API网关模式

## 快速开始

### 1. 编译服务

**Linux/macOS:**
```bash
chmod +x build.sh
./build.sh
```

**Windows:**
```cmd
build.bat
```

这会编译四个服务，生成可执行文件到 `bin/` 目录：
- `bin/center_service` - 服务注册中心（带GUI窗口）
- `bin/user_service` - 用户服务（带GUI窗口）
- `bin/order_service` - 订单服务（带GUI窗口）
- `bin/gateway_service` - API网关服务（带GUI窗口）

### 2. 启动服务

每个服务都会显示一个**跨平台的GUI窗口**，实时显示服务状态和日志。

**方式一：使用启动脚本（推荐）**

Linux/macOS:
```bash
chmod +x start.sh
./start.sh
```

Windows:
```cmd
start.bat
```

**方式二：手动启动（每个服务会弹出独立窗口）**

Linux/macOS:
```bash
# 终端1：启动服务注册中心（会弹出GUI窗口）
./bin/center_service

# 终端2：启动用户服务（会弹出GUI窗口）
./bin/user_service

# 终端3：启动订单服务（会弹出GUI窗口）
./bin/order_service

# 终端4：启动API网关服务（会弹出GUI窗口）
./bin/gateway_service
```

Windows:
```cmd
# 双击或在命令行运行
bin\center_service.exe
bin\user_service.exe
bin\order_service.exe
bin\gateway_service.exe
```

## GUI 窗口功能

每个服务都带有**跨平台GUI窗口**（支持 Windows、macOS、Linux），窗口显示：

1. **服务注册中心窗口**：
   - 实时显示已注册的服务列表
   - 服务注册/注销日志
   - 服务状态统计

2. **用户服务窗口**：
   - 实时日志输出（API调用记录）
   - 服务状态（端口、用户数、注册中心连接状态）

3. **订单服务窗口**：
   - 实时日志输出（API调用记录、服务发现过程）
   - 服务状态（端口、订单数、用户服务发现状态）

4. **API网关服务窗口**：
   - 实时日志输出（请求转发记录、服务发现过程）
   - 服务状态（端口、已发现服务数、注册中心连接状态）

## API 使用示例

### 服务注册中心 API

```bash
# 查看所有已注册的服务
curl http://localhost:8080/services

# 发现特定服务
curl http://localhost:8080/discover?name=user-service
```

### 用户服务 API

```bash
# 获取用户信息
curl http://localhost:8081/user?id=1

# 列出所有用户
curl http://localhost:8081/user

# 创建新用户
curl -X POST http://localhost:8081/user \
  -H "Content-Type: application/json" \
  -d '{"name":"王五","email":"wangwu@example.com"}'
```

### 订单服务 API

```bash
# 获取订单信息
curl http://localhost:8082/order?id=1

# 获取用户的订单列表
curl http://localhost:8082/order?user_id=1

# 获取订单（包含用户信息，演示服务间调用）
curl http://localhost:8082/order/with-user?id=1

# 创建订单（会自动验证用户是否存在）
curl -X POST http://localhost:8082/order \
  -H "Content-Type: application/json" \
  -d '{"user_id":1,"amount":299.99,"items":["商品D","商品E"]}'
```

## 微服务的核心特点

1. **独立部署**：每个服务都是独立的可执行文件，可以单独启动、停止、更新
2. **独立端口**：每个服务运行在不同的端口，互不干扰
3. **服务发现**：通过服务注册中心自动发现其他服务，无需硬编码服务地址
4. **服务间通信**：通过 HTTP REST API 进行通信（订单服务调用用户服务）
5. **心跳机制**：服务定期发送心跳，注册中心自动清理失效服务
6. **无依赖容器**：直接运行二进制文件，不需要 Docker、K8s 等复杂工具

## 为什么不需要 Docker/K8s？

- **简单场景**：对于简单的微服务，直接运行可执行文件就足够了
- **开发调试**：裸机运行更容易调试，启动更快
- **资源占用**：没有容器和编排系统的开销
- **学习成本**：不需要学习 Docker/K8s 的复杂概念

## 什么时候需要 Docker/K8s？

- 需要环境隔离
- 需要复杂的服务编排和自动扩缩容
- 需要服务发现、负载均衡等高级功能
- 需要跨机器部署和管理大量服务

但对于简单的微服务架构，**裸机运行完全够用**！

## 项目结构

```
.
├── center_service/         # 服务注册中心源码
│   └── main.go
├── user_service/          # 用户服务源码
│   └── main.go
├── order_service/         # 订单服务源码
│   └── main.go
├── bin/                   # 编译后的可执行文件（自动生成）
│   ├── center_service
│   ├── user_service
│   └── order_service
├── build.sh               # 编译脚本
├── start.sh               # 启动脚本
├── test.sh                # 测试脚本
├── go.mod                 # Go 模块文件
└── README.md              # 本文件
```

## 技术栈

- **语言**：Go 1.21+
- **HTTP 框架**：标准库 `net/http`
- **GUI 框架**：Fyne v2（跨平台GUI库，支持 Windows、macOS、Linux）
- **数据存储**：内存（map），仅用于演示

## Windows 编译说明

在 Windows 上编译需要先安装 Go，然后运行：

```cmd
go mod tidy
build.bat
```

编译后的可执行文件在 `bin\` 目录下，可以直接双击运行，每个服务会弹出独立的GUI窗口。

## 服务发现的工作原理

1. **服务注册**：用户服务启动时，自动向注册中心注册自己的信息（名称、地址、端口）
2. **服务发现**：订单服务启动时，从注册中心查询用户服务的地址
3. **心跳机制**：用户服务每10秒发送一次心跳，保持在线状态
4. **自动清理**：注册中心检测到服务30秒未心跳，自动将其从注册表中移除

这样，即使服务地址改变，其他服务也能自动发现新的地址，**无需修改配置或重启**！

## 扩展建议

如果想进一步扩展，可以考虑：
- 添加数据库（MySQL/PostgreSQL）
- 添加更完善的服务发现（Consul/etcd）
- 添加 API 网关
- 添加日志和监控
- 添加配置管理
- 添加负载均衡

但对于理解微服务的核心概念（独立部署、服务发现、服务间通信），当前这个简单示例已经足够了！

# testaction
