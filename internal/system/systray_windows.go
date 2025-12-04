//go:build windows
// +build windows

package system

import (
	"context"
	_ "embed"
	"os"
	"sync"
	"time"

	"github.com/getlantern/systray"
	log "github.com/sirupsen/logrus"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed icon.ico
var trayIconData []byte

// SystemTray 系统托盘管理器
type SystemTray struct {
	ctx          context.Context
	mShow        *systray.MenuItem
	mQuit        *systray.MenuItem
	isRunning    bool
	mu           sync.Mutex
	quitCallback func() // 退出回调
}

// NewSystemTray 创建系统托盘管理器
func NewSystemTray(ctx context.Context) *SystemTray {
	return &SystemTray{
		ctx: ctx,
	}
}

// SetQuitCallback 设置退出回调
func (s *SystemTray) SetQuitCallback(callback func()) {
	s.quitCallback = callback
}

// ShowWindow 显示窗口
func (s *SystemTray) ShowWindow() {
	log.Info("System tray: Showing window")
	runtime.WindowShow(s.ctx)
	runtime.WindowUnminimise(s.ctx)
	// 将窗口置于前台
	runtime.WindowSetAlwaysOnTop(s.ctx, true)
	time.Sleep(100 * time.Millisecond)
	runtime.WindowSetAlwaysOnTop(s.ctx, false)
}

// HideWindow 隐藏窗口
func (s *SystemTray) HideWindow() {
	log.Info("System tray: Hiding window")
	runtime.WindowHide(s.ctx)
}

// QuitApp 退出应用
func (s *SystemTray) QuitApp() {
	log.Info("System tray: Quitting application")
	s.mu.Lock()
	s.isRunning = false
	s.mu.Unlock()

	// 先退出托盘
	systray.Quit()

	// 调用退出回调或直接退出
	if s.quitCallback != nil {
		s.quitCallback()
	} else {
		// 强制退出进程
		os.Exit(0)
	}
}

// Setup 设置系统托盘
func (s *SystemTray) Setup() error {
	s.mu.Lock()
	if s.isRunning {
		s.mu.Unlock()
		return nil
	}
	s.isRunning = true
	s.mu.Unlock()

	// 在后台启动系统托盘
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Errorf("System tray setup panic: %v", r)
			}
		}()

		systray.Run(s.onReady, s.onExit)
	}()

	log.Info("System tray setup completed")
	return nil
}

// onReady 托盘就绪回调
func (s *SystemTray) onReady() {
	// 设置托盘图标
	if len(trayIconData) > 0 {
		systray.SetIcon(trayIconData)
	}
	systray.SetTitle("AnyProxyAi")
	systray.SetTooltip("AnyProxyAi - API 代理管理器")

	// 添加菜单项
	s.mShow = systray.AddMenuItem("显示窗口", "显示主窗口")
	systray.AddSeparator()
	s.mQuit = systray.AddMenuItem("退出程序", "退出 AnyProxyAi")

	// 处理菜单点击事件
	go s.handleMenuEvents()

	// 定期刷新图标保持托盘活跃
	go s.keepAlive()
}

// keepAlive 保持托盘活跃
func (s *SystemTray) keepAlive() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		running := s.isRunning
		s.mu.Unlock()
		if !running {
			return
		}
		// 刷新图标
		if len(trayIconData) > 0 {
			systray.SetIcon(trayIconData)
		}
	}
}

// handleMenuEvents 处理菜单事件
func (s *SystemTray) handleMenuEvents() {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("Tray menu event handler panic: %v", r)
		}
	}()

	for {
		select {
		case <-s.mShow.ClickedCh:
			s.ShowWindow()

		case <-s.mQuit.ClickedCh:
			log.Info("Quit menu clicked")
			s.QuitApp()
			return
		}
	}
}

// onExit 托盘退出回调
func (s *SystemTray) onExit() {
	s.mu.Lock()
	s.isRunning = false
	s.mu.Unlock()
	log.Info("System tray exited")
}

// Quit 退出托盘
func (s *SystemTray) Quit() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.isRunning {
		s.isRunning = false
		systray.Quit()
	}
}
