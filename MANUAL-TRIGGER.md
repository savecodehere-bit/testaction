# 手动触发GitHub Actions编译

## 问题：tag推送后没有自动编译

可能的原因：
1. tag推送时工作流文件还没在仓库中
2. GitHub Actions需要手动触发一次

## 解决方案：手动触发

### 步骤1：打开Actions页面

访问：https://github.com/savecodehere-bit/testaction/actions

### 步骤2：选择工作流

点击左侧的 **"Build Windows Executables"**

### 步骤3：手动触发

1. 点击右侧的 **"Run workflow"** 按钮
2. 选择分支：`main`
3. 点击绿色的 **"Run workflow"** 按钮

### 步骤4：等待编译

- 编译时间：约2-5分钟
- 可以在页面实时查看编译日志

### 步骤5：下载exe

编译完成后：
1. 点击运行记录
2. 滚动到底部
3. 在 **Artifacts** 部分下载 `windows-binaries.zip`

## 或者：重新推送tag

如果手动触发也不行，可以删除tag重新推送：

```bash
# 删除本地tag
git tag -d v1.0.0

# 删除远程tag
git push origin :refs/tags/v1.0.0

# 重新创建tag
git tag v1.0.0

# 推送tag（这次应该会触发）
git push origin v1.0.0
```

## 检查工作流是否启用

1. 打开仓库设置：https://github.com/savecodehere-bit/testaction/settings
2. 点击左侧 **Actions** → **General**
3. 确保 **"Allow all actions and reusable workflows"** 已启用

