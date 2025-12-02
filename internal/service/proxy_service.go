package service

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"openai-router-go/internal/adapters"
	"openai-router-go/internal/config"

	log "github.com/sirupsen/logrus"
)

type ProxyService struct {
	routeService *RouteService
	config       *config.Config
	httpClient   *http.Client
}

func NewProxyService(routeService *RouteService, cfg *config.Config) *ProxyService {
	return &ProxyService{
		routeService: routeService,
		config:       cfg,
		httpClient: &http.Client{
			Timeout: 0, // 不设置超时，因为大模型生成非常耗时
		},
	}
}

// ProxyRequest 代理请求
func (s *ProxyService) ProxyRequest(requestBody []byte, headers map[string]string) ([]byte, int, error) {
	// 解析请求
	var reqData map[string]interface{}
	if err := json.Unmarshal(requestBody, &reqData); err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("invalid JSON body: %v", err)
	}

	model, ok := reqData["model"].(string)
	if !ok || model == "" {
		return nil, http.StatusBadRequest, fmt.Errorf("'model' field is required")
	}

	log.Infof("Received request for model: %s", model)

	// 检查是否是重定向关键字
	if s.config.RedirectEnabled && model == s.config.RedirectKeyword {
		if s.config.RedirectTargetModel == "" {
			return nil, http.StatusNotFound, fmt.Errorf("redirect target model not configured")
		}
		log.Infof("Redirecting proxy_auto to model: %s", s.config.RedirectTargetModel)
		model = s.config.RedirectTargetModel
		reqData["model"] = model

		// 重新编码请求体
		requestBody, _ = json.Marshal(reqData)
	}

	// 查找路由
	route, err := s.routeService.GetRouteByModel(model)
	if err != nil {
		availableModels, _ := s.routeService.GetAvailableModels()
		return nil, http.StatusNotFound, fmt.Errorf("model '%s' not found in route list. Available models: %v", model, availableModels)
	}

	// 检查是否需要进行 API 转换
	var transformedBody []byte
	var targetURL string

	// 清理路由 API URL（移除末尾斜杠）
	cleanAPIUrl := strings.TrimSuffix(route.APIUrl, "/")

	// 判断是否需要使用适配器
	adapterName := s.detectAdapter(cleanAPIUrl, model)
	if adapterName != "" && s.config.RedirectEnabled && reqData["model"] == s.config.RedirectKeyword {
		// 使用适配器转换请求
		adapter := adapters.GetAdapter(adapterName)
		transformedReq, err := adapter.AdaptRequest(reqData, model)
		if err != nil {
			log.Errorf("Failed to adapt request: %v", err)
			return nil, http.StatusInternalServerError, err
		}
		transformedBody, _ = json.Marshal(transformedReq)
		targetURL = s.buildAdapterURL(cleanAPIUrl, adapterName)
	} else {
		// 不使用适配器，直接转发
		transformedBody = requestBody
		targetURL = cleanAPIUrl + "/v1/chat/completions"
	}

	log.Infof("Routing to: %s (route: %s)", targetURL, route.Name)

	// 创建代理请求
	proxyReq, err := http.NewRequest("POST", targetURL, bytes.NewReader(transformedBody))
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	// 设置请求头
	proxyReq.Header.Set("Content-Type", "application/json")

	// 使用路由配置的 API Key（如果有），否则透传原始 Authorization
	if route.APIKey != "" {
		proxyReq.Header.Set("Authorization", "Bearer "+route.APIKey)
	} else if auth := headers["Authorization"]; auth != "" {
		proxyReq.Header.Set("Authorization", auth)
	}

	// 发送请求
	startTime := time.Now()
	resp, err := s.httpClient.Do(proxyReq)
	if err != nil {
		s.routeService.LogRequest(model, route.ID, 0, 0, 0, false, err.Error())
		return nil, http.StatusServiceUnavailable, fmt.Errorf("backend service unavailable: %v", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		s.routeService.LogRequest(model, route.ID, 0, 0, 0, false, err.Error())
		return nil, http.StatusInternalServerError, err
	}

	log.Infof("Response received from %s in %v, status: %d", route.Name, time.Since(startTime), resp.StatusCode)

	// 记录使用情况（使用实际模型名而不是重定向关键字）
	if resp.StatusCode == http.StatusOK {
		var respData map[string]interface{}
		if err := json.Unmarshal(responseBody, &respData); err == nil {
			if usage, ok := respData["usage"].(map[string]interface{}); ok {
				totalTokens := int(usage["total_tokens"].(float64))
				promptTokens := int(usage["prompt_tokens"].(float64))
				completionTokens := int(usage["completion_tokens"].(float64))
				s.routeService.LogRequest(model, route.ID, promptTokens, completionTokens, totalTokens, true, "")
			}
		}
	} else {
		s.routeService.LogRequest(model, route.ID, 0, 0, 0, false, string(responseBody))
	}

	// 如果使用了适配器，转换响应
	if adapterName != "" && s.config.RedirectEnabled {
		adapter := adapters.GetAdapter(adapterName)
		var respData map[string]interface{}
		if err := json.Unmarshal(responseBody, &respData); err == nil {
			adaptedResp, err := adapter.AdaptResponse(respData)
			if err != nil {
				log.Errorf("Failed to adapt response: %v", err)
			} else {
				responseBody, _ = json.Marshal(adaptedResp)
			}
		}
	}

	return responseBody, resp.StatusCode, nil
}

// ProxyStreamRequest 代理流式请求
func (s *ProxyService) ProxyStreamRequest(requestBody []byte, headers map[string]string, writer io.Writer, flusher http.Flusher) error {
	// 解析请求
	var reqData map[string]interface{}
	if err := json.Unmarshal(requestBody, &reqData); err != nil {
		return fmt.Errorf("invalid JSON body: %v", err)
	}

	model, ok := reqData["model"].(string)
	if !ok || model == "" {
		return fmt.Errorf("'model' field is required")
	}

	originalModel := model

	// 检查是否是重定向关键字
	if s.config.RedirectEnabled && model == s.config.RedirectKeyword {
		if s.config.RedirectTargetModel == "" {
			return fmt.Errorf("redirect target model not configured")
		}
		model = s.config.RedirectTargetModel
		reqData["model"] = model
		requestBody, _ = json.Marshal(reqData)
	}

	// 查找路由
	route, err := s.routeService.GetRouteByModel(model)
	if err != nil {
		return err
	}

	// 清理路由 API URL（移除末尾斜杠）
	cleanAPIUrl := strings.TrimSuffix(route.APIUrl, "/")

	// 判断是否需要使用适配器
	var transformedBody []byte
	var targetURL string
	adapterName := s.detectAdapter(cleanAPIUrl, model)

	if adapterName != "" {
		// 使用适配器转换请求
		adapter := adapters.GetAdapter(adapterName)
		if adapter == nil {
			return fmt.Errorf("adapter not found: %s", adapterName)
		}

		// 确保开启stream
		reqData["stream"] = true
		transformedReq, err := adapter.AdaptRequest(reqData, model)
		if err != nil {
			log.Errorf("Failed to adapt request: %v", err)
			return err
		}
		transformedBody, _ = json.Marshal(transformedReq)
		targetURL = s.buildAdapterURL(cleanAPIUrl, adapterName)
		log.Infof("Streaming to: %s (route: %s, adapter: %s)", targetURL, route.Name, adapterName)
	} else {
		// 不使用适配器，直接转发
		transformedBody = requestBody
		targetURL = cleanAPIUrl + "/v1/chat/completions"
		log.Infof("Streaming to: %s (route: %s)", targetURL, route.Name)
	}

	// 创建代理请求
	proxyReq, err := http.NewRequest("POST", targetURL, bytes.NewReader(transformedBody))
	if err != nil {
		return err
	}

	proxyReq.Header.Set("Content-Type", "application/json")
	if route.APIKey != "" {
		proxyReq.Header.Set("Authorization", "Bearer "+route.APIKey)
	} else if auth := headers["Authorization"]; auth != "" {
		proxyReq.Header.Set("Authorization", auth)
	}

	// Claude需要特殊的版本头
	if adapterName == "anthropic" {
		proxyReq.Header.Set("anthropic-version", "2023-06-01")
	}

	// 发送请求
	resp, err := s.httpClient.Do(proxyReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("backend error: %d - %s", resp.StatusCode, string(body))
	}

	// 流式传输响应
	if adapterName != "" {
		// 需要转换SSE流
		return s.streamWithAdapter(resp.Body, writer, flusher, adapterName, originalModel, route.ID)
	} else {
		// 直接转发SSE流
		return s.streamDirect(resp.Body, writer, flusher, originalModel, route.ID)
	}
}

// streamWithAdapter 使用适配器处理流式响应
func (s *ProxyService) streamWithAdapter(reader io.Reader, writer io.Writer, flusher http.Flusher, adapterName, model string, routeID int64) error {
	adapter := adapters.GetAdapter(adapterName)
	if adapter == nil {
		return fmt.Errorf("adapter not found: %s", adapterName)
	}

	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 4096), 1024*1024) // 1MB max

	for scanner.Scan() {
		line := scanner.Text()

		// 跳过空行
		if line == "" {
			continue
		}

		// 处理SSE格式: "data: {...}"
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")

			// 检查是否是结束标记
			if data == "[DONE]" {
				fmt.Fprintf(writer, "data: [DONE]\n\n")
				flusher.Flush()
				s.routeService.LogRequest(model, routeID, 0, 0, 0, true, "")
				return nil
			}

			// 解析JSON
			var chunk map[string]interface{}
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				log.Warnf("Failed to parse chunk: %v, data: %s", err, data)
				continue
			}

			// 使用适配器转换chunk
			adaptedChunk, err := adapter.AdaptStreamChunk(chunk)
			if err != nil {
				log.Warnf("Failed to adapt chunk: %v", err)
				continue
			}

			// 发送转换后的chunk
			adaptedData, _ := json.Marshal(adaptedChunk)
			fmt.Fprintf(writer, "data: %s\n\n", string(adaptedData))
			flusher.Flush()
		}
	}

	if err := scanner.Err(); err != nil {
		s.routeService.LogRequest(model, routeID, 0, 0, 0, false, err.Error())
		return err
	}

	s.routeService.LogRequest(model, routeID, 0, 0, 0, true, "")
	return nil
}

// streamDirect 直接转发流式响应
func (s *ProxyService) streamDirect(reader io.Reader, writer io.Writer, flusher http.Flusher, model string, routeID int64) error {
	buf := make([]byte, 4096)
	for {
		n, err := reader.Read(buf)
		if n > 0 {
			if _, writeErr := writer.Write(buf[:n]); writeErr != nil {
				s.routeService.LogRequest(model, routeID, 0, 0, 0, false, writeErr.Error())
				return writeErr
			}
			flusher.Flush()
		}
		if err != nil {
			if err == io.EOF {
				s.routeService.LogRequest(model, routeID, 0, 0, 0, true, "")
				return nil
			}
			s.routeService.LogRequest(model, routeID, 0, 0, 0, false, err.Error())
			return err
		}
	}
}

// FetchRemoteModels 获取远程模型列表
func (s *ProxyService) FetchRemoteModels(apiUrl, apiKey string) ([]string, error) {
	// 移除末尾的斜杠
	apiUrl = strings.TrimSuffix(apiUrl, "/")

	// 添加 http/https 前缀（如果没有）
	if !strings.HasPrefix(apiUrl, "http://") && !strings.HasPrefix(apiUrl, "https://") {
		apiUrl = "https://" + apiUrl
	}

	url := apiUrl + "/v1/models"
	log.Infof("Fetching models from: %s", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %v (body: %s)", err, string(body))
	}

	models := make([]string, len(result.Data))
	for i, m := range result.Data {
		models[i] = m.ID
	}

	log.Infof("Successfully fetched %d models", len(models))
	return models, nil
}

// detectAdapter 检测需要使用的适配器
func (s *ProxyService) detectAdapter(apiUrl, model string) string {
	lowerURL := strings.ToLower(apiUrl)
	lowerModel := strings.ToLower(model)

	if strings.Contains(lowerURL, "anthropic") || strings.Contains(lowerModel, "claude") {
		return "anthropic"
	}
	if strings.Contains(lowerURL, "gemini") || strings.Contains(lowerModel, "gemini") {
		return "gemini"
	}
	if strings.Contains(lowerURL, "deepseek") || strings.Contains(lowerModel, "deepseek") {
		return "deepseek"
	}

	return "" // 不需要适配器
}

// buildAdapterURL 构建适配器URL
func (s *ProxyService) buildAdapterURL(baseURL, adapterName string) string {
	switch adapterName {
	case "anthropic":
		return baseURL + "/v1/messages"
	case "gemini":
		return baseURL + "/v1beta/models"
	default:
		return baseURL + "/v1/chat/completions"
	}
}
