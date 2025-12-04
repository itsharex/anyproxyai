//go:build !windows
// +build !windows

package system

import (
	"context"

	log "github.com/sirupsen/logrus"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// SystemTray 系统托盘管理器 (Linux/macOS - 无托盘支持)
type SystemTray struct {
	ctx          context.Context
	quitCallback func()
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
	runtime.WindowShow(s.ctx)
	runtime.WindowUnminimise(s.ctx)
}

// HideWindow 隐藏窗口
func (s *SystemTray) HideWindow() {
	runtime.WindowHide(s.ctx)
}

// QuitApp 退出应用
func (s *SystemTray) QuitApp() {
	if s.quitCallback != nil {
		s.quitCallback()
	} else {
		runtime.Quit(s.ctx)
	}
}

// Setup 设置系统托盘 (Linux/macOS 不支持托盘)
func (s *SystemTray) Setup() error {
	log.Info("System tray not available on this platform")
	return nil
}

// Quit 退出托盘
func (s *SystemTray) Quit() {}
