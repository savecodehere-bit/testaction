# 数据库配置说明

## 架构设计：每个服务独立数据库

遵循微服务最佳实践：**Database per Service**（每个服务独立数据库）

## 数据库分配

| 服务 | 数据库 | MySQL端口 | Redis端口 | 说明 |
|------|--------|-----------|-----------|------|
| 用户服务 | `user_db` | 3306 | 6379 | 用户数据存储和缓存 |
| 订单服务 | `order_db` | 3307 | 6380 | 订单数据存储和缓存 |
| 注册中心 | `registry_db` | 3308 | - | 服务注册信息持久化（可选） |

## 快速启动

### 1. 启动所有数据库

```bash
docker-compose up -d
```

这会启动：
- MySQL数据库（3个实例）
- Redis缓存（2个实例）

### 2. 检查数据库状态

```bash
# 查看运行状态
docker-compose ps

# 查看日志
docker-compose logs -f user-db
```

### 3. 连接数据库

**用户服务数据库：**
```bash
mysql -h localhost -P 3306 -u user_service -puser_pass user_db
```

**订单服务数据库：**
```bash
mysql -h localhost -P 3307 -u order_service -porder_pass order_db
```

**注册中心数据库：**
```bash
mysql -h localhost -P 3308 -u registry_service -pregistry_pass registry_db
```

## 数据库连接配置

### 用户服务配置

```go
// user_service/main.go
dsn := "user_service:user_pass@tcp(localhost:3306)/user_db?charset=utf8mb4&parseTime=True&loc=Local"
db, err := sql.Open("mysql", dsn)

// Redis连接
rdb := redis.NewClient(&redis.Options{
    Addr: "localhost:6379",
})
```

### 订单服务配置

```go
// order_service/main.go
dsn := "order_service:order_pass@tcp(localhost:3307)/order_db?charset=utf8mb4&parseTime=True&loc=Local"
db, err := sql.Open("mysql", dsn)

// Redis连接
rdb := redis.NewClient(&redis.Options{
    Addr: "localhost:6380",
})
```

## 数据库初始化

数据库首次启动时会自动执行 `db-init/` 目录下的SQL脚本：

- `init-user-db.sql` - 创建用户表并插入示例数据
- `init-order-db.sql` - 创建订单表并插入示例数据
- `init-registry-db.sql` - 创建服务注册表

## 数据持久化

所有数据库数据存储在Docker volumes中：

- `user_db_data` - 用户数据库数据
- `order_db_data` - 订单数据库数据
- `registry_db_data` - 注册中心数据库数据
- `user_redis_data` - 用户服务Redis数据
- `order_redis_data` - 订单服务Redis数据

即使删除容器，数据也不会丢失。

## 停止和清理

```bash
# 停止所有数据库（保留数据）
docker-compose stop

# 停止并删除容器（保留数据）
docker-compose down

# 停止并删除容器和数据
docker-compose down -v
```

## 环境变量配置（可选）

如果需要通过环境变量配置，可以创建 `.env` 文件：

```bash
# 用户服务
USER_DB_HOST=localhost
USER_DB_PORT=3306
USER_DB_NAME=user_db
USER_DB_USER=user_service
USER_DB_PASSWORD=user_pass
USER_REDIS_HOST=localhost
USER_REDIS_PORT=6379

# 订单服务
ORDER_DB_HOST=localhost
ORDER_DB_PORT=3307
ORDER_DB_NAME=order_db
ORDER_DB_USER=order_service
ORDER_DB_PASSWORD=order_pass
ORDER_REDIS_HOST=localhost
ORDER_REDIS_PORT=6380
```

## 优势

1. **数据隔离**：每个服务的数据完全独立，互不影响
2. **独立扩展**：可以根据服务负载独立扩展数据库
3. **技术选型灵活**：不同服务可以使用不同的数据库技术
4. **故障隔离**：一个服务的数据库故障不影响其他服务
5. **团队协作**：不同团队可以独立管理自己的数据库

## 注意事项

1. **端口映射**：每个MySQL实例使用不同端口避免冲突
2. **数据一致性**：跨服务数据一致性需要通过API调用保证
3. **事务处理**：跨服务事务需要使用分布式事务方案（如Saga模式）

