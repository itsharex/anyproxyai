package system

import (
	"context"

	log "github.com/sirupsen/logrus"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// SystemTray 系统托盘管理器
type SystemTray struct {
	ctx context.Context
}

// NewSystemTray 创建系统托盘管理器
func NewSystemTray(ctx context.Context) *SystemTray {
	return &SystemTray{
		ctx: ctx,
	}
}

// ShowWindow 显示窗口
func (s *SystemTray) ShowWindow() {
	log.Info("System tray: Showing window")
	runtime.WindowShow(s.ctx)
	runtime.WindowUnminimise(s.ctx)
}

// HideWindow 隐藏窗口
func (s *SystemTray) HideWindow() {
	log.Info("System tray: Hiding window")
	runtime.WindowHide(s.ctx)
}

// QuitApp 退出应用
func (s *SystemTray) QuitApp() {
	log.Info("System tray: Quitting application")
	runtime.Quit(s.ctx)
}

// Setup 设置系统托盘
// 注意: Wails v2 的原生系统托盘支持有限
// 建议使用第三方库如 getlantern/systray 或在前端实现托盘图标
func (s *SystemTray) Setup() error {
	log.Info("System tray setup completed")
	log.Info("Note: Full system tray support requires additional libraries")
	log.Info("Consider using getlantern/systray or implementing in frontend")
	return nil
}
