# GitHub Actions 自动编译指南

## 概述

本项目配置了GitHub Actions工作流，可以自动编译Windows exe文件（以及其他平台）。

## 工作流文件

### 1. build-windows.yml - Windows专用编译

**触发方式：**
- 手动触发（GitHub Actions页面）
- 推送tag时自动触发（如 `v1.0.0`）
- 推送到main/master分支时自动触发（当Go代码变更时）

**功能：**
- 在Windows环境中编译所有服务
- 生成 `.exe` 文件
- 上传构建产物
- 如果推送tag，自动创建Release

### 2. build-all-platforms.yml - 跨平台编译

**触发方式：**
- 手动触发
- 推送tag时自动触发

**功能：**
- 同时编译Windows、Linux、macOS版本
- 生成所有平台的可执行文件

## 使用方法

### 方法1：手动触发编译

1. 打开GitHub仓库页面
2. 点击 **Actions** 标签
3. 选择 **Build Windows Executables** 工作流
4. 点击 **Run workflow**
5. 选择分支，点击 **Run workflow** 按钮

### 方法2：推送tag自动编译

```bash
# 创建并推送tag
git tag v1.0.0
git push origin v1.0.0
```

推送tag后，GitHub Actions会自动：
1. 编译所有服务
2. 创建Release
3. 上传exe文件到Release

### 方法3：代码变更自动编译

当推送到main/master分支，且Go代码有变更时，会自动触发编译。

## 下载编译好的文件

### 从Artifacts下载

1. 打开GitHub Actions页面
2. 点击最新的workflow运行
3. 滚动到底部，找到 **Artifacts** 部分
4. 点击 **windows-binaries** 下载zip文件

### 从Release下载

如果通过tag触发，文件会上传到Release：

1. 打开仓库的 **Releases** 页面
2. 找到对应的版本
3. 下载附件中的exe文件

## 工作流配置说明

### Windows编译工作流

```yaml
on:
  workflow_dispatch:     # 手动触发
  push:
    tags: ['v*']        # tag触发
    branches: ['main']  # 分支触发
```

### 编译步骤

1. **Checkout code** - 检出代码
2. **Set up Go** - 安装Go 1.21
3. **Create bin directory** - 创建输出目录
4. **Download dependencies** - 下载依赖
5. **Build services** - 编译各个服务
6. **Upload artifacts** - 上传构建产物

## 编译的服务

- `center_service.exe` - 服务注册中心
- `user_service.exe` - 用户服务
- `order_service.exe` - 订单服务
- `gateway_service.exe` - API网关服务
- `client.exe` - 测试客户端

## 注意事项

### 1. GUI应用编译限制

由于使用了Fyne GUI框架，需要CGO支持。Windows环境下的编译：
- ✅ **Windows runner**: 可以正常编译GUI应用
- ⚠️ **Linux/macOS runner**: 交叉编译Windows GUI应用可能失败

### 2. 依赖下载

工作流会自动运行 `go mod download`，确保所有依赖都已下载。

### 3. 构建时间

- Windows编译：约2-5分钟
- 跨平台编译：约5-10分钟（并行编译）

### 4. 文件大小

编译后的exe文件大小：
- 每个服务约：10-20MB（包含GUI框架）
- 总共约：50-100MB

## 故障排查

### 编译失败

1. **检查Go版本**：确保使用Go 1.21+
2. **检查依赖**：确保 `go.mod` 和 `go.sum` 已提交
3. **查看日志**：在Actions页面查看详细错误信息

### 找不到Artifacts

1. 确保工作流运行完成（不是被取消）
2. 检查是否在7天内（Artifacts默认保留7天，已设置为30天）

### GUI应用无法运行

1. 确保在Windows系统上运行
2. 检查是否有必要的系统库
3. 查看错误日志

## 自定义配置

### 修改Go版本

编辑 `.github/workflows/build-windows.yml`：

```yaml
- name: Set up Go
  uses: actions/setup-go@v4
  with:
    go-version: '1.22'  # 修改版本号
```

### 添加编译参数

在build步骤中添加参数：

```yaml
- name: Build center_service
  run: |
    cd center_service
    go build -ldflags="-s -w" -o ..\bin\center_service.exe main.go
```

### 修改触发条件

编辑 `on:` 部分：

```yaml
on:
  workflow_dispatch:
  push:
    branches:
      - main
      - develop  # 添加其他分支
```

## 最佳实践

1. **使用tag触发Release**：重要版本使用tag，自动创建Release
2. **定期清理Artifacts**：避免占用过多存储空间
3. **测试编译产物**：下载后在实际Windows环境中测试
4. **版本管理**：使用语义化版本（如 v1.0.0）

## 示例工作流

### 创建Release并上传exe

```bash
# 1. 创建tag
git tag -a v1.0.0 -m "Release version 1.0.0"

# 2. 推送tag
git push origin v1.0.0

# 3. GitHub Actions自动：
#    - 编译所有服务
#    - 创建Release
#    - 上传exe文件
```

## 相关文件

- `.github/workflows/build-windows.yml` - Windows编译工作流
- `.github/workflows/build-all-platforms.yml` - 跨平台编译工作流
- `build.bat` - 本地Windows编译脚本

