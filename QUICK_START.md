# 快速开始指南

## 新功能快速测试

### 在前端调用新功能（JavaScript/TypeScript）

#### 1. 获取应用设置
```javascript
const settings = await window.go.main.App.GetAppSettings();
console.log(settings);
// 输出: { minimizeToTray: true, autoStart: false, autoStartEnabled: false }
```

#### 2. 启用开机自启动
```javascript
try {
  await window.go.main.App.SetAutoStart(true);
  alert('开机自启动已启用');
} catch (error) {
  alert('设置失败: ' + error);
}
```

#### 3. 启用最小化到托盘
```javascript
try {
  await window.go.main.App.SetMinimizeToTray(true);
  alert('最小化到托盘已启用');
} catch (error) {
  alert('设置失败: ' + error);
}
```

#### 4. 窗口控制
```javascript
// 隐藏窗口
await window.go.main.App.HideWindow();

// 显示窗口
await window.go.main.App.ShowWindow();

// 退出应用
await window.go.main.App.QuitApp();
```

## 简单的设置界面 HTML 示例

将此代码添加到你的前端页面：

```html
<!DOCTYPE html>
<html>
<head>
  <title>应用设置</title>
  <style>
    body { font-family: Arial, sans-serif; padding: 20px; }
    .setting { margin: 20px 0; padding: 15px; border: 1px solid #ddd; border-radius: 5px; }
    label { display: flex; align-items: center; cursor: pointer; }
    input[type="checkbox"] { margin-right: 10px; width: 20px; height: 20px; }
    button { padding: 10px 20px; margin: 5px; cursor: pointer; }
    .info { color: #666; font-size: 14px; margin-top: 5px; }
  </style>
</head>
<body>
  <h1>应用设置</h1>

  <!-- 最小化到托盘 -->
  <div class="setting">
    <label>
      <input type="checkbox" id="minimizeToTray">
      关闭窗口时最小化到托盘
    </label>
    <div class="info">启用后，点击关闭按钮会将窗口隐藏到系统托盘</div>
  </div>

  <!-- 开机自启动 -->
  <div class="setting">
    <label>
      <input type="checkbox" id="autoStart">
      开机自动启动
    </label>
    <div class="info">启用后，应用会在 Windows 启动时自动运行</div>
  </div>

  <!-- 窗口控制 -->
  <div class="setting">
    <h3>窗口控制</h3>
    <button onclick="hideWindow()">隐藏窗口</button>
    <button onclick="showWindow()">显示窗口</button>
    <button onclick="quitApp()" style="background-color: #dc3545; color: white; border: none;">
      退出应用
    </button>
  </div>

  <script>
    // 页面加载时获取设置
    async function loadSettings() {
      try {
        const settings = await window.go.main.App.GetAppSettings();
        document.getElementById('minimizeToTray').checked = settings.minimizeToTray;
        document.getElementById('autoStart').checked = settings.autoStart;
      } catch (error) {
        console.error('加载设置失败:', error);
      }
    }

    // 最小化到托盘切换
    document.getElementById('minimizeToTray').addEventListener('change', async (e) => {
      try {
        await window.go.main.App.SetMinimizeToTray(e.target.checked);
        alert('设置已保存');
      } catch (error) {
        alert('设置失败: ' + error);
        e.target.checked = !e.target.checked;
      }
    });

    // 开机自启动切换
    document.getElementById('autoStart').addEventListener('change', async (e) => {
      try {
        await window.go.main.App.SetAutoStart(e.target.checked);
        alert('设置已保存');
      } catch (error) {
        alert('设置失败: ' + error);
        e.target.checked = !e.target.checked;
      }
    });

    // 窗口控制函数
    async function hideWindow() {
      await window.go.main.App.HideWindow();
    }

    async function showWindow() {
      await window.go.main.App.ShowWindow();
    }

    async function quitApp() {
      if (confirm('确定要退出应用吗？')) {
        await window.go.main.App.QuitApp();
      }
    }

    // 页面加载时初始化
    window.addEventListener('DOMContentLoaded', loadSettings);
  </script>
</body>
</html>
```

## 命令行测试（开发者）

### 编译应用
```bash
cd C:\Users\Administrator\Desktop\codingfile\openaichage\openai-router-go
wails build
```

### 运行开发模式
```bash
wails dev
```

### 检查编译
```bash
go build -v
```

### 查看日志
运行应用后，所有操作都会记录在控制台日志中：
```
INFO[0000] Setting auto-start: true
INFO[0001] Auto-start enabled: OpenAIRouter -> C:\...\app.exe
INFO[0002] Auto-start setting updated successfully
```

## 手动验证（Windows）

### 验证开机自启动是否生效

1. 按 `Win + R` 打开运行对话框
2. 输入 `regedit` 并回车
3. 导航到：`HKEY_CURRENT_USER\Software\Microsoft\Windows\CurrentVersion\Run`
4. 查找 `OpenAIRouter` 键，值应该是应用程序的完整路径

### 验证配置文件

配置文件位置：`config.json`

查看内容：
```json
{
  "minimize_to_tray": true,
  "auto_start": false
}
```

## 常见问题

### Q: 点击关闭按钮后应用不见了怎么办？
A: 这是正常的，如果启用了"最小化到托盘"，窗口会隐藏。目前可以通过以下方式恢复：
- 在任务栏找到应用图标并点击
- 或者在前端添加一个托盘图标，点击后调用 `ShowWindow()`

### Q: 如何完全退出应用？
A: 有两种方式：
1. 禁用"最小化到托盘"后再关闭窗口
2. 调用 `QuitApp()` 方法强制退出

### Q: 开机自启动不生效？
A: 检查以下几点：
1. 是否有管理员权限
2. 注册表中的路径是否正确
3. 查看应用日志是否有错误信息

### Q: 如何在前端添加系统托盘图标？
A: Wails v2 的系统托盘支持有限，建议：
1. 使用第三方库如 `getlantern/systray`
2. 或者只使用窗口显示/隐藏功能

## 完整示例项目

查看以下文档获取更多信息：

- **`IMPLEMENTATION_NOTES.md`** - 详细技术实现
- **`FRONTEND_INTEGRATION.md`** - React/Vue 完整示例
- **`IMPLEMENTATION_SUMMARY.md`** - 功能总结

## API 速查表

| 方法 | 参数 | 返回值 | 说明 |
|------|------|--------|------|
| `GetAppSettings()` | 无 | `{minimizeToTray, autoStart, autoStartEnabled}` | 获取应用设置 |
| `SetMinimizeToTray(enabled)` | `bool` | `error` | 设置最小化到托盘 |
| `SetAutoStart(enabled)` | `bool` | `error` | 设置开机自启动 |
| `ShowWindow()` | 无 | 无 | 显示窗口 |
| `HideWindow()` | 无 | 无 | 隐藏窗口 |
| `QuitApp()` | 无 | 无 | 退出应用 |

## 立即开始

1. 编译应用：`wails build`
2. 运行应用
3. 在浏览器控制台测试 API：
   ```javascript
   await window.go.main.App.GetAppSettings()
   ```
4. 集成到你的前端界面

祝使用愉快！
