# GitHub Actions 快速开始

## 一键编译Windows exe

### 步骤1：推送代码到GitHub

```bash
# 如果还没有GitHub仓库
git init
git add .
git commit -m "Initial commit"
git remote add origin https://github.com/你的用户名/你的仓库名.git
git push -u origin main
```

### 步骤2：触发编译

#### 方式A：手动触发（推荐首次使用）

1. 打开GitHub仓库页面
2. 点击 **Actions** 标签
3. 选择 **Build Windows Executables**
4. 点击 **Run workflow** → **Run workflow**

#### 方式B：推送tag自动编译

```bash
git tag v1.0.0
git push origin v1.0.0
```

#### 方式C：代码变更自动编译

```bash
# 修改代码后推送
git add .
git commit -m "Update code"
git push origin main
```

### 步骤3：下载编译好的exe

1. 等待编译完成（约2-5分钟）
2. 点击workflow运行记录
3. 滚动到底部，找到 **Artifacts**
4. 点击 **windows-binaries** 下载zip文件
5. 解压后即可使用

## 编译产物

下载的zip文件包含：
- `center_service.exe` - 服务注册中心
- `user_service.exe` - 用户服务
- `order_service.exe` - 订单服务
- `gateway_service.exe` - API网关服务
- `client.exe` - 测试客户端

## 常见问题

### Q: 编译失败怎么办？
A: 查看Actions页面的错误日志，通常是依赖问题或Go版本问题。

### Q: 找不到Artifacts？
A: 确保工作流运行完成，Artifacts会在运行完成后出现。

### Q: 可以编译其他平台吗？
A: 使用 `build-all-platforms.yml` 工作流可以编译Windows、Linux、macOS。

### Q: 如何自动创建Release？
A: 推送tag（如 `v1.0.0`）时会自动创建Release并上传exe文件。

## 示例命令

```bash
# 完整流程示例
git add .
git commit -m "Ready for release"
git tag v1.0.0
git push origin main
git push origin v1.0.0  # 触发编译和Release创建
```

