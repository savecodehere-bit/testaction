package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"image/color"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
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

// ServiceInfo 服务信息（从注册中心获取）
type ServiceInfo struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Port    int    `json:"port"`
	URL     string `json:"url"`
}

// GatewayService 网关服务
type GatewayService struct {
	port         int
	registryURL  string
	services     map[string]string // 服务名 -> URL
	mu           sync.RWMutex
	logContainer *fyne.Container
	logScroll    *container.Scroll
	statusLabel  *widget.Label
	servicesList *widget.List
	servicesData []ServiceItem
	muData       sync.RWMutex
	updateListFn func() // 更新列表的函数
}

// ServiceItem 服务列表项
type ServiceItem struct {
	Name string
	URL  string
}

// NewGatewayService 创建新的网关服务
func NewGatewayService(port int, registryURL string, logContainer *fyne.Container, logScroll *container.Scroll, statusLabel *widget.Label, servicesList *widget.List) *GatewayService {
	return &GatewayService{
		port:         port,
		registryURL:  registryURL,
		services:     make(map[string]string),
		logContainer: logContainer,
		logScroll:    logScroll,
		statusLabel:  statusLabel,
		servicesList: servicesList,
		servicesData: make([]ServiceItem, 0),
	}
}

// logMessage 添加日志消息（彩色显示）
func (gs *GatewayService) logMessage(msg string) {
	if gs.logContainer != nil {
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
		} else if contains(msg, "启动") || contains(msg, "就绪") || contains(msg, "路由") {
			msgColor = color.NRGBA{R: 0, G: 150, B: 255, A: 255} // 蓝色
		} else {
			msgColor = color.NRGBA{R: 200, G: 200, B: 200, A: 255} // 灰色（默认）
		}

		// 创建带颜色的文本
		logText := canvas.NewText(fullMsg, msgColor)
		logText.TextStyle = fyne.TextStyle{Monospace: true}
		logText.Alignment = fyne.TextAlignLeading

		// 添加到容器
		gs.logContainer.Add(logText)

		// 限制日志条目数量（保留最后200条）
		if len(gs.logContainer.Objects) > 200 {
			oldObjs := gs.logContainer.Objects
			gs.logContainer.Objects = oldObjs[len(oldObjs)-200:]
			gs.logContainer.Refresh()
		}

		// 滚动到底部
		if gs.logScroll != nil {
			gs.logScroll.ScrollToBottom()
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
func (gs *GatewayService) updateStatus() {
	if gs.statusLabel != nil {
		gs.mu.RLock()
		serviceCount := len(gs.services)
		gs.mu.RUnlock()
		gs.statusLabel.SetText(fmt.Sprintf("状态: 运行中 | 端口: %d | 已发现服务: %d | 注册中心: %s", gs.port, serviceCount, gs.registryURL))
	}
}

// RegisterToRegistry 注册到服务注册中心
func (gs *GatewayService) RegisterToRegistry() {
	if gs.registryURL == "" {
		return
	}

	serviceInfo := map[string]interface{}{
		"name":    "gateway-service",
		"address": "localhost",
		"port":    gs.port,
	}

	jsonData, _ := json.Marshal(serviceInfo)
	resp, err := http.Post(gs.registryURL+"/register", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		gs.logMessage(fmt.Sprintf("警告: 无法注册到服务注册中心: %v", err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		gs.logMessage("✓ 已注册到服务注册中心")
		// 启动心跳协程
		go gs.startHeartbeat()
	}
}

// startHeartbeat 启动心跳（每5秒发送一次）
func (gs *GatewayService) startHeartbeat() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if gs.registryURL != "" {
			http.Post(gs.registryURL+"/heartbeat?name=gateway-service", "application/json", nil)
		}
	}
}

// UnregisterFromRegistry 从服务注册中心注销（同步等待完成）
func (gs *GatewayService) UnregisterFromRegistry() {
	if gs.registryURL == "" {
		return
	}

	gs.logMessage("正在注销服务...")

	// 创建带超时的HTTP客户端
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Post(gs.registryURL+"/unregister?name=gateway-service", "application/json", nil)
	if err != nil {
		gs.logMessage(fmt.Sprintf("注销失败: %v", err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		gs.logMessage("✓ 已从服务注册中心注销")
	} else {
		gs.logMessage(fmt.Sprintf("注销失败，状态码: %d", resp.StatusCode))
	}
}

// DiscoverService 从注册中心发现单个服务
func (gs *GatewayService) DiscoverService(serviceName string) {
	if gs.registryURL == "" {
		return
	}

	resp, err := http.Get(gs.registryURL + "/discover?name=" + serviceName)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var service ServiceInfo
		if err := json.NewDecoder(resp.Body).Decode(&service); err == nil {
			gs.mu.Lock()
			oldURL, existed := gs.services[serviceName]
			gs.services[serviceName] = service.URL
			gs.mu.Unlock()

			if !existed {
				gs.logMessage(fmt.Sprintf("✓ 新服务注册: %s -> %s", serviceName, service.URL))
			} else if oldURL != service.URL {
				gs.logMessage(fmt.Sprintf("⚠ 服务地址变更: %s %s -> %s", serviceName, oldURL, service.URL))
			}
			gs.updateStatus()
		}
	}
}

// RefreshAllServices 从服务中心获取所有服务列表并更新
func (gs *GatewayService) RefreshAllServices() {
	if gs.registryURL == "" {
		return
	}

	resp, err := http.Get(gs.registryURL + "/services")
	if err != nil {
		gs.logMessage(fmt.Sprintf("✗ 无法获取服务列表: %v", err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return
	}

	var services []ServiceInfo
	if err := json.NewDecoder(resp.Body).Decode(&services); err != nil {
		return
	}

	// 获取当前服务列表的快照
	gs.mu.RLock()
	oldServices := make(map[string]string)
	for k, v := range gs.services {
		oldServices[k] = v
	}
	gs.mu.RUnlock()

	// 构建新的服务列表
	newServices := make(map[string]string)
	for _, service := range services {
		// 跳过网关服务自己
		if service.Name == "gateway-service" {
			continue
		}
		newServices[service.Name] = service.URL
	}

	// 检测新增的服务
	gs.mu.Lock()
	for name, url := range newServices {
		if oldURL, existed := gs.services[name]; !existed {
			gs.services[name] = url
			gs.logMessage(fmt.Sprintf("✓ 新服务上线: %s -> %s", name, url))
		} else if oldURL != url {
			gs.services[name] = url
			gs.logMessage(fmt.Sprintf("⚠ 服务地址变更: %s %s -> %s", name, oldURL, url))
		}
	}

	// 检测下线的服务
	for name, oldURL := range oldServices {
		if _, exists := newServices[name]; !exists {
			delete(gs.services, name)
			gs.logMessage(fmt.Sprintf("✗ 服务下线: %s (%s)", name, oldURL))
		}
	}
	gs.mu.Unlock()

	// 更新服务列表显示
	gs.updateServicesList()
	gs.updateStatus()
}

// updateServicesList 更新服务列表显示
func (gs *GatewayService) updateServicesList() {
	gs.mu.RLock()
	servicesData := make([]ServiceItem, 0, len(gs.services))
	for name, url := range gs.services {
		servicesData = append(servicesData, ServiceItem{
			Name: name,
			URL:  url,
		})
	}
	gs.mu.RUnlock()

	// 按服务名称排序，确保顺序稳定
	sort.Slice(servicesData, func(i, j int) bool {
		return servicesData[i].Name < servicesData[j].Name
	})

	gs.muData.Lock()
	gs.servicesData = servicesData
	gs.muData.Unlock()

	if gs.servicesList != nil {
		gs.servicesList.Refresh()
	}

	// 调用外部设置的更新函数
	if gs.updateListFn != nil {
		gs.updateListFn()
	}
}

// getServiceURL 获取服务URL（带自动发现）
func (gs *GatewayService) getServiceURL(serviceName string) string {
	gs.mu.RLock()
	url := gs.services[serviceName]
	gs.mu.RUnlock()

	// 如果URL为空，尝试重新发现
	if url == "" {
		gs.DiscoverService(serviceName)
		gs.mu.RLock()
		url = gs.services[serviceName]
		gs.mu.RUnlock()
	}

	return url
}

// proxyRequest 代理请求到目标服务
func (gs *GatewayService) proxyRequest(w http.ResponseWriter, r *http.Request, targetURL string) {
	// 解析目标URL
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		http.Error(w, "Invalid target URL", http.StatusInternalServerError)
		return
	}

	// 添加查询参数（保留原有的查询参数，如果有的话）
	if r.URL.RawQuery != "" {
		if parsedURL.RawQuery != "" {
			parsedURL.RawQuery += "&" + r.URL.RawQuery
		} else {
			parsedURL.RawQuery = r.URL.RawQuery
		}
	}

	// 创建新请求
	req, err := http.NewRequest(r.Method, parsedURL.String(), r.Body)
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	// 复制请求头
	for key, values := range r.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// 发送请求
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		gs.logMessage(fmt.Sprintf("✗ 转发请求失败: %v", err))
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	// 复制响应头
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// 设置状态码
	w.WriteHeader(resp.StatusCode)

	// 复制响应体
	io.Copy(w, resp.Body)
}

// handleUserService 处理用户服务请求
func (gs *GatewayService) handleUserService(w http.ResponseWriter, r *http.Request) {
	userServiceURL := gs.getServiceURL("user-service")
	if userServiceURL == "" {
		gs.logMessage("✗ 用户服务不可用")
		http.Error(w, "User service unavailable", http.StatusServiceUnavailable)
		return
	}

	// 提取路径（去掉 /api/user）
	path := strings.TrimPrefix(r.URL.Path, "/api/user")
	if path == "" {
		path = "/user"
	} else {
		path = "/user" + path
	}

	targetURL := userServiceURL + path
	gs.logMessage(fmt.Sprintf("→ 转发到用户服务: %s %s", r.Method, targetURL))
	gs.proxyRequest(w, r, targetURL)
}

// handleOrderService 处理订单服务请求
func (gs *GatewayService) handleOrderService(w http.ResponseWriter, r *http.Request) {
	orderServiceURL := gs.getServiceURL("order-service")
	if orderServiceURL == "" {
		gs.logMessage("✗ 订单服务不可用")
		http.Error(w, "Order service unavailable", http.StatusServiceUnavailable)
		return
	}

	// 提取路径（去掉 /api/order）
	path := strings.TrimPrefix(r.URL.Path, "/api/order")
	if path == "" {
		path = "/order"
	} else {
		path = "/order" + path
	}

	targetURL := orderServiceURL + path
	gs.logMessage(fmt.Sprintf("→ 转发到订单服务: %s %s", r.Method, targetURL))
	gs.proxyRequest(w, r, targetURL)
}

// handleDynamicRoute 动态路由处理（根据服务名自动路由）
func (gs *GatewayService) handleDynamicRoute(w http.ResponseWriter, r *http.Request) {
	// 路径格式: /api/{service-name}/...
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/"), "/")
	if len(pathParts) == 0 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	serviceName := pathParts[0]
	serviceURL := gs.getServiceURL(serviceName)
	if serviceURL == "" {
		gs.logMessage(fmt.Sprintf("✗ 服务不可用: %s", serviceName))
		http.Error(w, fmt.Sprintf("Service %s unavailable", serviceName), http.StatusServiceUnavailable)
		return
	}

	// 构建目标路径（保留原始路径结构）
	remainingPath := strings.Join(pathParts[1:], "/")

	// 根据服务名映射到实际路径
	var targetPath string
	if remainingPath == "" {
		// 如果没有剩余路径，根据服务名设置默认路径
		switch serviceName {
		case "user-service", "user":
			targetPath = "/user"
		case "order-service", "order":
			targetPath = "/order"
		default:
			targetPath = "/" + serviceName
		}
	} else {
		// 有剩余路径，直接使用（remainingPath已经包含了服务路径，如"user"或"order"）
		// 确保路径以 / 开头
		if !strings.HasPrefix(remainingPath, "/") {
			targetPath = "/" + remainingPath
		} else {
			targetPath = remainingPath
		}
	}

	targetURL := serviceURL + targetPath
	// 如果有查询参数，添加到日志中
	queryStr := ""
	if r.URL.RawQuery != "" {
		queryStr = "?" + r.URL.RawQuery
	}
	gs.logMessage(fmt.Sprintf("→ 动态路由: %s -> %s %s%s", serviceName, r.Method, targetURL, queryStr))
	gs.proxyRequest(w, r, targetURL)
}

// handleHealth 健康检查
func (gs *GatewayService) handleHealth(w http.ResponseWriter, r *http.Request) {
	gs.mu.RLock()
	services := make(map[string]string)
	for k, v := range gs.services {
		services[k] = v
	}
	gs.mu.RUnlock()

	response := map[string]interface{}{
		"status":   "ok",
		"gateway":  fmt.Sprintf("http://localhost:%d", gs.port),
		"services": services,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	// 创建GUI应用
	myApp := app.New()
	// 设置支持中文的主题（使用系统默认字体，支持中文）
	myApp.Settings().SetTheme(newChineseTheme())
	myWindow := myApp.NewWindow("API网关服务 (端口: 8083)")
	myWindow.Resize(fyne.NewSize(800, 600))

	// 创建日志显示区域（使用canvas.Text支持彩色显示）
	logContainer := container.NewVBox()
	// 添加初始消息
	initText := canvas.NewText("API网关服务启动中...", color.NRGBA{R: 200, G: 200, B: 200, A: 255})
	initText.TextStyle = fyne.TextStyle{Monospace: true}
	logContainer.Add(initText)
	logScroll := container.NewScroll(logContainer)
	logScroll.SetMinSize(fyne.NewSize(0, 0))

	// 创建状态标签
	statusLabel := widget.NewLabel("状态: 启动中...")

	// 创建服务列表
	servicesData := make([]ServiceItem, 0)
	servicesList := widget.NewList(
		func() int {
			return len(servicesData)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabel(""),
				widget.NewLabel(""),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < len(servicesData) {
				service := servicesData[id]
				boxes := obj.(*fyne.Container)
				boxes.Objects[0].(*widget.Label).SetText(service.Name)
				boxes.Objects[1].(*widget.Label).SetText(service.URL)
			}
		},
	)

	port := 8083
	registryURL := "http://localhost:8080"

	service := NewGatewayService(port, registryURL, logContainer, logScroll, statusLabel, servicesList)

	// 设置路由
	// 固定路由（向后兼容）
	http.HandleFunc("/api/user", service.handleUserService)
	http.HandleFunc("/api/user/", service.handleUserService)
	http.HandleFunc("/api/order", service.handleOrderService)
	http.HandleFunc("/api/order/", service.handleOrderService)
	// 动态路由（支持任意服务）
	http.HandleFunc("/api/", service.handleDynamicRoute)
	http.HandleFunc("/health", service.handleHealth)

	// 启动HTTP服务器
	go func() {
		service.logMessage(fmt.Sprintf("API网关服务启动在端口 %d", port))
		service.logMessage(fmt.Sprintf("服务注册中心: %s", registryURL))

		// 注册到服务注册中心
		service.RegisterToRegistry()

		// 立即获取所有服务列表
		service.RefreshAllServices()

		// 定期从服务中心获取所有服务列表（监听服务变动，每1秒刷新一次）
		go func() {
			ticker := time.NewTicker(1 * time.Second)
			defer ticker.Stop()
			for range ticker.C {
				service.RefreshAllServices()
			}
		}()

		service.logMessage("API端点:")
		service.logMessage(fmt.Sprintf("  GET  http://localhost:%d/api/user?id=1 - 获取用户", port))
		service.logMessage(fmt.Sprintf("  GET  http://localhost:%d/api/user - 列出所有用户", port))
		service.logMessage(fmt.Sprintf("  POST http://localhost:%d/api/user - 创建用户", port))
		service.logMessage(fmt.Sprintf("  GET  http://localhost:%d/api/order?id=1 - 获取订单", port))
		service.logMessage(fmt.Sprintf("  GET  http://localhost:%d/api/order?user_id=1 - 获取用户的订单", port))
		service.logMessage(fmt.Sprintf("  POST http://localhost:%d/api/order - 创建订单", port))
		service.logMessage(fmt.Sprintf("  GET  http://localhost:%d/health - 健康检查（查看所有已发现的服务）", port))
		service.logMessage("")
		service.logMessage("网关会自动监听服务中心的服务列表变动")
		service.logMessage("每1秒从服务中心获取最新服务列表")

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

	// 更新服务列表的刷新函数（使用闭包共享变量）
	refreshServicesList := func() {
		service.mu.RLock()
		servicesData = make([]ServiceItem, 0, len(service.services))
		for name, url := range service.services {
			servicesData = append(servicesData, ServiceItem{
				Name: name,
				URL:  url,
			})
		}
		service.mu.RUnlock()

		// 按服务名称排序，确保顺序稳定
		sort.Slice(servicesData, func(i, j int) bool {
			return servicesData[i].Name < servicesData[j].Name
		})

		servicesList.Refresh()
	}

	// 设置更新列表函数
	service.updateListFn = refreshServicesList

	// 创建HSplit并设置分割比例（左侧占30%，右侧占70%）
	split := container.NewHSplit(
		container.NewBorder(
			widget.NewLabel("已发现服务列表"), nil, nil, nil,
			servicesList,
		),
		container.NewBorder(
			widget.NewLabel("日志输出"), nil, nil, nil,
			logScroll,
		),
	)
	split.SetOffset(0.3) // 左侧占30%宽度

	// 创建UI布局
	content := container.NewBorder(
		statusLabel,
		nil,
		nil,
		nil,
		split,
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
