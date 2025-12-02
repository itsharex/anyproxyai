# 系统托盘和开机自启动功能实现说明

## 实现概述

本次实现为 Wails 应用添加了以下功能：
1. **系统托盘支持** - 窗口最小化到托盘
2. **开机自启动** - Windows 注册表方式实现
3. **应用设置管理** - 前端可配置的托盘和自启动选项

## 文件变更清单

### 新增文件

1. **`internal/system/autostart.go`**
   - Windows 开机自启动管理
   - 使用注册表实现（`HKEY_CURRENT_USER\Software\Microsoft\Windows\CurrentVersion\Run`）
   - 提供的方法：
     - `EnableAutoStart()` - 启用开机自启动
     - `DisableAutoStart()` - 禁用开机自启动
     - `IsAutoStartEnabled()` - 检查是否已启用

2. **`internal/system/systray_windows.go`**
   - 系统托盘管理器
   - 提供窗口显示/隐藏控制
   - 为未来扩展预留接口

### 修改文件

1. **`internal/config/config.go`**
   - 添加字段：
     ```go
     MinimizeToTray bool `json:"minimize_to_tray"` // 默认: true
     AutoStart      bool `json:"auto_start"`       // 默认: false
     ```

2. **`main.go`**
   - 添加导入：
     ```go
     "openai-router-go/internal/system"
     "github.com/wailsapp/wails/v2/pkg/runtime"
     ```
   - 更新 `App` 结构体：
     ```go
     type App struct {
         ctx          context.Context
         routeService *service.RouteService
         proxyService *service.ProxyService
         config       *config.Config
         autoStart    *system.AutoStart
         systemTray   *system.SystemTray
     }
     ```
   - 添加钩子：
     - `OnBeforeClose` - 根据配置决定最小化或退出
   - 新增方法（可从前端调用）：
     - `GetAppSettings()` - 获取应用设置
     - `SetMinimizeToTray(enabled bool)` - 设置最小化到托盘
     - `SetAutoStart(enabled bool)` - 设置开机自启动
     - `ShowWindow()` - 显示窗口
     - `HideWindow()` - 隐藏窗口
     - `QuitApp()` - 退出应用

## 功能说明

### 1. 系统托盘功能

#### 工作原理
- 当用户关闭窗口时，如果 `MinimizeToTray` 为 `true`，窗口会隐藏而不是退出
- 通过 `OnBeforeClose` 钩子拦截关闭事件
- 使用 Wails runtime API 控制窗口可见性

#### 前端调用示例
```javascript
// 获取应用设置
const settings = await window.go.main.App.GetAppSettings();
console.log(settings.minimizeToTray); // true/false

// 设置最小化到托盘
await window.go.main.App.SetMinimizeToTray(true);

// 显示窗口
await window.go.main.App.ShowWindow();

// 隐藏窗口
await window.go.main.App.HideWindow();

// 退出应用
await window.go.main.App.QuitApp();
```

#### 注意事项
- Wails v2 的系统托盘支持有限，建议考虑以下方案：
  - 使用第三方库如 `getlantern/systray`
  - 在前端实现托盘图标（通过 Electron-like API）
  - 当前实现提供了基础的窗口控制功能

### 2. 开机自启动功能

#### 工作原理
- Windows: 通过修改注册表实现
  - 路径: `HKEY_CURRENT_USER\Software\Microsoft\Windows\CurrentVersion\Run`
  - 键名: `OpenAIRouter`
  - 键值: 应用程序的完整路径

#### 前端调用示例
```javascript
// 获取应用设置
const settings = await window.go.main.App.GetAppSettings();
console.log(settings.autoStart); // 配置文件中的设置
console.log(settings.autoStartEnabled); // 注册表中的实际状态

// 启用开机自启动
await window.go.main.App.SetAutoStart(true);

// 禁用开机自启动
await window.go.main.App.SetAutoStart(false);
```

#### 错误处理
- 所有方法都包含完善的错误处理
- 使用 logrus 记录详细日志
- 注册表操作失败会返回具体错误信息

### 3. 配置管理

#### 配置文件 (`config.json`)
```json
{
  "host": "localhost",
  "port": 8000,
  "database_path": "routes.db",
  "local_api_key": "sk-local-default-key",
  "redirect_enabled": false,
  "redirect_keyword": "proxy_auto",
  "redirect_target_model": "",
  "redirect_target_name": "",
  "minimize_to_tray": true,
  "auto_start": false
}
```

#### 配置说明
- `minimize_to_tray`: 关闭窗口时最小化到托盘（默认: true）
- `auto_start`: 是否启用开机自启动（默认: false）

## API 接口

### 应用设置相关

#### `GetAppSettings()`
获取应用设置

**返回值:**
```go
{
    "minimizeToTray": bool,      // 配置中的设置
    "autoStart": bool,           // 配置中的设置
    "autoStartEnabled": bool     // 注册表中的实际状态
}
```

#### `SetMinimizeToTray(enabled bool) error`
设置关闭时最小化到托盘

**参数:**
- `enabled`: true 启用，false 禁用

**返回值:**
- `error`: 错误信息（成功返回 nil）

#### `SetAutoStart(enabled bool) error`
设置开机自启动

**参数:**
- `enabled`: true 启用，false 禁用

**返回值:**
- `error`: 错误信息（成功返回 nil）

**说明:**
- 会同时更新注册表和配置文件
- 失败时不会回滚，建议前端处理错误后重试

### 窗口控制相关

#### `ShowWindow()`
显示并激活窗口

#### `HideWindow()`
隐藏窗口（不退出应用）

#### `QuitApp()`
强制退出应用（忽略 MinimizeToTray 设置）

## 依赖关系

### Go 依赖
- `golang.org/x/sys/windows/registry` - Windows 注册表操作（已包含在 go.mod 中）
- `github.com/wailsapp/wails/v2/pkg/runtime` - Wails 运行时 API
- `github.com/sirupsen/logrus` - 日志记录

### 版本要求
- Wails: v2.11.0+
- Go: 1.22.0+
- Windows: 支持所有 Windows 版本

## 日志记录

所有关键操作都会记录日志：
```
INFO[0000] Auto-start enabled: OpenAIRouter -> C:\path\to\app.exe
INFO[0001] Setting minimize to tray: true
INFO[0002] Minimize to tray setting updated successfully
INFO[0003] System tray setup completed
```

## 测试建议

### 1. 测试开机自启动
1. 调用 `SetAutoStart(true)`
2. 打开注册表编辑器（regedit）
3. 导航至 `HKEY_CURRENT_USER\Software\Microsoft\Windows\CurrentVersion\Run`
4. 验证 `OpenAIRouter` 键是否存在
5. 重启电脑，验证应用是否自动启动

### 2. 测试最小化到托盘
1. 调用 `SetMinimizeToTray(true)`
2. 点击窗口关闭按钮
3. 验证窗口是否隐藏而不是退出
4. 调用 `ShowWindow()` 验证窗口是否重新显示

### 3. 测试配置持久化
1. 修改设置
2. 重启应用
3. 验证设置是否保持

## 已知限制

1. **系统托盘图标**
   - Wails v2 不原生支持系统托盘图标
   - 建议使用第三方库或前端实现

2. **跨平台支持**
   - 当前仅实现 Windows 版本
   - macOS/Linux 需要不同的实现方式

3. **权限要求**
   - 修改注册表需要用户权限
   - UAC 可能会影响开机自启动

## 未来改进方向

1. **完整的系统托盘支持**
   - 集成 `getlantern/systray` 库
   - 添加托盘图标和右键菜单

2. **跨平台支持**
   - macOS: 使用 Launch Agents
   - Linux: 使用 systemd 或 .desktop 文件

3. **更多配置选项**
   - 最小化到托盘的行为（关闭按钮 vs 最小化按钮）
   - 启动时自动隐藏窗口
   - 托盘菜单自定义

## 故障排除

### 问题: 开机自启动不生效
1. 检查日志确认注册表操作成功
2. 验证注册表键值是否正确
3. 确认应用路径是否有效
4. 检查 UAC 设置

### 问题: 窗口隐藏后无法显示
1. 使用 `ShowWindow()` 方法
2. 检查 `a.ctx` 是否正确初始化
3. 确认 Wails runtime 正常工作

### 问题: 配置不保存
1. 检查文件写入权限
2. 验证 JSON 格式
3. 查看日志中的错误信息

## 总结

本次实现提供了完整的系统托盘和开机自启动基础功能，所有方法都可以从前端直接调用。虽然 Wails v2 的系统托盘支持有限，但提供的 API 已经足够实现基本的窗口控制和应用管理功能。
