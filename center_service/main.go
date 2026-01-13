package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"image/color"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
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
// 使用方法：将中文字体文件复制到 fonts 目录下
// 支持的格式：.ttf, .otf（注意：Fyne不支持.ttc字体集合文件）
// 如果没有字体文件，程序会自动使用系统字体作为fallback
//go:embed fonts/*
var embeddedFonts embed.FS

// ServiceInfo 服务信息
type ServiceInfo struct {
	Name          string    `json:"name"`
	Address       string    `json:"address"`
	Port          int       `json:"port"`
	URL           string    `json:"url"`
	LastHeartbeat time.Time `json:"last_heartbeat"`
}

// LogMessage 日志消息结构
type LogMessage struct {
	Text  string
	Color color.Color
}

// ServiceRegistry 服务注册中心
type ServiceRegistry struct {
	services     map[string]*ServiceInfo
	mu           sync.RWMutex
	logContainer *fyne.Container
	logScroll    *container.Scroll
	servicesList *widget.List
	servicesData []*ServiceInfo
	muData       sync.RWMutex
	updateListFn func()
	refreshChan  chan struct{}   // 用于触发UI刷新
	logChan      chan LogMessage // 用于在主线程中更新日志
}

// NewServiceRegistry 创建新的服务注册中心
func NewServiceRegistry(logContainer *fyne.Container, logScroll *container.Scroll, servicesList *widget.List, refreshChan chan struct{}, logChan chan LogMessage) *ServiceRegistry {
	return &ServiceRegistry{
		services:     make(map[string]*ServiceInfo),
		logContainer: logContainer,
		logScroll:    logScroll,
		servicesList: servicesList,
		servicesData: make([]*ServiceInfo, 0),
		refreshChan:  refreshChan,
		logChan:      logChan,
	}
}

// logMessage 添加日志消息（通过channel在主线程中更新）
func (sr *ServiceRegistry) logMessage(msg string) {
	if sr.logChan != nil {
		timestamp := time.Now().Format("15:04:05")
		timestampText := fmt.Sprintf("[%s] ", timestamp)

		// 根据消息内容确定颜色
		var msgColor color.Color
		if contains(msg, "注册") {
			msgColor = color.NRGBA{R: 0, G: 200, B: 0, A: 255} // 绿色
		} else if contains(msg, "注销") {
			msgColor = color.NRGBA{R: 255, G: 165, B: 0, A: 255} // 橙色
		} else if contains(msg, "过期") || contains(msg, "移除") {
			msgColor = color.NRGBA{R: 255, G: 0, B: 0, A: 255} // 红色
		} else if contains(msg, "启动") || contains(msg, "就绪") {
			msgColor = color.NRGBA{R: 0, G: 150, B: 255, A: 255} // 蓝色
		} else {
			msgColor = color.NRGBA{R: 200, G: 200, B: 200, A: 255} // 灰色（默认）
		}

		logMsg := LogMessage{
			Text:  fmt.Sprintf("%s%s\n", timestampText, msg),
			Color: msgColor,
		}

		select {
		case sr.logChan <- logMsg:
		default:
			// channel满了，跳过
		}
	}
}

// contains 检查字符串是否包含子串（不区分大小写）
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			containsHelper(s, substr))))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// splitLogMessage 解析日志消息，返回时间戳和消息内容
func splitLogMessage(msg string) []string {
	// 格式: [15:04:05] 消息内容\n
	if len(msg) < 12 || msg[0] != '[' {
		return []string{"", msg}
	}

	// 查找时间戳结束位置
	endIdx := -1
	for i := 1; i < len(msg) && i < 12; i++ {
		if msg[i] == ']' {
			endIdx = i
			break
		}
	}

	if endIdx == -1 {
		return []string{"", msg}
	}

	timestamp := msg[:endIdx+1] + " "
	message := msg[endIdx+2:]
	return []string{timestamp, message}
}

// updateServicesList 更新服务列表显示
// 这个函数会：
// 1. 调用refreshServicesList更新数据
// 2. 发送refreshChan信号触发UI刷新
func (sr *ServiceRegistry) updateServicesList() {
	// 先更新数据
	if sr.updateListFn != nil {
		sr.updateListFn()
	}
	// 然后触发UI刷新（必须在数据更新后）
	if sr.refreshChan != nil {
		select {
		case sr.refreshChan <- struct{}{}:
		default:
			// channel满了，跳过（说明已经有刷新请求在队列中）
		}
	}
}

// Register 注册服务
func (sr *ServiceRegistry) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var service ServiceInfo
	if err := json.NewDecoder(r.Body).Decode(&service); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	service.LastHeartbeat = time.Now()
	service.URL = fmt.Sprintf("http://%s:%d", service.Address, service.Port)

	sr.mu.Lock()
	sr.services[service.Name] = &service
	sr.mu.Unlock()

	// 在锁外执行日志和UI更新，避免死锁
	msg := fmt.Sprintf("服务注册: %s -> %s", service.Name, service.URL)
	sr.logMessage(msg)
	// 立即更新服务列表并刷新UI
	sr.updateServicesList()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "registered",
		"name":   service.Name,
		"url":    service.URL,
	})
}

// Discover 发现服务
func (sr *ServiceRegistry) Discover(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	serviceName := r.URL.Query().Get("name")
	if serviceName == "" {
		http.Error(w, "Missing name parameter", http.StatusBadRequest)
		return
	}

	sr.mu.RLock()
	service, exists := sr.services[serviceName]
	sr.mu.RUnlock()

	if !exists {
		http.Error(w, "Service not found", http.StatusNotFound)
		return
	}

	// 检查服务是否过期（10秒未心跳则认为服务下线）
	if time.Since(service.LastHeartbeat) > 10*time.Second {
		sr.mu.Lock()
		delete(sr.services, serviceName)
		sr.mu.Unlock()
		sr.updateServicesList()
		http.Error(w, "Service expired", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(service)
}

// ListServices 列出所有服务
func (sr *ServiceRegistry) ListServices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sr.mu.RLock()
	services := make([]*ServiceInfo, 0, len(sr.services))
	for _, service := range sr.services {
		// 过滤过期服务（10秒未心跳则认为服务下线）
		if time.Since(service.LastHeartbeat) <= 10*time.Second {
			services = append(services, service)
		}
	}
	sr.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(services)
}

// Heartbeat 心跳更新
func (sr *ServiceRegistry) Heartbeat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	serviceName := r.URL.Query().Get("name")
	if serviceName == "" {
		http.Error(w, "Missing name parameter", http.StatusBadRequest)
		return
	}

	sr.mu.Lock()
	if service, exists := sr.services[serviceName]; exists {
		service.LastHeartbeat = time.Now()
	}
	sr.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// Unregister 注销服务
func (sr *ServiceRegistry) Unregister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	serviceName := r.URL.Query().Get("name")
	if serviceName == "" {
		http.Error(w, "Missing name parameter", http.StatusBadRequest)
		return
	}

	sr.mu.Lock()
	existed := false
	if _, exists := sr.services[serviceName]; exists {
		delete(sr.services, serviceName)
		existed = true
	}
	sr.mu.Unlock()

	// 如果服务存在，记录日志并更新UI（在锁外执行，避免死锁）
	if existed {
		msg := fmt.Sprintf("服务注销: %s", serviceName)
		sr.logMessage(msg)
		// 立即更新服务列表并刷新UI
		sr.updateServicesList()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "unregistered",
		"name":   serviceName,
	})
}

func main() {
	// 创建GUI应用
	myApp := app.New()
	// 设置支持中文的主题（使用系统默认字体，支持中文）
	myApp.Settings().SetTheme(newChineseTheme())
	myWindow := myApp.NewWindow("服务注册中心 (端口: 8080)")
	myWindow.Resize(fyne.NewSize(800, 600))

	// 创建日志显示区域（使用canvas.Text支持彩色显示）
	logContainer := container.NewVBox()
	// 添加初始消息
	initText := canvas.NewText("服务注册中心启动中...", color.NRGBA{R: 200, G: 200, B: 200, A: 255})
	initText.TextStyle = fyne.TextStyle{Monospace: true}
	logContainer.Add(initText)
	logScroll := container.NewScroll(logContainer)
	logScroll.SetMinSize(fyne.NewSize(0, 0))

	// 创建刷新channel，用于在主线程中刷新UI
	refreshChan := make(chan struct{}, 10)
	// 创建日志channel，用于在主线程中更新日志
	logChan := make(chan LogMessage, 100)

	// 使用mutex保护服务列表数据
	var servicesDataMu sync.RWMutex
	var servicesData []*ServiceInfo

	// 创建服务列表
	servicesList := widget.NewList(
		func() int {
			servicesDataMu.RLock()
			defer servicesDataMu.RUnlock()
			return len(servicesData)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabel(""),
				widget.NewLabel(""),
				widget.NewLabel(""),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			servicesDataMu.RLock()
			defer servicesDataMu.RUnlock()
			if id < len(servicesData) {
				service := servicesData[id]
				boxes := obj.(*fyne.Container)
				boxes.Objects[0].(*widget.Label).SetText(service.Name)
				boxes.Objects[1].(*widget.Label).SetText(fmt.Sprintf(":%d", service.Port))
				boxes.Objects[2].(*widget.Label).SetText(service.URL)
			}
		},
	)

	// 创建注册中心实例
	registry := NewServiceRegistry(logContainer, logScroll, servicesList, refreshChan, logChan)

	// 在主线程中处理日志更新和UI刷新（必须在创建registry之后启动）
	// 先启动日志处理goroutine，确保能接收日志消息
	logHandlerReady := make(chan bool, 1)
	go func() {
		logHandlerReady <- true
		for {
			select {
			case logMsg := <-logChan:
				// 更新日志容器（线程安全）
				if registry.logContainer != nil {
					// 创建带颜色的文本
					logText := canvas.NewText(logMsg.Text, logMsg.Color)
					logText.TextStyle = fyne.TextStyle{Monospace: true}
					logText.Alignment = fyne.TextAlignLeading

					// 添加到容器
					registry.logContainer.Add(logText)

					// 限制日志条目数量，避免内存占用过大（保留最后200条）
					if len(registry.logContainer.Objects) > 200 {
						// 移除最旧的条目
						oldObjs := registry.logContainer.Objects
						registry.logContainer.Objects = oldObjs[len(oldObjs)-200:]
						registry.logContainer.Refresh()
					}

					// 滚动到底部
					logScroll.ScrollToBottom()
				}
			case <-refreshChan:
				// 刷新列表
				servicesList.Refresh()
			}
		}
	}()

	// 等待日志处理goroutine启动
	<-logHandlerReady

	// 更新服务列表数据的函数（只更新数据，不触发UI刷新）
	refreshServicesList := func() {
		registry.mu.RLock()
		newServicesData := make([]*ServiceInfo, 0, len(registry.services))
		for _, service := range registry.services {
			// 过滤过期服务（10秒未心跳则认为服务下线）
			if time.Since(service.LastHeartbeat) <= 10*time.Second {
				newServicesData = append(newServicesData, service)
			}
		}
		registry.mu.RUnlock()

		// 按服务名称排序，确保顺序稳定
		sort.Slice(newServicesData, func(i, j int) bool {
			return newServicesData[i].Name < newServicesData[j].Name
		})

		// 更新服务列表数据（使用mutex保护）
		servicesDataMu.Lock()
		servicesData = newServicesData
		registry.servicesData = newServicesData
		servicesDataMu.Unlock()
	}
	registry.updateListFn = refreshServicesList

	// 定期刷新UI（每200毫秒检查一次，确保UI及时更新）
	// 这个goroutine会定期更新数据并触发UI刷新，作为fallback机制
	go func() {
		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()
		for range ticker.C {
			// 更新数据
			refreshServicesList()
			// 触发UI刷新
			select {
			case refreshChan <- struct{}{}:
			default:
				// channel满了，跳过
			}
		}
	}()

	// 设置HTTP路由
	http.HandleFunc("/register", registry.Register)
	http.HandleFunc("/unregister", registry.Unregister)
	http.HandleFunc("/discover", registry.Discover)
	http.HandleFunc("/services", registry.ListServices)
	http.HandleFunc("/heartbeat", registry.Heartbeat)

	port := 8080

	// 启动HTTP服务器
	go func() {
		registry.logMessage(fmt.Sprintf("服务注册中心启动在端口 %d", port))
		registry.logMessage("API端点:")
		registry.logMessage(fmt.Sprintf("  POST http://localhost:%d/register - 注册服务", port))
		registry.logMessage(fmt.Sprintf("  POST http://localhost:%d/unregister?name=服务名 - 注销服务", port))
		registry.logMessage(fmt.Sprintf("  GET  http://localhost:%d/discover?name=服务名 - 发现服务", port))
		registry.logMessage(fmt.Sprintf("  GET  http://localhost:%d/services - 列出所有服务", port))
		registry.logMessage(fmt.Sprintf("  POST http://localhost:%d/heartbeat?name=服务名 - 发送心跳", port))
		registry.logMessage("服务已就绪，等待服务注册...")
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
	}()

	// 定期清理过期服务（每2秒检查一次，10秒未心跳则移除）
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			removed := false
			removedNames := make([]string, 0)
			registry.mu.Lock()
			for name, service := range registry.services {
				if time.Since(service.LastHeartbeat) > 10*time.Second {
					delete(registry.services, name)
					removedNames = append(removedNames, name)
					removed = true
				}
			}
			registry.mu.Unlock()
			// 如果有服务被移除，记录日志并更新UI（在锁外执行）
			if removed {
				for _, name := range removedNames {
					registry.logMessage(fmt.Sprintf("服务过期已移除: %s", name))
				}
				registry.updateServicesList()
			}
		}
	}()

	// 创建UI布局
	statusLabel := widget.NewLabel(fmt.Sprintf("状态: 运行中 | 端口: %d | 已注册服务: 0", port))

	// 定期更新状态
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			registry.mu.RLock()
			count := len(registry.services)
			registry.mu.RUnlock()
			statusLabel.SetText(fmt.Sprintf("状态: 运行中 | 端口: %d | 已注册服务: %d", port, count))
		}
	}()

	// 创建HSplit并设置分割比例（左侧占30%，右侧占70%）
	split := container.NewHSplit(
		container.NewBorder(
			widget.NewLabel("已注册服务列表"), nil, nil, nil,
			servicesList,
		),
		container.NewBorder(
			widget.NewLabel("日志输出"), nil, nil, nil,
			logScroll,
		),
	)
	split.SetOffset(0.3) // 左侧占30%宽度

	content := container.NewBorder(
		statusLabel,
		nil,
		nil,
		nil,
		split,
	)

	myWindow.SetContent(content)
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
		
		// 注意：Fyne不支持.ttc字体集合，只使用.ttf文件
		fontPaths = []string{
			filepath.Join(windir, "Fonts", "simhei.ttf"),  // SimHei (黑体)
			filepath.Join(windir, "Fonts", "simsun.ttf"),  // SimSun (宋体)
			filepath.Join(windir, "Fonts", "simkai.ttf"),  // SimKai (楷体)
			filepath.Join(windir, "Fonts", "simli.ttf"),   // SimLi (隶书)
		}
		
		if windir != "C:\\Windows" {
			fontPaths = append(fontPaths,
				filepath.Join("C:\\Windows", "Fonts", "simhei.ttf"),
				filepath.Join("C:\\Windows", "Fonts", "simsun.ttf"),
			)
		}
	case "darwin": // macOS
		// 只使用.ttf文件，Fyne不支持.ttc
		fontPaths = []string{
			"/Library/Fonts/Arial Unicode.ttf",
		}
	case "linux":
		// 只使用.ttf文件，Fyne不支持.ttc
		fontPaths = []string{
			"/usr/share/fonts/truetype/droid/DroidSansFallbackFull.ttf",
		}
	}

	// 尝试加载系统字体文件（只加载.ttf和.otf，跳过.ttc）
	for _, path := range fontPaths {
		// 检查文件扩展名，只加载.ttf和.otf文件
		ext := filepath.Ext(path)
		if ext != ".ttf" && ext != ".otf" {
			continue // 跳过非.ttf/.otf文件
		}

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
// 注意：Fyne不支持.ttc（TrueType Collection）字体集合文件，只支持.ttf和.otf
func loadEmbeddedFont() fyne.Resource {
	// 直接读取fonts目录下的文件
	entries, err := embeddedFonts.ReadDir("fonts")
	if err != nil {
		return nil
	}

	// 按优先级查找字体文件（只查找.ttf和.otf，跳过.ttc）
	preferredNames := []string{"chinese.ttf", "chinese.otf", "msyh.ttf", "simsun.ttf", "font.ttf", "font.otf"}

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

	// 如果优先字体没找到，查找任何.ttf或.otf字体文件（跳过.ttc）
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		ext := filepath.Ext(name)
		// 只加载.ttf和.otf文件，跳过.ttc（Fyne不支持）
		if ext == ".ttf" || ext == ".otf" {
			data, err := embeddedFonts.ReadFile("fonts/" + name)
			if err == nil && len(data) > 0 {
				return fyne.NewStaticResource(name, data)
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
