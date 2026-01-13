package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"image/color"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// 嵌入字体文件
// 使用方法：将中文字体文件（如 msyh.ttc 或 simsun.ttc）复制到 fonts 目录下
// 支持的格式：.ttf, .ttc, .otf
// 如果没有字体文件，程序会自动使用系统字体作为fallback
//go:embed fonts/*
var embeddedFonts embed.FS

// Order 订单结构
type Order struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Amount    float64   `json:"amount"`
	Status    string    `json:"status"`
	Items     []string  `json:"items"`
	CreatedAt time.Time `json:"created_at"`
}

// ServiceInfo 服务信息（从注册中心获取）
type ServiceInfo struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Port    int    `json:"port"`
	URL     string `json:"url"`
}

// OrderService 订单服务
type OrderService struct {
	orders         map[int]*Order
	mu             sync.RWMutex
	nextID         int
	port           int
	registryURL    string
	userServiceURL string
	muURL          sync.RWMutex
	logContainer   *fyne.Container
	logScroll      *container.Scroll
	statusLabel    *widget.Label
}

// NewOrderService 创建新的订单服务
func NewOrderService(port int, registryURL string, logContainer *fyne.Container, logScroll *container.Scroll, statusLabel *widget.Label) *OrderService {
	os := &OrderService{
		orders:       make(map[int]*Order),
		nextID:       1,
		port:         port,
		registryURL:  registryURL,
		logContainer: logContainer,
		logScroll:    logScroll,
		statusLabel:  statusLabel,
	}
	// 初始化一些示例数据
	os.orders[1] = &Order{
		ID:        1,
		UserID:    1,
		Amount:    99.99,
		Status:    "已完成",
		Items:     []string{"商品A", "商品B"},
		CreatedAt: time.Now().Add(-24 * time.Hour),
	}
	os.orders[2] = &Order{
		ID:        2,
		UserID:    2,
		Amount:    199.99,
		Status:    "待发货",
		Items:     []string{"商品C"},
		CreatedAt: time.Now().Add(-12 * time.Hour),
	}
	os.nextID = 3
	return os
}

// logMessage 添加日志消息（彩色显示）
func (os *OrderService) logMessage(msg string) {
	if os.logContainer != nil {
		timestamp := time.Now().Format("15:04:05")
		fullMsg := fmt.Sprintf("[%s] %s", timestamp, msg)

		// 根据消息内容确定颜色
		var msgColor color.Color
		if contains(msg, "注册") || contains(msg, "成功") || contains(msg, "发现") {
			msgColor = color.NRGBA{R: 0, G: 200, B: 0, A: 255} // 绿色
		} else if contains(msg, "警告") || contains(msg, "失败") {
			msgColor = color.NRGBA{R: 255, G: 165, B: 0, A: 255} // 橙色
		} else if contains(msg, "错误") || contains(msg, "异常") || contains(msg, "无法") {
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
		os.logContainer.Add(logText)

		// 限制日志条目数量（保留最后200条）
		if len(os.logContainer.Objects) > 200 {
			oldObjs := os.logContainer.Objects
			os.logContainer.Objects = oldObjs[len(oldObjs)-200:]
			os.logContainer.Refresh()
		}

		// 滚动到底部
		if os.logScroll != nil {
			os.logScroll.ScrollToBottom()
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
func (os *OrderService) updateStatus() {
	if os.statusLabel != nil {
		os.mu.RLock()
		orderCount := len(os.orders)
		os.mu.RUnlock()
		os.muURL.RLock()
		userServiceURL := os.userServiceURL
		os.muURL.RUnlock()
		if userServiceURL == "" {
			userServiceURL = "未发现"
		}
		os.statusLabel.SetText(fmt.Sprintf("状态: 运行中 | 端口: %d | 订单数: %d | 用户服务: %s", os.port, orderCount, userServiceURL))
	}
}

// RegisterToRegistry 注册到服务注册中心
func (os *OrderService) RegisterToRegistry() {
	if os.registryURL == "" {
		return
	}

	serviceInfo := map[string]interface{}{
		"name":    "order-service",
		"address": "localhost",
		"port":    os.port,
	}

	jsonData, _ := json.Marshal(serviceInfo)
	resp, err := http.Post(os.registryURL+"/register", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		os.logMessage(fmt.Sprintf("警告: 无法注册到服务注册中心: %v", err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		os.logMessage("✓ 已注册到服务注册中心")
		// 启动心跳协程
		go os.startHeartbeat()
	}
}

// startHeartbeat 启动心跳（每5秒发送一次）
func (os *OrderService) startHeartbeat() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if os.registryURL != "" {
			http.Post(os.registryURL+"/heartbeat?name=order-service", "application/json", nil)
		}
	}
}

// UnregisterFromRegistry 从服务注册中心注销（同步等待完成）
func (os *OrderService) UnregisterFromRegistry() {
	if os.registryURL == "" {
		return
	}

	os.logMessage("正在注销服务...")

	// 创建带超时的HTTP客户端
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Post(os.registryURL+"/unregister?name=order-service", "application/json", nil)
	if err != nil {
		os.logMessage(fmt.Sprintf("注销失败: %v", err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		os.logMessage("✓ 已从服务注册中心注销")
	} else {
		os.logMessage(fmt.Sprintf("注销失败，状态码: %d", resp.StatusCode))
	}
}

// DiscoverUserService 从注册中心发现用户服务
func (os *OrderService) DiscoverUserService() {
	if os.registryURL == "" {
		return
	}

	resp, err := http.Get(os.registryURL + "/discover?name=user-service")
	if err != nil {
		os.logMessage(fmt.Sprintf("警告: 无法从注册中心发现用户服务: %v", err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var service ServiceInfo
		if err := json.NewDecoder(resp.Body).Decode(&service); err == nil {
			os.muURL.Lock()
			os.userServiceURL = service.URL
			os.muURL.Unlock()
			os.logMessage(fmt.Sprintf("✓ 发现用户服务: %s", service.URL))
			os.updateStatus()
		}
	}
}

// getUserServiceURL 获取用户服务URL（带重试发现）
func (os *OrderService) getUserServiceURL() string {
	os.muURL.RLock()
	url := os.userServiceURL
	os.muURL.RUnlock()

	// 如果URL为空，尝试重新发现
	if url == "" {
		os.DiscoverUserService()
		os.muURL.RLock()
		url = os.userServiceURL
		os.muURL.RUnlock()
	}

	return url
}

// GetOrder 获取订单信息
func (os *OrderService) GetOrder(w http.ResponseWriter, r *http.Request) {
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

	os.mu.RLock()
	order, exists := os.orders[id]
	os.mu.RUnlock()

	if !exists {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	os.logMessage(fmt.Sprintf("GET /order?id=%d - 返回订单信息", id))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

// ListOrders 获取所有订单列表
func (os *OrderService) ListOrders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	os.mu.RLock()
	orders := make([]*Order, 0, len(os.orders))
	for _, order := range os.orders {
		orders = append(orders, order)
	}
	os.mu.RUnlock()

	os.logMessage(fmt.Sprintf("GET /order - 返回所有订单列表（共 %d 条）", len(orders)))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

// GetOrdersByUser 根据用户ID获取订单列表
func (os *OrderService) GetOrdersByUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		http.Error(w, "Missing user_id parameter", http.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user_id parameter", http.StatusBadRequest)
		return
	}

	os.mu.RLock()
	orders := make([]*Order, 0)
	for _, order := range os.orders {
		if order.UserID == userID {
			orders = append(orders, order)
		}
	}
	os.mu.RUnlock()

	os.logMessage(fmt.Sprintf("GET /order?user_id=%d - 返回用户订单列表", userID))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

// CreateOrder 创建订单（会调用用户服务验证用户是否存在）
func (os *OrderService) CreateOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var order Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 通过服务发现获取用户服务URL并验证用户是否存在
	userServiceURL := os.getUserServiceURL()
	if userServiceURL != "" {
		resp, err := http.Get(fmt.Sprintf("%s/user?id=%d", userServiceURL, order.UserID))
		if err != nil || resp.StatusCode != http.StatusOK {
			os.logMessage(fmt.Sprintf("创建订单失败: 用户 %d 不存在", order.UserID))
			http.Error(w, "User not found", http.StatusBadRequest)
			return
		}
		resp.Body.Close()
		os.logMessage(fmt.Sprintf("通过服务发现验证用户 %d 存在", order.UserID))
	}

	os.mu.Lock()
	order.ID = os.nextID
	os.nextID++
	order.CreatedAt = time.Now()
	if order.Status == "" {
		order.Status = "待支付"
	}
	os.orders[order.ID] = &order
	os.mu.Unlock()

	os.logMessage(fmt.Sprintf("POST /order - 创建新订单: ID=%d, 用户=%d, 金额=%.2f", order.ID, order.UserID, order.Amount))
	os.updateStatus()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}

// GetOrderWithUserInfo 获取订单信息（包含用户信息，演示服务间调用）
func (os *OrderService) GetOrderWithUserInfo(w http.ResponseWriter, r *http.Request) {
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

	os.mu.RLock()
	order, exists := os.orders[id]
	os.mu.RUnlock()

	if !exists {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	// 通过服务发现调用用户服务获取用户信息
	type OrderWithUser struct {
		Order *Order      `json:"order"`
		User  interface{} `json:"user"`
	}

	result := OrderWithUser{Order: order}

	userServiceURL := os.getUserServiceURL()
	if userServiceURL != "" {
		os.logMessage(fmt.Sprintf("通过服务发现调用用户服务: %s/user?id=%d", userServiceURL, order.UserID))
		resp, err := http.Get(fmt.Sprintf("%s/user?id=%d", userServiceURL, order.UserID))
		if err == nil && resp.StatusCode == http.StatusOK {
			var user interface{}
			json.NewDecoder(resp.Body).Decode(&user)
			result.User = user
			resp.Body.Close()
			os.logMessage("✓ 成功获取用户信息")
		}
	}

	os.logMessage(fmt.Sprintf("GET /order/with-user?id=%d - 返回订单和用户信息", id))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func main() {
	// 创建GUI应用
	myApp := app.New()
	// 设置支持中文的主题（使用系统默认字体，支持中文）
	myApp.Settings().SetTheme(newChineseTheme())
	myWindow := myApp.NewWindow("订单服务 (端口: 8082)")
	myWindow.Resize(fyne.NewSize(800, 600))

	// 创建日志显示区域（使用canvas.Text支持彩色显示）
	logContainer := container.NewVBox()
	// 添加初始消息
	initText := canvas.NewText("订单服务启动中...", color.NRGBA{R: 200, G: 200, B: 200, A: 255})
	initText.TextStyle = fyne.TextStyle{Monospace: true}
	logContainer.Add(initText)
	logScroll := container.NewScroll(logContainer)
	logScroll.SetMinSize(fyne.NewSize(0, 0))

	// 创建状态标签
	statusLabel := widget.NewLabel("状态: 启动中...")

	port := 8082
	registryURL := "http://localhost:8080"

	service := NewOrderService(port, registryURL, logContainer, logScroll, statusLabel)

	http.HandleFunc("/order", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if r.URL.Query().Get("id") != "" {
				service.GetOrder(w, r)
			} else if r.URL.Query().Get("user_id") != "" {
				service.GetOrdersByUser(w, r)
			} else {
				// 没有参数时，返回所有订单
				service.ListOrders(w, r)
			}
		case http.MethodPost:
			service.CreateOrder(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/order/with-user", service.GetOrderWithUserInfo)

	// 启动HTTP服务器
	go func() {
		service.logMessage(fmt.Sprintf("订单服务启动在端口 %d", port))
		service.logMessage(fmt.Sprintf("服务注册中心: %s", registryURL))

		// 注册到服务注册中心
		service.RegisterToRegistry()

		// 从注册中心发现用户服务
		service.DiscoverUserService()

		// 如果用户服务未发现，定期尝试发现（不发送实际请求）
		go func() {
			ticker := time.NewTicker(2 * time.Second)
			defer ticker.Stop()
			for range ticker.C {
				service.muURL.RLock()
				url := service.userServiceURL
				service.muURL.RUnlock()

				if url == "" {
					// 如果还没发现，尝试发现
					service.DiscoverUserService()
				}
				// 不再主动检查服务可用性，只在实际使用时发现服务不可用再重新发现
			}
		}()

		service.logMessage("API端点:")
		service.logMessage(fmt.Sprintf("  GET  http://localhost:%d/order?id=1 - 获取订单", port))
		service.logMessage(fmt.Sprintf("  GET  http://localhost:%d/order?user_id=1 - 获取用户的订单列表", port))
		service.logMessage(fmt.Sprintf("  GET  http://localhost:%d/order/with-user?id=1 - 获取订单（包含用户信息，演示服务间调用）", port))
		service.logMessage(fmt.Sprintf("  POST http://localhost:%d/order - 创建订单（会验证用户是否存在）", port))

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
		// 如果找不到系统字体，在Windows上直接使用默认主题字体
		// Fyne在Windows上会自动使用系统默认字体，通常支持中文
		t.chineseFont = t.baseTheme.Font(fyne.TextStyle{})
		if t.chineseFont == nil {
			t.chineseFont = theme.DefaultTheme().Font(fyne.TextStyle{})
		}
	}

	return t
}

// loadSystemChineseFont 加载中文字体（优先使用嵌入的字体，否则使用系统字体）
func loadSystemChineseFont() fyne.Resource {
	// 1. 优先尝试加载嵌入的字体文件
	if embeddedFont := loadEmbeddedFont(); embeddedFont != nil {
		return embeddedFont
	}

	// 2. 如果嵌入字体不存在，尝试加载系统字体
	var fontPaths []string

	switch runtime.GOOS {
	case "windows":
		// Windows字体路径，按优先级排序
		// 优先使用常见的Windows中文字体
		windir := os.Getenv("WINDIR")
		if windir == "" {
			windir = "C:\\Windows"
		}
		
		// 尝试多个可能的字体路径
		fontPaths = []string{
			filepath.Join(windir, "Fonts", "msyh.ttc"),    // Microsoft YaHei (微软雅黑) - 最常见
			filepath.Join(windir, "Fonts", "simsun.ttc"),  // SimSun (宋体) - 传统字体
			filepath.Join(windir, "Fonts", "simhei.ttf"),  // SimHei (黑体)
			filepath.Join(windir, "Fonts", "msyhbd.ttc"),  // Microsoft YaHei Bold
			filepath.Join(windir, "Fonts", "msyhl.ttc"),   // Microsoft YaHei Light
			filepath.Join(windir, "Fonts", "simkai.ttf"),  // SimKai (楷体)
			filepath.Join(windir, "Fonts", "simli.ttf"),   // SimLi (隶书)
			filepath.Join(windir, "Fonts", "msjh.ttc"),    // Microsoft JhengHei (微软正黑体)
			filepath.Join(windir, "Fonts", "mingliu.ttc"), // MingLiU (新细明体)
		}
		
		// 也尝试使用绝对路径（某些情况下可能更可靠）
		if windir != "C:\\Windows" {
			fontPaths = append(fontPaths,
				filepath.Join("C:\\Windows", "Fonts", "msyh.ttc"),
				filepath.Join("C:\\Windows", "Fonts", "simsun.ttc"),
			)
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
		// 检查文件是否存在
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			// 尝试使用file://协议加载字体
			uri := storage.NewFileURI(path)
			res, err := storage.LoadResourceFromURI(uri)
			if err == nil && res != nil {
				// 验证字体资源是否有效
				content := res.Content()
				if len(content) > 0 {
					// 字体加载成功
					return res
				}
			}
			
			// 如果storage.LoadResourceFromURI失败，尝试直接读取文件
			if data, err := os.ReadFile(path); err == nil && len(data) > 0 {
				// 创建内存资源
				res := fyne.NewStaticResource(filepath.Base(path), data)
				return res
			}
		}
	}

	// 如果所有字体都加载失败，返回nil（会使用默认字体）
	return nil
}

// loadEmbeddedFont 加载嵌入的字体文件
func loadEmbeddedFont() fyne.Resource {
	// 支持的字体文件扩展名
	fontExtensions := []string{".ttf", ".ttc", ".otf"}

	// 直接读取fonts目录下的文件
	entries, err := embeddedFonts.ReadDir("fonts")
	if err != nil {
		return nil
	}

	// 按优先级查找字体文件（优先查找常见的字体文件名）
	preferredNames := []string{"chinese.ttf", "chinese.ttc", "msyh.ttc", "simsun.ttc", "font.ttf", "font.ttc"}

	// 先查找优先字体
	for _, preferredName := range preferredNames {
		for _, entry := range entries {
			if entry.Name() == preferredName && !entry.IsDir() {
				data, err := embeddedFonts.ReadFile("fonts/" + preferredName)
				if err == nil && len(data) > 0 {
					return fyne.NewStaticResource(preferredName, data)
				}
			}
		}
	}

	// 如果优先字体没找到，查找任何字体文件
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		for _, ext := range fontExtensions {
			if filepath.Ext(name) == ext {
				data, err := embeddedFonts.ReadFile("fonts/" + name)
				if err == nil && len(data) > 0 {
					return fyne.NewStaticResource(name, data)
				}
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
