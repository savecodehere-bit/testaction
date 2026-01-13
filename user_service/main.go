package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"time"

	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// User 用户结构
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// UserService 用户服务
type UserService struct {
	users        map[int]*User
	mu           sync.RWMutex
	nextID       int
	port         int
	registryURL  string
	logContainer *fyne.Container
	logScroll    *container.Scroll
	statusLabel  *widget.Label
}

// NewUserService 创建新的用户服务
func NewUserService(port int, registryURL string, logContainer *fyne.Container, logScroll *container.Scroll, statusLabel *widget.Label) *UserService {
	us := &UserService{
		users:        make(map[int]*User),
		nextID:       1,
		port:         port,
		registryURL:  registryURL,
		logContainer: logContainer,
		logScroll:    logScroll,
		statusLabel:  statusLabel,
	}
	// 初始化一些示例数据
	us.users[1] = &User{ID: 1, Name: "张三", Email: "zhangsan@example.com"}
	us.users[2] = &User{ID: 2, Name: "李四", Email: "lisi@example.com"}
	us.nextID = 3
	return us
}

// logMessage 添加日志消息（彩色显示）
func (us *UserService) logMessage(msg string) {
	if us.logContainer != nil {
		timestamp := time.Now().Format("15:04:05")
		fullMsg := fmt.Sprintf("[%s] %s", timestamp, msg)

		// 根据消息内容确定颜色
		var msgColor color.Color
		if contains(msg, "注册") || contains(msg, "成功") {
			msgColor = color.NRGBA{R: 0, G: 200, B: 0, A: 255} // 绿色
		} else if contains(msg, "警告") || contains(msg, "失败") {
			msgColor = color.NRGBA{R: 255, G: 165, B: 0, A: 255} // 橙色
		} else if contains(msg, "错误") || contains(msg, "异常") {
			msgColor = color.NRGBA{R: 255, G: 0, B: 0, A: 255} // 红色
		} else if contains(msg, "启动") || contains(msg, "就绪") {
			msgColor = color.NRGBA{R: 0, G: 150, B: 255, A: 255} // 蓝色
		} else {
			msgColor = color.NRGBA{R: 200, G: 200, B: 200, A: 255} // 灰色（默认）
		}

		// 创建带颜色的文本
		logText := canvas.NewText(fullMsg, msgColor)
		logText.TextStyle = fyne.TextStyle{Monospace: true}
		logText.Alignment = fyne.TextAlignLeading

		// 添加到容器
		us.logContainer.Add(logText)

		// 限制日志条目数量（保留最后200条）
		if len(us.logContainer.Objects) > 200 {
			oldObjs := us.logContainer.Objects
			us.logContainer.Objects = oldObjs[len(oldObjs)-200:]
			us.logContainer.Refresh()
		}

		// 滚动到底部
		if us.logScroll != nil {
			us.logScroll.ScrollToBottom()
		}
	}
}

// contains 检查字符串是否包含子串
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// updateStatus 更新状态
func (us *UserService) updateStatus() {
	if us.statusLabel != nil {
		us.mu.RLock()
		userCount := len(us.users)
		us.mu.RUnlock()
		us.statusLabel.SetText(fmt.Sprintf("状态: 运行中 | 端口: %d | 用户数: %d | 注册中心: %s", us.port, userCount, us.registryURL))
	}
}

// RegisterToRegistry 注册到服务注册中心
func (us *UserService) RegisterToRegistry() {
	if us.registryURL == "" {
		return
	}

	serviceInfo := map[string]interface{}{
		"name":    "user-service",
		"address": "localhost",
		"port":    us.port,
	}

	jsonData, _ := json.Marshal(serviceInfo)
	resp, err := http.Post(us.registryURL+"/register", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		us.logMessage(fmt.Sprintf("警告: 无法注册到服务注册中心: %v", err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		us.logMessage("✓ 已注册到服务注册中心")
		// 启动心跳协程
		go us.startHeartbeat()
	}
}

// startHeartbeat 启动心跳（每5秒发送一次）
func (us *UserService) startHeartbeat() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if us.registryURL != "" {
			http.Post(us.registryURL+"/heartbeat?name=user-service", "application/json", nil)
		}
	}
}

// UnregisterFromRegistry 从服务注册中心注销（同步等待完成）
func (us *UserService) UnregisterFromRegistry() {
	if us.registryURL == "" {
		return
	}

	us.logMessage("正在注销服务...")

	// 创建带超时的HTTP客户端
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Post(us.registryURL+"/unregister?name=user-service", "application/json", nil)
	if err != nil {
		us.logMessage(fmt.Sprintf("注销失败: %v", err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		us.logMessage("✓ 已从服务注册中心注销")
	} else {
		us.logMessage(fmt.Sprintf("注销失败，状态码: %d", resp.StatusCode))
	}
}

// GetUser 获取用户信息
func (us *UserService) GetUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "Missing id parameter", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid id parameter", http.StatusBadRequest)
		return
	}

	us.mu.RLock()
	user, exists := us.users[id]
	us.mu.RUnlock()

	if !exists {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	us.logMessage(fmt.Sprintf("GET /user?id=%d - 返回用户信息", id))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// ListUsers 列出所有用户
func (us *UserService) ListUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	us.mu.RLock()
	users := make([]*User, 0, len(us.users))
	for _, user := range us.users {
		users = append(users, user)
	}
	us.mu.RUnlock()

	us.logMessage("GET /user - 返回所有用户列表")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// CreateUser 创建用户
func (us *UserService) CreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	us.mu.Lock()
	user.ID = us.nextID
	us.nextID++
	us.users[user.ID] = &user
	us.mu.Unlock()

	us.logMessage(fmt.Sprintf("POST /user - 创建新用户: %s (ID: %d)", user.Name, user.ID))
	us.updateStatus()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func main() {
	// 创建GUI应用
	myApp := app.New()
	// 设置支持中文的主题（使用系统默认字体，支持中文）
	myApp.Settings().SetTheme(newChineseTheme())
	myWindow := myApp.NewWindow("用户服务 (端口: 8081)")
	myWindow.Resize(fyne.NewSize(800, 600))

	// 创建日志显示区域（使用canvas.Text支持彩色显示）
	logContainer := container.NewVBox()
	// 添加初始消息
	initText := canvas.NewText("用户服务启动中...", color.NRGBA{R: 200, G: 200, B: 200, A: 255})
	initText.TextStyle = fyne.TextStyle{Monospace: true}
	logContainer.Add(initText)
	logScroll := container.NewScroll(logContainer)
	logScroll.SetMinSize(fyne.NewSize(0, 0))

	// 创建状态标签
	statusLabel := widget.NewLabel("状态: 启动中...")

	port := 8081
	registryURL := "http://localhost:8080"

	service := NewUserService(port, registryURL, logContainer, logScroll, statusLabel)

	http.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if r.URL.Query().Get("id") != "" {
				service.GetUser(w, r)
			} else {
				service.ListUsers(w, r)
			}
		case http.MethodPost:
			service.CreateUser(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// 启动HTTP服务器
	go func() {
		service.logMessage(fmt.Sprintf("用户服务启动在端口 %d", port))
		service.logMessage(fmt.Sprintf("服务注册中心: %s", registryURL))
		service.logMessage("API端点:")
		service.logMessage(fmt.Sprintf("  GET  http://localhost:%d/user?id=1 - 获取用户", port))
		service.logMessage(fmt.Sprintf("  GET  http://localhost:%d/user - 列出所有用户", port))
		service.logMessage(fmt.Sprintf("  POST http://localhost:%d/user - 创建用户", port))

		// 注册到服务注册中心
		service.RegisterToRegistry()
		service.updateStatus()

		// 定期更新状态
		go func() {
			ticker := time.NewTicker(2 * time.Second)
			defer ticker.Stop()
			for range ticker.C {
				service.updateStatus()
			}
		}()

		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
	}()

	// 创建UI布局
	content := container.NewBorder(
		statusLabel,
		nil,
		nil,
		nil,
		container.NewBorder(
			widget.NewLabel("日志输出"), nil, nil, nil,
			logScroll,
		),
	)

	myWindow.SetContent(content)

	// 设置窗口关闭拦截，在关闭前注销服务
	myWindow.SetCloseIntercept(func() {
		service.logMessage("窗口关闭，正在注销服务...")
		service.UnregisterFromRegistry()
		// 稍微等待一下，确保注销请求完成
		time.Sleep(300 * time.Millisecond)
		// 手动关闭窗口
		myWindow.Close()
	})

	myWindow.ShowAndRun()
}

// chineseTheme 支持中文的主题
type chineseTheme struct {
	baseTheme   fyne.Theme
	chineseFont fyne.Resource
}

func newChineseTheme() *chineseTheme {
	t := &chineseTheme{
		baseTheme: theme.DefaultTheme(),
	}

	// 尝试加载系统字体
	t.chineseFont = loadSystemChineseFont()
	if t.chineseFont == nil {
		// 如果找不到系统字体，使用默认字体
		t.chineseFont = t.baseTheme.Font(fyne.TextStyle{})
	}

	return t
}

// loadSystemChineseFont 加载系统中文字体
func loadSystemChineseFont() fyne.Resource {
	var fontPaths []string

	switch runtime.GOOS {
	case "windows":
		windir := os.Getenv("WINDIR")
		if windir == "" {
			windir = "C:\\Windows"
		}
		fontPaths = []string{
			filepath.Join(windir, "Fonts", "msyh.ttc"),   // Microsoft YaHei
			filepath.Join(windir, "Fonts", "msyhbd.ttc"), // Microsoft YaHei Bold
			filepath.Join(windir, "Fonts", "simsun.ttc"), // SimSun
			filepath.Join(windir, "Fonts", "simhei.ttf"), // SimHei
		}
	case "darwin": // macOS
		fontPaths = []string{
			"/System/Library/Fonts/PingFang.ttc",
			"/Library/Fonts/Arial Unicode.ttf",
			"/System/Library/Fonts/STHeiti Light.ttc",
		}
	case "linux":
		fontPaths = []string{
			"/usr/share/fonts/truetype/wqy/wqy-zenhei.ttc",
			"/usr/share/fonts/truetype/noto/NotoSansCJK-Regular.ttc",
			"/usr/share/fonts/truetype/droid/DroidSansFallbackFull.ttf",
		}
	}

	// 尝试加载第一个存在的字体文件
	for _, path := range fontPaths {
		if _, err := os.Stat(path); err == nil {
			uri := storage.NewFileURI(path)
			res, err := storage.LoadResourceFromURI(uri)
			if err == nil {
				return res
			}
		}
	}

	return nil
}

func (t *chineseTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	return t.baseTheme.Color(name, variant)
}

func (t *chineseTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return t.baseTheme.Icon(name)
}

func (t *chineseTheme) Font(style fyne.TextStyle) fyne.Resource {
	// 使用加载的中文字体
	return t.chineseFont
}

func (t *chineseTheme) Size(name fyne.ThemeSizeName) float32 {
	return t.baseTheme.Size(name)
}
