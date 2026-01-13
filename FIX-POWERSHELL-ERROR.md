# 修复PowerShell语法错误

## 问题

GitHub Actions在Windows上默认使用PowerShell，但我们的命令是CMD语法，导致错误：
```
ParserError: Missing '(' after 'if' in if statement.
```

## 解决方案

需要在工作流文件中指定使用CMD shell。已修复的工作流文件内容如下：

## 修复后的工作流文件

文件位置：`.github/workflows/build-windows.yml`

关键修改：在每个需要CMD语法的步骤添加 `shell: cmd`

例如：
```yaml
- name: Create bin directory
  shell: cmd  # 添加这一行
  run: |
    if not exist bin mkdir bin
```

## 如何更新GitHub上的文件

### 方法1：直接在GitHub网页上编辑（推荐）

1. 打开：https://github.com/savecodehere-bit/testaction
2. 找到 `.github/workflows/build-windows.yml` 文件
3. 点击编辑按钮（铅笔图标）
4. 在每个需要CMD的步骤添加 `shell: cmd`
5. 点击 **Commit changes**

### 方法2：使用修复后的文件

我已经修复了本地文件，你可以：
1. 复制修复后的内容
2. 在GitHub网页上替换文件内容
3. 提交更改

## 需要添加 `shell: cmd` 的步骤

- Create bin directory
- Build center_service
- Build user_service
- Build order_service
- Build gateway_service
- Build client
- List built files

## 验证

修复后重新运行workflow，应该可以正常编译了。

