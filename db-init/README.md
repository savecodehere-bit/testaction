# 数据库初始化脚本

这个目录包含各个服务的数据库初始化SQL脚本。

## 脚本说明

- `init-user-db.sql` - 用户服务数据库初始化
  - 创建 `users` 表
  - 插入示例数据

- `init-order-db.sql` - 订单服务数据库初始化
  - 创建 `orders` 表
  - 插入示例数据

- `init-registry-db.sql` - 注册中心数据库初始化（可选）
  - 创建 `services` 表用于持久化服务注册信息

## 自动执行

这些脚本会在Docker容器首次启动时自动执行（通过 `docker-entrypoint-initdb.d` 目录）。

## 手动执行

如果需要手动执行：

```bash
# 用户数据库
mysql -h localhost -P 3306 -u user_service -puser_pass user_db < init-user-db.sql

# 订单数据库
mysql -h localhost -P 3307 -u order_service -porder_pass order_db < init-order-db.sql

# 注册中心数据库
mysql -h localhost -P 3308 -u registry_service -pregistry_pass registry_db < init-registry-db.sql
```

