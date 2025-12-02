# 前端集成指南

## 系统托盘和开机自启动功能

本文档介绍如何在前端调用新增的系统托盘和开机自启动 API。

## 可用的 API 方法

### 1. 获取应用设置

```javascript
// 获取应用设置
const settings = await window.go.main.App.GetAppSettings();

// 返回格式:
// {
//   minimizeToTray: true,      // 是否启用最小化到托盘
//   autoStart: false,          // 配置文件中的自启动设置
//   autoStartEnabled: false    // 注册表中的实际自启动状态
// }
```

### 2. 设置最小化到托盘

```javascript
// 启用最小化到托盘
try {
  await window.go.main.App.SetMinimizeToTray(true);
  console.log('最小化到托盘已启用');
} catch (error) {
  console.error('设置失败:', error);
}

// 禁用最小化到托盘
try {
  await window.go.main.App.SetMinimizeToTray(false);
  console.log('最小化到托盘已禁用');
} catch (error) {
  console.error('设置失败:', error);
}
```

### 3. 设置开机自启动

```javascript
// 启用开机自启动
try {
  await window.go.main.App.SetAutoStart(true);
  console.log('开机自启动已启用');
} catch (error) {
  console.error('设置失败:', error);
}

// 禁用开机自启动
try {
  await window.go.main.App.SetAutoStart(false);
  console.log('开机自启动已禁用');
} catch (error) {
  console.error('设置失败:', error);
}
```

### 4. 窗口控制

```javascript
// 显示窗口
await window.go.main.App.ShowWindow();

// 隐藏窗口（不退出应用）
await window.go.main.App.HideWindow();

// 退出应用
await window.go.main.App.QuitApp();
```

## React 组件示例

### 设置页面组件

```jsx
import React, { useState, useEffect } from 'react';

function SettingsPage() {
  const [settings, setSettings] = useState({
    minimizeToTray: true,
    autoStart: false,
    autoStartEnabled: false
  });
  const [loading, setLoading] = useState(false);

  // 加载设置
  useEffect(() => {
    loadSettings();
  }, []);

  const loadSettings = async () => {
    try {
      const data = await window.go.main.App.GetAppSettings();
      setSettings(data);
    } catch (error) {
      console.error('加载设置失败:', error);
    }
  };

  // 切换最小化到托盘
  const handleMinimizeToTrayChange = async (enabled) => {
    setLoading(true);
    try {
      await window.go.main.App.SetMinimizeToTray(enabled);
      setSettings(prev => ({ ...prev, minimizeToTray: enabled }));
      alert('设置已保存');
    } catch (error) {
      alert('设置失败: ' + error);
    } finally {
      setLoading(false);
    }
  };

  // 切换开机自启动
  const handleAutoStartChange = async (enabled) => {
    setLoading(true);
    try {
      await window.go.main.App.SetAutoStart(enabled);
      await loadSettings(); // 重新加载以获取实际状态
      alert('设置已保存');
    } catch (error) {
      alert('设置失败: ' + error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="settings-page">
      <h2>应用设置</h2>

      {/* 最小化到托盘 */}
      <div className="setting-item">
        <label>
          <input
            type="checkbox"
            checked={settings.minimizeToTray}
            onChange={(e) => handleMinimizeToTrayChange(e.target.checked)}
            disabled={loading}
          />
          关闭窗口时最小化到托盘
        </label>
        <p className="setting-description">
          启用后，点击关闭按钮会将窗口隐藏到系统托盘，而不是退出应用
        </p>
      </div>

      {/* 开机自启动 */}
      <div className="setting-item">
        <label>
          <input
            type="checkbox"
            checked={settings.autoStart}
            onChange={(e) => handleAutoStartChange(e.target.checked)}
            disabled={loading}
          />
          开机自动启动
        </label>
        <p className="setting-description">
          启用后，应用会在 Windows 启动时自动运行
        </p>
        {settings.autoStart !== settings.autoStartEnabled && (
          <p className="warning">
            注意: 配置与实际状态不一致，请重新设置
          </p>
        )}
      </div>

      {/* 窗口控制 */}
      <div className="window-controls">
        <h3>窗口控制</h3>
        <button onClick={() => window.go.main.App.HideWindow()}>
          隐藏窗口
        </button>
        <button onClick={() => window.go.main.App.ShowWindow()}>
          显示窗口
        </button>
        <button
          onClick={() => {
            if (confirm('确定要退出应用吗？')) {
              window.go.main.App.QuitApp();
            }
          }}
          className="danger"
        >
          退出应用
        </button>
      </div>
    </div>
  );
}

export default SettingsPage;
```

### 样式示例 (CSS)

```css
.settings-page {
  padding: 20px;
  max-width: 600px;
  margin: 0 auto;
}

.setting-item {
  margin-bottom: 24px;
  padding: 16px;
  border: 1px solid #e0e0e0;
  border-radius: 8px;
}

.setting-item label {
  display: flex;
  align-items: center;
  font-size: 16px;
  font-weight: 500;
  cursor: pointer;
}

.setting-item input[type="checkbox"] {
  margin-right: 12px;
  width: 20px;
  height: 20px;
  cursor: pointer;
}

.setting-description {
  margin-top: 8px;
  font-size: 14px;
  color: #666;
}

.warning {
  margin-top: 8px;
  padding: 8px;
  background-color: #fff3cd;
  border: 1px solid #ffc107;
  border-radius: 4px;
  color: #856404;
  font-size: 14px;
}

.window-controls {
  margin-top: 32px;
  padding: 16px;
  background-color: #f5f5f5;
  border-radius: 8px;
}

.window-controls h3 {
  margin-top: 0;
  margin-bottom: 16px;
}

.window-controls button {
  margin-right: 12px;
  padding: 8px 16px;
  border: none;
  border-radius: 4px;
  background-color: #007bff;
  color: white;
  cursor: pointer;
  font-size: 14px;
}

.window-controls button:hover {
  background-color: #0056b3;
}

.window-controls button:disabled {
  background-color: #ccc;
  cursor: not-allowed;
}

.window-controls button.danger {
  background-color: #dc3545;
}

.window-controls button.danger:hover {
  background-color: #c82333;
}
```

## Vue 组件示例

```vue
<template>
  <div class="settings-page">
    <h2>应用设置</h2>

    <!-- 最小化到托盘 -->
    <div class="setting-item">
      <label>
        <input
          type="checkbox"
          v-model="settings.minimizeToTray"
          @change="handleMinimizeToTrayChange"
          :disabled="loading"
        />
        关闭窗口时最小化到托盘
      </label>
      <p class="setting-description">
        启用后，点击关闭按钮会将窗口隐藏到系统托盘，而不是退出应用
      </p>
    </div>

    <!-- 开机自启动 -->
    <div class="setting-item">
      <label>
        <input
          type="checkbox"
          v-model="settings.autoStart"
          @change="handleAutoStartChange"
          :disabled="loading"
        />
        开机自动启动
      </label>
      <p class="setting-description">
        启用后，应用会在 Windows 启动时自动运行
      </p>
    </div>

    <!-- 窗口控制 -->
    <div class="window-controls">
      <h3>窗口控制</h3>
      <button @click="hideWindow">隐藏窗口</button>
      <button @click="showWindow">显示窗口</button>
      <button @click="quitApp" class="danger">退出应用</button>
    </div>
  </div>
</template>

<script>
export default {
  name: 'SettingsPage',
  data() {
    return {
      settings: {
        minimizeToTray: true,
        autoStart: false,
        autoStartEnabled: false
      },
      loading: false
    };
  },
  mounted() {
    this.loadSettings();
  },
  methods: {
    async loadSettings() {
      try {
        const data = await window.go.main.App.GetAppSettings();
        this.settings = data;
      } catch (error) {
        console.error('加载设置失败:', error);
      }
    },
    async handleMinimizeToTrayChange() {
      this.loading = true;
      try {
        await window.go.main.App.SetMinimizeToTray(this.settings.minimizeToTray);
        this.$message.success('设置已保存');
      } catch (error) {
        this.$message.error('设置失败: ' + error);
        await this.loadSettings();
      } finally {
        this.loading = false;
      }
    },
    async handleAutoStartChange() {
      this.loading = true;
      try {
        await window.go.main.App.SetAutoStart(this.settings.autoStart);
        await this.loadSettings();
        this.$message.success('设置已保存');
      } catch (error) {
        this.$message.error('设置失败: ' + error);
        await this.loadSettings();
      } finally {
        this.loading = false;
      }
    },
    async hideWindow() {
      await window.go.main.App.HideWindow();
    },
    async showWindow() {
      await window.go.main.App.ShowWindow();
    },
    async quitApp() {
      if (confirm('确定要退出应用吗？')) {
        await window.go.main.App.QuitApp();
      }
    }
  }
};
</script>
```

## 注意事项

### 1. 错误处理
所有 API 调用都应该包含 try-catch 错误处理：

```javascript
try {
  await window.go.main.App.SetAutoStart(true);
} catch (error) {
  // 处理错误
  console.error('操作失败:', error);
  // 显示用户友好的错误消息
  alert('设置失败，请重试');
}
```

### 2. 状态同步
设置更改后，建议重新加载设置以确保状态同步：

```javascript
await window.go.main.App.SetAutoStart(true);
await loadSettings(); // 重新加载设置
```

### 3. 配置持久化
- 所有设置都会自动保存到 `config.json` 文件
- 应用重启后设置会自动加载
- 开机自启动会同时更新注册表和配置文件

### 4. 权限要求
- 修改注册表需要用户权限
- 如果遇到权限问题，可能需要以管理员身份运行应用

## 测试清单

- [ ] 设置页面正常加载
- [ ] 可以切换最小化到托盘选项
- [ ] 可以切换开机自启动选项
- [ ] 关闭窗口时根据设置执行相应操作
- [ ] 窗口显示/隐藏功能正常
- [ ] 退出功能正常工作
- [ ] 设置在重启后保持
- [ ] 错误情况有友好提示

## 调试技巧

### 查看后端日志
后端所有操作都会记录详细日志，可以帮助诊断问题：

```
INFO[0000] Setting minimize to tray: true
INFO[0001] Minimize to tray setting updated successfully
INFO[0002] Setting auto-start: true
INFO[0003] Auto-start enabled: OpenAIRouter -> C:\path\to\app.exe
```

### 检查注册表
手动检查注册表确认自启动是否设置成功：
1. 按 Win+R 运行 `regedit`
2. 导航至 `HKEY_CURRENT_USER\Software\Microsoft\Windows\CurrentVersion\Run`
3. 查找 `OpenAIRouter` 键

### 测试窗口行为
```javascript
// 测试窗口控制
console.log('测试隐藏窗口');
await window.go.main.App.HideWindow();

setTimeout(async () => {
  console.log('测试显示窗口');
  await window.go.main.App.ShowWindow();
}, 3000);
```

## 相关文档

- [IMPLEMENTATION_NOTES.md](./IMPLEMENTATION_NOTES.md) - 详细的实现说明
- [Wails v2 文档](https://wails.io/docs/reference/runtime/intro) - Wails 运行时 API
