package adapters

type GeminiAdapter struct{}

func (a *GeminiAdapter) AdaptRequest(request map[string]interface{}, targetModel string) (map[string]interface{}, error) {
	// Gemini 适配器 - 基础实现
	// TODO: 实现完整的 Gemini API 转换逻辑
	adapted := make(map[string]interface{})

	if targetModel != "" {
		adapted["model"] = targetModel
	} else {
		adapted["model"] = getOrDefault(request, "model", "gemini-pro")
	}

	// 转换消息格式为 Gemini 格式
	if messages, ok := request["messages"].([]interface{}); ok {
		adapted["contents"] = a.convertMessages(messages)
	}

	// 其他参数
	if temp, ok := request["temperature"]; ok {
		adapted["temperature"] = temp
	}
	if maxTokens, ok := request["max_tokens"]; ok {
		adapted["maxOutputTokens"] = maxTokens
	}

	return adapted, nil
}

func (a *GeminiAdapter) AdaptResponse(response map[string]interface{}) (map[string]interface{}, error) {
	// 将 Gemini 响应转换为 OpenAI 格式
	adapted := map[string]interface{}{
		"id":      "chatcmpl-gemini",
		"object":  "chat.completion",
		"created": 0,
		"model":   getOrDefault(response, "model", "gemini-pro"),
	}

	// 提取内容
	var contentText string
	if candidates, ok := response["candidates"].([]interface{}); ok && len(candidates) > 0 {
		if candidate, ok := candidates[0].(map[string]interface{}); ok {
			if content, ok := candidate["content"].(map[string]interface{}); ok {
				if parts, ok := content["parts"].([]interface{}); ok && len(parts) > 0 {
					if part, ok := parts[0].(map[string]interface{}); ok {
						contentText = getOrDefault(part, "text", "").(string)
					}
				}
			}
		}
	}

	adapted["choices"] = []map[string]interface{}{
		{
			"index": 0,
			"message": map[string]interface{}{
				"role":    "assistant",
				"content": contentText,
			},
			"finish_reason": "stop",
		},
	}

	adapted["usage"] = map[string]interface{}{
		"prompt_tokens":     0,
		"completion_tokens": 0,
		"total_tokens":      0,
	}

	return adapted, nil
}

func (a *GeminiAdapter) AdaptStreamChunk(chunk map[string]interface{}) (map[string]interface{}, error) {
	return chunk, nil // TODO: 实现流式响应转换
}

func (a *GeminiAdapter) convertMessages(messages []interface{}) []map[string]interface{} {
	contents := make([]map[string]interface{}, 0)

	for _, msg := range messages {
		if msgMap, ok := msg.(map[string]interface{}); ok {
			role := msgMap["role"].(string)
			content := msgMap["content"]

			// Gemini 使用 "user" 和 "model" 作为角色
			geminiRole := "user"
			if role == "assistant" {
				geminiRole = "model"
			}

			geminiMsg := map[string]interface{}{
				"role": geminiRole,
				"parts": []map[string]interface{}{
					{"text": content},
				},
			}
			contents = append(contents, geminiMsg)
		}
	}

	return contents
}
