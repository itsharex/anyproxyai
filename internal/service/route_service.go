package service

import (
	"database/sql"
	"fmt"
	"time"

	"openai-router-go/internal/database"

	log "github.com/sirupsen/logrus"
)

type RouteService struct {
	db *sql.DB
}

func NewRouteService(db *sql.DB) *RouteService {
	return &RouteService{db: db}
}

// GetAllRoutes 获取所有路由
func (s *RouteService) GetAllRoutes() ([]database.ModelRoute, error) {
	query := `SELECT id, name, model, api_url, api_key, "group", enabled, created_at, updated_at
	          FROM model_routes ORDER BY created_at DESC`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var routes []database.ModelRoute
	for rows.Next() {
		var route database.ModelRoute
		err := rows.Scan(&route.ID, &route.Name, &route.Model, &route.APIUrl, &route.APIKey,
			&route.Group, &route.Enabled, &route.CreatedAt, &route.UpdatedAt)
		if err != nil {
			return nil, err
		}
		routes = append(routes, route)
	}

	return routes, nil
}

// GetRouteByModel 根据模型名获取路由(支持负载均衡)
func (s *RouteService) GetRouteByModel(model string) (*database.ModelRoute, error) {
	query := `SELECT id, name, model, api_url, api_key, "group", enabled, created_at, updated_at
	          FROM model_routes WHERE model = ? AND enabled = 1 ORDER BY RANDOM() LIMIT 1`

	var route database.ModelRoute
	err := s.db.QueryRow(query, model).Scan(&route.ID, &route.Name, &route.Model, &route.APIUrl,
		&route.APIKey, &route.Group, &route.Enabled, &route.CreatedAt, &route.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("model not found: %s", model)
	}
	if err != nil {
		return nil, err
	}

	return &route, nil
}

// AddRoute 添加路由
func (s *RouteService) AddRoute(name, model, apiUrl, apiKey, group string) error {
	query := `INSERT INTO model_routes (name, model, api_url, api_key, "group", enabled, created_at, updated_at)
	          VALUES (?, ?, ?, ?, ?, 1, ?, ?)`

	now := time.Now()
	_, err := s.db.Exec(query, name, model, apiUrl, apiKey, group, now, now)
	if err != nil {
		log.Errorf("Failed to add route: %v", err)
		return err
	}

	log.Infof("Route added: %s -> %s (%s)", model, apiUrl, name)
	return nil
}

// UpdateRoute 更新路由
func (s *RouteService) UpdateRoute(id int64, name, model, apiUrl, apiKey, group string) error {
	query := `UPDATE model_routes SET name = ?, model = ?, api_url = ?, api_key = ?, "group" = ?, updated_at = ?
	          WHERE id = ?`

	result, err := s.db.Exec(query, name, model, apiUrl, apiKey, group, time.Now(), id)
	if err != nil {
		log.Errorf("Failed to update route: %v", err)
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("route not found: id=%d", id)
	}

	log.Infof("Route updated: id=%d", id)
	return nil
}

// DeleteRoute 删除路由
func (s *RouteService) DeleteRoute(id int64) error {
	query := `DELETE FROM model_routes WHERE id = ?`

	result, err := s.db.Exec(query, id)
	if err != nil {
		log.Errorf("Failed to delete route: %v", err)
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("route not found: id=%d", id)
	}

	log.Infof("Route deleted: id=%d", id)
	return nil
}

// ToggleRoute 启用/禁用路由
func (s *RouteService) ToggleRoute(id int64, enabled bool) error {
	query := `UPDATE model_routes SET enabled = ?, updated_at = ? WHERE id = ?`

	_, err := s.db.Exec(query, enabled, time.Now(), id)
	if err != nil {
		log.Errorf("Failed to toggle route: %v", err)
		return err
	}

	log.Infof("Route toggled: id=%d, enabled=%v", id, enabled)
	return nil
}

// GetStats 获取统计信息
func (s *RouteService) GetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	today := time.Now().Format("2006-01-02")

	// 路由总数
	var routeCount int
	err := s.db.QueryRow("SELECT COUNT(*) FROM model_routes WHERE enabled = 1").Scan(&routeCount)
	if err != nil {
		return nil, err
	}
	stats["route_count"] = routeCount

	// 模型总数（去重）
	var modelCount int
	err = s.db.QueryRow("SELECT COUNT(DISTINCT model) FROM model_routes WHERE enabled = 1").Scan(&modelCount)
	if err != nil {
		return nil, err
	}
	stats["model_count"] = modelCount

	// 总请求数
	var totalRequests int
	err = s.db.QueryRow("SELECT COUNT(*) FROM request_logs").Scan(&totalRequests)
	if err != nil {
		return nil, err
	}
	stats["total_requests"] = totalRequests

	// 总Token使用量
	var totalTokens int
	err = s.db.QueryRow("SELECT COALESCE(SUM(total_tokens), 0) FROM request_logs").Scan(&totalTokens)
	if err != nil {
		return nil, err
	}
	stats["total_tokens"] = totalTokens

	// 今日请求数
	var todayRequests int
	err = s.db.QueryRow("SELECT COUNT(*) FROM request_logs WHERE DATE(created_at) = ?", today).Scan(&todayRequests)
	if err != nil {
		return nil, err
	}
	stats["today_requests"] = todayRequests

	// 今日Token消耗
	var todayTokens int
	err = s.db.QueryRow("SELECT COALESCE(SUM(total_tokens), 0) FROM request_logs WHERE DATE(created_at) = ?", today).Scan(&todayTokens)
	if err != nil {
		return nil, err
	}
	stats["today_tokens"] = todayTokens

	// 成功率
	var successCount int
	err = s.db.QueryRow("SELECT COUNT(*) FROM request_logs WHERE success = 1").Scan(&successCount)
	if err != nil {
		return nil, err
	}

	successRate := 0.0
	if totalRequests > 0 {
		successRate = float64(successCount) / float64(totalRequests) * 100
	}
	stats["success_rate"] = successRate

	return stats, nil
}

// LogRequest 记录请求日志
func (s *RouteService) LogRequest(model string, routeID int64, requestTokens, responseTokens, totalTokens int, success bool, errorMsg string) error {
	query := `INSERT INTO request_logs (model, route_id, request_tokens, response_tokens, total_tokens, success, error_message, created_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := s.db.Exec(query, model, routeID, requestTokens, responseTokens, totalTokens, success, errorMsg, time.Now())
	return err
}

// GetAvailableModels 获取所有可用的模型列表
func (s *RouteService) GetAvailableModels() ([]string, error) {
	query := `SELECT DISTINCT model FROM model_routes WHERE enabled = 1 ORDER BY model`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var models []string
	for rows.Next() {
		var model string
		if err := rows.Scan(&model); err != nil {
			return nil, err
		}
		models = append(models, model)
	}

	return models, nil
}

// GetTodayStats 获取今日统计
func (s *RouteService) GetTodayStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	today := time.Now().Format("2006-01-02")

	// 今日请求数
	var todayRequests int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM request_logs WHERE DATE(created_at) = ?`, today).Scan(&todayRequests)
	if err != nil {
		return nil, err
	}
	stats["today_requests"] = todayRequests

	// 今日Token消耗
	var todayTokens int
	err = s.db.QueryRow(`SELECT COALESCE(SUM(total_tokens), 0) FROM request_logs WHERE DATE(created_at) = ?`, today).Scan(&todayTokens)
	if err != nil {
		return nil, err
	}
	stats["today_tokens"] = todayTokens

	return stats, nil
}

// GetDailyStats 获取每日统计（用于热力图）
func (s *RouteService) GetDailyStats(days int) ([]map[string]interface{}, error) {
	query := `
		SELECT
			DATE(created_at) as date,
			COUNT(*) as requests,
			COALESCE(SUM(request_tokens), 0) as request_tokens,
			COALESCE(SUM(response_tokens), 0) as response_tokens,
			COALESCE(SUM(total_tokens), 0) as total_tokens
		FROM request_logs
		WHERE created_at >= DATE('now', ?)
		GROUP BY DATE(created_at)
		ORDER BY date
	`

	rows, err := s.db.Query(query, fmt.Sprintf("-%d days", days))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []map[string]interface{}
	for rows.Next() {
		var date string
		var requests, requestTokens, responseTokens, totalTokens int
		err := rows.Scan(&date, &requests, &requestTokens, &responseTokens, &totalTokens)
		if err != nil {
			return nil, err
		}

		stats = append(stats, map[string]interface{}{
			"date":            date,
			"requests":        requests,
			"request_tokens":  requestTokens,
			"response_tokens": responseTokens,
			"total_tokens":    totalTokens,
		})
	}

	return stats, nil
}

// GetHourlyStats 获取今日按小时统计
func (s *RouteService) GetHourlyStats() ([]map[string]interface{}, error) {
	today := time.Now().Format("2006-01-02")
	query := `
		SELECT
			CAST(strftime('%H', created_at) AS INTEGER) as hour,
			COUNT(*) as requests,
			COALESCE(SUM(total_tokens), 0) as total_tokens
		FROM request_logs
		WHERE DATE(created_at) = ?
		GROUP BY hour
		ORDER BY hour
	`

	rows, err := s.db.Query(query, today)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []map[string]interface{}
	for rows.Next() {
		var hour, requests, totalTokens int
		err := rows.Scan(&hour, &requests, &totalTokens)
		if err != nil {
			return nil, err
		}

		stats = append(stats, map[string]interface{}{
			"hour":         hour,
			"requests":     requests,
			"total_tokens": totalTokens,
		})
	}

	return stats, nil
}

// GetModelRanking 获取模型使用排行
func (s *RouteService) GetModelRanking(limit int) ([]map[string]interface{}, error) {
	query := `
		SELECT
			model,
			COUNT(*) as requests,
			COALESCE(SUM(request_tokens), 0) as request_tokens,
			COALESCE(SUM(response_tokens), 0) as response_tokens,
			COALESCE(SUM(total_tokens), 0) as total_tokens,
			ROUND(AVG(CASE WHEN success = 1 THEN 100.0 ELSE 0.0 END), 2) as success_rate
		FROM request_logs
		GROUP BY model
		ORDER BY total_tokens DESC
		LIMIT ?
	`

	rows, err := s.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ranking []map[string]interface{}
	rank := 1
	for rows.Next() {
		var model string
		var requests, requestTokens, responseTokens, totalTokens int
		var successRate float64
		err := rows.Scan(&model, &requests, &requestTokens, &responseTokens, &totalTokens, &successRate)
		if err != nil {
			return nil, err
		}

		ranking = append(ranking, map[string]interface{}{
			"rank":            rank,
			"model":           model,
			"requests":        requests,
			"request_tokens":  requestTokens,
			"response_tokens": responseTokens,
			"total_tokens":    totalTokens,
			"success_rate":    successRate,
		})
		rank++
	}

	return ranking, nil
}
