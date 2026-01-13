package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image/color"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// TestClient 测试客户端
type TestClient struct {
	gatewayURL   string
	logContainer *fyne.Container
	logScroll    *container.Scroll
}

// NewTestClient 创建新的测试客户端
func NewTestClient(gatewayURL string, logContainer *fyne.Container, logScroll *container.Scroll) *TestClient {
	return &TestClient{
		gatewayURL:   gatewayURL,
		logContainer: logContainer,
		logScroll:    logScroll,
	}
}

// logMessage 添加日志消息（彩色显示）
func (tc *TestClient) logMessage(msg string, msgColor color.Color) {
	if tc.logContainer != nil {
		timestamp := time.Now().Format("15:04:05")
		fullMsg := fmt.Sprintf("[%s] %s", timestamp, msg)

		// 创建带颜色的文本
		logText := canvas.NewText(fullMsg, msgColor)
		logText.TextStyle = fyne.TextStyle{Monospace: true}
		logText.Alignment = fyne.TextAlignLeading

		// 添加到容器
		tc.logContainer.Add(logText)

		// 限制日志条目数量（保留最后200条）
		if len(tc.logContainer.Objects) > 200 {
			oldObjs := tc.logContainer.Objects
			tc.logContainer.Objects = oldObjs[len(oldObjs)-200:]
			tc.logContainer.Refresh()
		}

		// 滚动到底部
		if tc.logScroll != nil {
			tc.logScroll.ScrollToBottom()
		}
	}
}

// sendRequest 发送HTTP请求
func (tc *TestClient) sendRequest(method, url string, body []byte) {
	tc.logMessage(fmt.Sprintf("发送请求: %s %s", method, url), color.NRGBA{R: 0, G: 150, B: 255, A: 255})

	var req *http.Request
	var err error

	if body != nil {
		req, err = http.NewRequest(method, url, bytes.NewBuffer(body))
		if err != nil {
			tc.logMessage(fmt.Sprintf("错误: 创建请求失败 - %v", err), color.NRGBA{R: 255, G: 0, B: 0, A: 255})
			return
		}
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, err = http.NewRequest(method, url, nil)
		if err != nil {
			tc.logMessage(fmt.Sprintf("错误: 创建请求失败 - %v", err), color.NRGBA{R: 255, G: 0, B: 0, A: 255})
			return
		}
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		tc.logMessage(fmt.Sprintf("错误: 请求失败 - %v", err), color.NRGBA{R: 255, G: 0, B: 0, A: 255})
		return
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		tc.logMessage(fmt.Sprintf("错误: 读取响应失败 - %v", err), color.NRGBA{R: 255, G: 0, B: 0, A: 255})
		return
	}

	// 格式化JSON响应
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, respBody, "", "  "); err != nil {
		// 如果不是JSON，直接显示原始内容
		tc.logMessage(fmt.Sprintf("响应 [%d]: %s", resp.StatusCode, string(respBody)), color.NRGBA{R: 0, G: 200, B: 0, A: 255})
	} else {
		tc.logMessage(fmt.Sprintf("响应 [%d]:\n%s", resp.StatusCode, prettyJSON.String()), color.NRGBA{R: 0, G: 200, B: 0, A: 255})
	}
}

func main() {
	// 创建GUI应用
	myApp := app.New()
	// 设置支持中文的主题
	myApp.Settings().SetTheme(newChineseTheme())
	myWindow := myApp.NewWindow("API测试客户端")
	myWindow.Resize(fyne.NewSize(800, 600))

	gatewayURL := "http://localhost:8083"

	// 创建日志显示区域
	logContainer := container.NewVBox()
	initText := canvas.NewText("API测试客户端就绪...", color.NRGBA{R: 200, G: 200, B: 200, A: 255})
	initText.TextStyle = fyne.TextStyle{Monospace: true}
	logContainer.Add(initText)
	logScroll := container.NewScroll(logContainer)
	logScroll.SetMinSize(fyne.NewSize(0, 0))

	client := NewTestClient(gatewayURL, logContainer, logScroll)

	// 创建测试按钮
	btnGetUsers := widget.NewButton("获取所有用户", func() {
		url := fmt.Sprintf("%s/api/user-service/user", gatewayURL)
		client.sendRequest("GET", url, nil)
	})

	btnGetUser1 := widget.NewButton("获取用户1", func() {
		url := fmt.Sprintf("%s/api/user-service/user?id=1", gatewayURL)
		client.sendRequest("GET", url, nil)
	})

	btnCreateUser := widget.NewButton("创建用户", func() {
		url := fmt.Sprintf("%s/api/user-service/user", gatewayURL)
		userData := map[string]interface{}{
			"name":  "测试用户",
			"email": "test@example.com",
		}
		body, _ := json.Marshal(userData)
		client.sendRequest("POST", url, body)
	})

	btnGetOrders := widget.NewButton("获取所有订单", func() {
		url := fmt.Sprintf("%s/api/order-service/order", gatewayURL)
		client.sendRequest("GET", url, nil)
	})

	btnGetOrder1 := widget.NewButton("获取订单1", func() {
		url := fmt.Sprintf("%s/api/order-service/order?id=1", gatewayURL)
		client.sendRequest("GET", url, nil)
	})

	btnGetOrderWithUser := widget.NewButton("获取订单1(含用户信息)", func() {
		url := fmt.Sprintf("%s/api/order-service/order/with-user?id=1", gatewayURL)
		client.sendRequest("GET", url, nil)
	})

	btnCreateOrder := widget.NewButton("创建订单", func() {
		url := fmt.Sprintf("%s/api/order-service/order", gatewayURL)
		orderData := map[string]interface{}{
			"user_id": 1,
			"amount":  199.99,
			"status":  "待支付",
			"items":   []string{"商品A", "商品B"},
		}
		body, _ := json.Marshal(orderData)
		client.sendRequest("POST", url, body)
	})

	btnHealthCheck := widget.NewButton("网关健康检查", func() {
		url := fmt.Sprintf("%s/health", gatewayURL)
		client.sendRequest("GET", url, nil)
	})

	// 创建按钮布局
	buttonContainer := container.NewVBox(
		widget.NewLabel("用户服务测试:"),
		btnGetUsers,
		btnGetUser1,
		btnCreateUser,
		widget.NewSeparator(),
		widget.NewLabel("订单服务测试:"),
		btnGetOrders,
		btnGetOrder1,
		btnGetOrderWithUser,
		btnCreateOrder,
		widget.NewSeparator(),
		widget.NewLabel("网关测试:"),
		btnHealthCheck,
	)

	// 创建主布局
	content := container.NewHSplit(
		container.NewBorder(
			widget.NewLabel("测试操作"), nil, nil, nil,
			container.NewScroll(buttonContainer),
		),
		container.NewBorder(
			widget.NewLabel("请求日志"), nil, nil, nil,
			logScroll,
		),
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
