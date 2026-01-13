# 中文字体说明

为了支持中文显示，需要将中文字体文件放在此目录。

## 推荐字体

### Windows
Windows 系统自带中文字体，可以使用：
- `C:\Windows\Fonts\msyh.ttc` (Microsoft YaHei)
- `C:\Windows\Fonts\simsun.ttc` (SimSun)

### macOS
macOS 系统自带中文字体：
- `/System/Library/Fonts/PingFang.ttc` (PingFang)
- `/Library/Fonts/Arial Unicode.ttf`

### Linux
可以使用开源字体：
- 文泉驿正黑体 (WenQuanYi Zen Hei)
- Noto Sans CJK

## 使用方法

1. 将字体文件复制到此目录，命名为 `chinese.ttf` 或 `chinese.ttc`
2. 重新编译服务

## 或者使用系统字体路径

如果不想嵌入字体，可以修改代码直接使用系统字体路径。

