package adapters

import (
	"fmt"
	"time"
)

// ClaudeToOpenAIAdapter 将 Claude (Anthropic) 格式转换为 OpenAI 格式
type ClaudeToOpenAIAdapter struct{}

func init() {
	RegisterAdapter("claude-to-openai", &ClaudeToOpenAIAdapter{})
}

// AdaptRequest 将 Claude 请求转换为 OpenAI 请求
func (a *ClaudeToOpenAIAdapter) AdaptRequest(reqData map[string]interface{}, model string) (map[string]interface{}, error) {
	// Claude 请求格式示例:
	// {
	//   "model": "claude-3-haiku-20240307",
	//   "max_tokens": 1000,
	//   "messages": [
	//     {"role": "user", "content": "Hello"}
	//   ]
	// }
	//
	// OpenAI 请求格式示例:
	// {
	//   "model": "gpt-3.5-turbo",
	//   "messages": [
	//     {"role": "user", "content": "Hello"}
	//   ],
	//   "max_tokens": 1000
	// }

	openaiReq := make(map[string]interface{})

	// 复制模型名
	openaiReq["model"] = model

	// 处理 system 参数 - Claude 支持单独的 system 字段，OpenAI 需要作为第一条消息
	var systemMessage string
	if system, ok := reqData["system"].(string); ok && system != "" {
		systemMessage = system
	}

	// 转换消息格式 - Claude 和 OpenAI 的消息格式基本相同
	if messages, ok := reqData["messages"].([]interface{}); ok {
		openaiMessages := make([]interface{}, 0, len(messages)+1)

		// 如果有 system 消息，添加为第一条消息
		if systemMessage != "" {
			openaiMessages = append(openaiMessages, map[string]interface{}{
				"role":    "system",
				"content": systemMessage,
			})
		}

		for _, msg := range messages {
			if msgMap, ok := msg.(map[string]interface{}); ok {
				openaiMsg := make(map[string]interface{})

				// 复制 role 和 content
				if role, ok := msgMap["role"].(string); ok {
					openaiMsg["role"] = role
				}

				// 处理 content - 可能是字符串或数组
				if content, ok := msgMap["content"]; ok {
					switch v := content.(type) {
					case string:
						// 简单的文本内容
						openaiMsg["content"] = v
					case []interface{}:
						// 多模态内容 - Claude 格式
						// [{"type": "text", "text": "..."}]
						// 转换为 OpenAI 格式的文本
						var textContent string
						for _, part := range v {
							if partMap, ok := part.(map[string]interface{}); ok {
								if partMap["type"] == "text" {
									if text, ok := partMap["text"].(string); ok {
										textContent += text
									}
								}
							}
						}
						openaiMsg["content"] = textContent
					default:
						openaiMsg["content"] = fmt.Sprintf("%v", v)
					}
				}

				openaiMessages = append(openaiMessages, openaiMsg)
			}
		}

		openaiReq["messages"] = openaiMessages
	}

	// 转换其他参数
	if maxTokens, ok := reqData["max_tokens"]; ok {
		openaiReq["max_tokens"] = maxTokens
	}

	if temperature, ok := reqData["temperature"]; ok {
		openaiReq["temperature"] = temperature
	}

	if topP, ok := reqData["top_p"]; ok {
		openaiReq["top_p"] = topP
	}

	// 处理流式参数
	if stream, ok := reqData["stream"]; ok {
		openaiReq["stream"] = stream
	}

	// 处理 stop sequences
	if stopSequences, ok := reqData["stop_sequences"]; ok {
		openaiReq["stop"] = stopSequences
	}

	// 注意：不要转发以下 Claude 特有的字段，因为 OpenAI API 不支持：
	// - metadata (Claude 特有)
	// - anthropic_version (Claude 特有)
	// - system (应该合并到 messages 中)

	return openaiReq, nil
}

// AdaptResponse 将 OpenAI 响应转换为 Claude 响应（不需要，因为 Claude 接口直接返回 OpenAI 响应）
func (a *ClaudeToOpenAIAdapter) AdaptResponse(respData map[string]interface{}) (map[string]interface{}, error) {
	// 这个适配器主要用于请求转换，响应不需要转换
	// 因为 /api/anthropic 接口返回的是 OpenAI 格式的响应
	return respData, nil
}

// AdaptStreamChunk 转换流式响应块 - Claude SSE → OpenAI SSE
func (a *ClaudeToOpenAIAdapter) AdaptStreamChunk(chunk map[string]interface{}) (map[string]interface{}, error) {
	chunkType, _ := chunk["type"].(string)

	switch chunkType {
	case "message_start":
		// 跳过 message_start 事件，OpenAI 不需要
		return nil, nil

	case "content_block_start":
		// 跳过 content_block_start 事件
		return nil, nil

	case "content_block_delta":
		// 提取文本内容并转换为 OpenAI 格式
		if delta, ok := chunk["delta"].(map[string]interface{}); ok {
			deltaType, _ := delta["type"].(string)
			if deltaType == "text_delta" {
				if text, ok := delta["text"].(string); ok {
					// 构建 OpenAI 格式的流式响应
					return map[string]interface{}{
						"id":      "chatcmpl-" + fmt.Sprintf("%d", time.Now().UnixNano()),
						"object":  "chat.completion.chunk",
						"created": time.Now().Unix(),
						"model":   "claude",
						"choices": []interface{}{
							map[string]interface{}{
								"index": 0,
								"delta": map[string]interface{}{
									"content": text,
								},
								"finish_reason": nil,
							},
						},
					}, nil
				}
			}
		}
		return nil, nil

	case "content_block_stop":
		// 跳过 content_block_stop 事件
		return nil, nil

	case "message_delta":
		// 提取 finish_reason 并发送最终的 chunk
		if delta, ok := chunk["delta"].(map[string]interface{}); ok {
			stopReason, _ := delta["stop_reason"].(string)

			// 转换 stop_reason: end_turn → stop, max_tokens → length
			openaiStopReason := "stop"
			if stopReason == "max_tokens" {
				openaiStopReason = "length"
			}

			return map[string]interface{}{
				"id":      "chatcmpl-" + fmt.Sprintf("%d", time.Now().UnixNano()),
				"object":  "chat.completion.chunk",
				"created": time.Now().Unix(),
				"model":   "claude",
				"choices": []interface{}{
					map[string]interface{}{
						"index":         0,
						"delta":         map[string]interface{}{},
						"finish_reason": openaiStopReason,
					},
				},
			}, nil
		}
		return nil, nil

	case "message_stop":
		// 已经在 message_delta 中处理了 finish_reason，跳过
		return nil, nil

	default:
		// 未知类型，跳过
		return nil, nil
	}
}

// AdaptStreamStart 流式响应开始
func (a *ClaudeToOpenAIAdapter) AdaptStreamStart(model string) []map[string]interface{} {
	// 不需要额外的开始消息
	return nil
}

// AdaptStreamEnd 流式响应结束
func (a *ClaudeToOpenAIAdapter) AdaptStreamEnd() []map[string]interface{} {
	// 不需要额外的结束消息
	return nil
}
