package adapters

type DeepSeekAdapter struct{}

func (a *DeepSeekAdapter) AdaptRequest(request map[string]interface{}, targetModel string) (map[string]interface{}, error) {
	// DeepSeek 使用 OpenAI 兼容格式，基本不需要转换
	adapted := make(map[string]interface{})

	for k, v := range request {
		adapted[k] = v
	}

	if targetModel != "" {
		adapted["model"] = targetModel
	}

	return adapted, nil
}

func (a *DeepSeekAdapter) AdaptResponse(response map[string]interface{}) (map[string]interface{}, error) {
	// DeepSeek 响应格式与 OpenAI 兼容，直接返回
	return response, nil
}

func (a *DeepSeekAdapter) AdaptStreamChunk(chunk map[string]interface{}) (map[string]interface{}, error) {
	return chunk, nil
}
