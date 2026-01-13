# 修复Windows中文显示乱码问题

## 问题描述

在Windows上运行时：
- ✅ 窗口标题中文正常显示
- ❌ GUI界面内中文显示乱码

## 原因分析

1. **字体加载失败**：Fyne可能无法正确加载Windows系统字体
2. **字体路径问题**：Windows字体文件路径可能不正确
3. **字体资源验证不足**：没有验证加载的字体资源是否有效

## 已修复的内容

### 1. 扩展字体列表

增加了更多Windows中文字体选项：
- `msyh.ttc` - 微软雅黑（优先）
- `msyhbd.ttc` - 微软雅黑粗体
- `msyhl.ttc` - 微软雅黑细体
- `simsun.ttc` - 宋体
- `simhei.ttf` - 黑体
- `simkai.ttf` - 楷体
- `simli.ttf` - 隶书
- `msjh.ttc` - 微软正黑体
- `mingliu.ttc` - 新细明体

### 2. 改进字体加载验证

```go
// 验证字体资源是否有效
if res.Content() != nil && len(res.Content()) > 0 {
    return res
}
```

### 3. 改进Fallback机制

如果系统字体加载失败，使用Fyne默认主题字体（Windows上通常支持中文）。

## 测试方法

### 方法1：检查字体文件是否存在

在Windows PowerShell中运行：
```powershell
Test-Path "C:\Windows\Fonts\msyh.ttc"
Test-Path "C:\Windows\Fonts\simsun.ttc"
```

### 方法2：重新编译测试

```bash
# 重新编译
go build -o bin/center_service.exe center_service/main.go

# 运行测试
bin/center_service.exe
```

### 方法3：添加调试日志（可选）

如果需要调试，可以在`loadSystemChineseFont`函数中添加日志：

```go
fmt.Printf("尝试加载字体: %s\n", path)
if res != nil {
    fmt.Printf("字体加载成功: %s\n", path)
}
```

## 如果仍然乱码

### 检查1：系统区域设置

1. 打开"控制面板" → "区域"
2. 确保"格式"设置为"中文(简体，中国)"
3. 在"管理"选项卡，点击"更改系统区域设置"
4. 设置为"中文(简体，中国)"
5. 重启计算机

### 检查2：安装中文字体

确保Windows系统安装了中文字体：
- 微软雅黑（Microsoft YaHei）
- 宋体（SimSun）

### 检查3：使用Fyne内置字体

如果系统字体都不行，可以考虑：
1. 将中文字体文件嵌入到程序中
2. 使用Fyne的资源系统加载字体

## 已更新的文件

- ✅ `center_service/main.go`
- ✅ `user_service/main.go`
- ✅ `order_service/main.go`
- ✅ `gateway_service/main.go`
- ✅ `client/main.go`

## 下一步

1. 重新编译所有服务
2. 在Windows上测试运行
3. 检查中文显示是否正常

如果还有问题，可能需要考虑使用embed嵌入字体文件。

