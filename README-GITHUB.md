# GitHub Actions 使用说明

## 快速开始

### 1. 提交代码到GitHub

```bash
# 初始化git（如果还没有）
git init

# 添加文件
git add .

# 提交
git commit -m "Add GitHub Actions for Windows build"

# 添加远程仓库（替换为你的仓库地址）
git remote add origin https://github.com/你的用户名/你的仓库名.git

# 推送
git push -u origin main
```

### 2. 触发编译

#### 方式A：手动触发（最简单）

1. 打开GitHub仓库页面
2. 点击 **Actions** 标签
3. 选择 **Build Windows Executables**
4. 点击 **Run workflow** → 选择分支 → **Run workflow**

#### 方式B：推送tag自动编译

```bash
git tag v1.0.0
git push origin v1.0.0
```

这会自动：
- ✅ 编译所有Windows exe文件
- ✅ 创建GitHub Release
- ✅ 上传exe文件到Release

### 3. 下载编译好的文件

**从Artifacts下载：**
1. Actions页面 → 点击最新的workflow运行
2. 滚动到底部 → Artifacts部分
3. 点击 **windows-binaries** 下载zip

**从Release下载：**
1. 仓库页面 → Releases
2. 找到对应版本
3. 下载附件中的exe文件

## 编译的服务

- `center_service.exe` - 服务注册中心（端口8080）
- `user_service.exe` - 用户服务（端口8081）
- `order_service.exe` - 订单服务（端口8082）
- `gateway_service.exe` - API网关服务（端口8083）
- `client.exe` - 测试客户端

## 工作流说明

### build-windows.yml
- **触发方式**：手动、tag推送、代码变更
- **运行环境**：Windows Latest
- **编译时间**：约2-5分钟
- **产物保留**：30天

### build-all-platforms.yml
- **触发方式**：手动、tag推送
- **运行环境**：Windows、Linux、macOS
- **编译时间**：约5-10分钟（并行）

## 常见问题

**Q: 编译失败？**
- 检查Go版本（需要1.21+）
- 查看Actions日志中的错误信息
- 确保go.mod和go.sum已提交

**Q: 找不到Artifacts？**
- 确保workflow运行完成
- 检查是否在30天内（保留期）

**Q: GUI应用无法运行？**
- 确保在Windows系统上运行
- 检查系统依赖库

## 下一步

1. ✅ 提交代码到GitHub
2. ✅ 触发编译
3. ✅ 下载exe文件
4. ✅ 在Windows上测试运行

