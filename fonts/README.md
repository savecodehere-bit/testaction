# 中文字体说明

为了支持中文显示，现在**字体文件会直接嵌入到程序中**，确保在所有平台上都能正常显示中文。

## 使用方法

### 1. 添加字体文件到各服务目录

将中文字体文件复制到**每个服务目录下的 `fonts` 目录**：

```
client/fonts/
gateway_service/fonts/
user_service/fonts/
order_service/fonts/
center_service/fonts/
```

### 2. 支持的字体格式

- `.ttf` - TrueType Font
- `.ttc` - TrueType Collection
- `.otf` - OpenType Font

### 3. 推荐的字体文件名（按优先级）

- `chinese.ttf` 或 `chinese.ttc` - 通用中文字体
- `msyh.ttc` - 微软雅黑（Windows）
- `simsun.ttc` - 宋体（Windows）
- `font.ttf` 或 `font.ttc` - 通用字体名

### 4. 如何获取字体文件

#### Windows
从 Windows 系统复制字体文件：
```powershell
# 复制微软雅黑
Copy-Item "C:\Windows\Fonts\msyh.ttc" -Destination "client/fonts/msyh.ttc"
Copy-Item "C:\Windows\Fonts\msyh.ttc" -Destination "gateway_service/fonts/msyh.ttc"
Copy-Item "C:\Windows\Fonts\msyh.ttc" -Destination "user_service/fonts/msyh.ttc"
Copy-Item "C:\Windows\Fonts\msyh.ttc" -Destination "order_service/fonts/msyh.ttc"
Copy-Item "C:\Windows\Fonts\msyh.ttc" -Destination "center_service/fonts/msyh.ttc"
```

#### macOS
从 macOS 系统复制字体文件：
```bash
# 复制 PingFang
cp "/System/Library/Fonts/PingFang.ttc" client/fonts/chinese.ttc
cp "/System/Library/Fonts/PingFang.ttc" gateway_service/fonts/chinese.ttc
cp "/System/Library/Fonts/PingFang.ttc" user_service/fonts/chinese.ttc
cp "/System/Library/Fonts/PingFang.ttc" order_service/fonts/chinese.ttc
cp "/System/Library/Fonts/PingFang.ttc" center_service/fonts/chinese.ttc
```

#### Linux
使用开源字体（如文泉驿或 Noto Sans CJK）

### 5. 重新编译

添加字体文件后，重新编译所有服务：
```bash
go build -o bin/client.exe client/main.go
go build -o bin/gateway_service.exe gateway_service/main.go
go build -o bin/user_service.exe user_service/main.go
go build -o bin/order_service.exe order_service/main.go
go build -o bin/center_service.exe center_service/main.go
```

## 工作原理

1. **优先使用嵌入字体**：程序首先尝试加载嵌入的字体文件
2. **Fallback 到系统字体**：如果嵌入字体不存在，会自动使用系统字体
3. **跨平台兼容**：嵌入字体确保在所有平台上都能正常显示中文

## 注意事项

- 字体文件会被编译到可执行文件中，会增加文件大小（通常增加 5-20MB）
- 如果 `fonts` 目录下没有任何字体文件，编译会失败（因为 embed 需要至少一个文件）
- 如果不想嵌入字体，可以注释掉代码中的 `//go:embed fonts/*` 行，程序会使用系统字体

