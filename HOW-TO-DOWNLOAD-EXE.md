# 如何下载编译好的exe文件

## 方法1：从GitHub Actions Artifacts下载（推荐）

### 步骤1：打开Actions页面
访问：https://github.com/savecodehere-bit/testaction/actions

### 步骤2：找到最新的workflow运行
- 找到最新的运行记录（最上面那个）
- 绿色✓ = 编译成功
- 黄色圆圈 = 正在编译
- 红色✗ = 编译失败

### 步骤3：下载Artifacts
1. 点击最新的运行记录
2. 滚动到页面**最底部**
3. 找到 **Artifacts** 部分
4. 点击 **windows-binaries**（蓝色链接）
5. 会自动下载一个zip文件

### 步骤4：解压文件
解压zip文件后，你会得到：
- `center_service.exe`
- `user_service.exe`
- `order_service.exe`
- `gateway_service.exe`
- `client.exe`

## 方法2：从Release下载（如果推送了tag）

### 步骤1：打开Releases页面
访问：https://github.com/savecodehere-bit/testaction/releases

### 步骤2：找到对应版本
- 如果推送了tag（如v1.0.0），会看到对应的Release
- 点击版本号

### 步骤3：下载文件
在 **Assets** 部分，下载exe文件

## 快速链接

- **Actions页面**：https://github.com/savecodehere-bit/testaction/actions
- **Releases页面**：https://github.com/savecodehere-bit/testaction/releases

## 找不到Artifacts？

### 检查编译是否完成
- 确保workflow显示绿色✓（成功）
- 如果还在运行（黄色圆圈），需要等待

### 检查是否在页面底部
- Artifacts在页面**最底部**，需要滚动到底

### 检查编译是否成功
- 如果显示红色✗，说明编译失败
- 点击查看错误日志

## 文件位置总结

编译完成后，exe文件在：
1. **GitHub Actions Artifacts**（每次编译都有）
2. **GitHub Release**（只有推送tag时才有）

**推荐使用方法1**，因为每次编译都会生成Artifacts。

