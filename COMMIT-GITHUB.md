# 提交GitHub Actions配置指南

## 步骤1：检查文件

确保以下文件已创建：
- ✅ `.github/workflows/build-windows.yml`
- ✅ `.github/workflows/build-all-platforms.yml`
- ✅ `.gitignore`

## 步骤2：提交到Git

```bash
# 1. 添加GitHub Actions文件
git add .github/

# 2. 添加其他配置文件（可选）
git add .gitignore
git add README-GITHUB.md
git add GITHUB-ACTIONS.md
git add QUICK-START-GITHUB.md

# 3. 提交
git commit -m "Add GitHub Actions for automatic Windows exe build"

# 4. 如果还没有远程仓库，先创建GitHub仓库，然后：
git remote add origin https://github.com/你的用户名/你的仓库名.git

# 5. 推送
git push -u origin main
```

## 步骤3：在GitHub上触发

1. 打开仓库页面：`https://github.com/你的用户名/你的仓库名`
2. 点击 **Actions** 标签
3. 如果看到 "Build Windows Executables" 工作流，说明配置成功
4. 点击 **Run workflow** → **Run workflow**

## 步骤4：等待编译完成

- 编译时间：约2-5分钟
- 可以在Actions页面实时查看日志

## 步骤5：下载exe文件

编译完成后：
1. 点击workflow运行记录
2. 滚动到底部
3. 在 **Artifacts** 部分下载 `windows-binaries.zip`

## 验证清单

- [ ] GitHub Actions文件已提交
- [ ] 代码已推送到GitHub
- [ ] 可以在Actions页面看到工作流
- [ ] 手动触发编译成功
- [ ] 可以下载编译好的exe文件

## 测试tag触发（可选）

```bash
# 创建tag
git tag v1.0.0

# 推送tag
git push origin v1.0.0
```

这会自动：
- 触发编译
- 创建Release
- 上传exe文件到Release

## 故障排查

如果编译失败：
1. 查看Actions页面的错误日志
2. 检查Go版本（需要1.21+）
3. 确保所有依赖都在go.mod中
4. 检查文件路径是否正确

