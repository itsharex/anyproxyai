package database

import (
	"database/sql"
	"time"

	log "github.com/sirupsen/logrus"
	_ "modernc.org/sqlite"
)

// ModelRoute 模型路由表结构
type ModelRoute struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Model     string    `json:"model"`
	APIUrl    string    `json:"api_url"`
	APIKey    string    `json:"api_key"`
	Group     string    `json:"group"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// RequestLog 请求日志表结构
type RequestLog struct {
	ID             int64     `json:"id"`
	Model          string    `json:"model"`
	RouteID        int64     `json:"route_id"`
	RequestTokens  int       `json:"request_tokens"`
	ResponseTokens int       `json:"response_tokens"`
	TotalTokens    int       `json:"total_tokens"`
	Success        bool      `json:"success"`
	ErrorMessage   string    `json:"error_message"`
	CreatedAt      time.Time `json:"created_at"`
}

func InitDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	// 创建表
	if err := createTables(db); err != nil {
		db.Close()
		return nil, err
	}

	log.Info("Database initialized successfully")
	return db, nil
}

func createTables(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS model_routes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		model TEXT NOT NULL,
		api_url TEXT NOT NULL,
		api_key TEXT,
		"group" TEXT,
		enabled INTEGER DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_model_routes_model ON model_routes(model);
	CREATE INDEX IF NOT EXISTS idx_model_routes_enabled ON model_routes(enabled);
	CREATE INDEX IF NOT EXISTS idx_model_routes_group ON model_routes("group");

	CREATE TABLE IF NOT EXISTS request_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		model TEXT NOT NULL,
		route_id INTEGER,
		request_tokens INTEGER DEFAULT 0,
		response_tokens INTEGER DEFAULT 0,
		total_tokens INTEGER DEFAULT 0,
		success INTEGER DEFAULT 1,
		error_message TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (route_id) REFERENCES model_routes(id) ON DELETE SET NULL
	);

	CREATE INDEX IF NOT EXISTS idx_request_logs_model ON request_logs(model);
	CREATE INDEX IF NOT EXISTS idx_request_logs_route_id ON request_logs(route_id);
	CREATE INDEX IF NOT EXISTS idx_request_logs_created_at ON request_logs(created_at);
	CREATE INDEX IF NOT EXISTS idx_request_logs_success ON request_logs(success);
	`

	_, err := db.Exec(schema)
	return err
}
