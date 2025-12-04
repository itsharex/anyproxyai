# AnyProxyAi - GUI 管理工具

Golang + Wails 桌面应用，OpenAI API 路由管理。

## 快速开始

**Windows 双击运行**：`build\bin\anyproxyai.exe`

或使用：`start.bat`

## 功能

- GUI 管理界面（Vue3 + Naive UI 深色主题）
- 路由管理（添加/编辑/删除）
- proxy_auto 智能重定向
- 远程模型列表获取
- 统计信息
- 嵌入式 API 服务器（localhost:8000）
- API 适配器（Claude / DeepSeek）

## 构建

```bash
wails build      # 当前平台
wails dev        # 开发模式
```

## API 使用

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:8000/api/v1",
    api_key="sk-local-default-key"
)

response = client.chat.completions.create(
    model="proxy_auto",
    messages=[{"role": "user", "content": "Hello!"}]
)
```

## 技术栈

- 后端：Go + Gin + SQLite
- 前端：Vue 3 + Naive UI
- 桌面：Wails v2

## GitHub

https://github.com/cniu6/anyproxyai
