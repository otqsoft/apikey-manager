// Package initialize 启动初始化：自动建表 + 首次启动种子数据。
// DDL 同时给出 SQLite（纯 Go 驱动）与 MySQL 两套方言，按 config.yaml 中 driver 选择。
package initialize

import (
	"context"
	"os"

	"github.com/gogf/gf/v2/frame/g"
)

// driverName 读取当前数据库类型（GoFrame gdb 配置键 type）
func driverName(ctx context.Context) string {
	d := g.Cfg().MustGet(ctx, "database.default.type", "sqlite").String()
	if d == "" {
		return "sqlite"
	}
	return d
}

// CreateTables 按契约第 1 节创建全部表与索引（IF NOT EXISTS，可重复执行）。
func CreateTables(ctx context.Context) error {
	// SQLite 需要 data 目录预先存在，避免文件创建失败
	if driverName(ctx) == "sqlite" {
		if err := os.MkdirAll("./data", 0o755); err != nil {
			return err
		}
	}
	for _, sql := range ddlMap[driverName(ctx)] {
		if _, err := g.DB().Exec(ctx, sql); err != nil {
			g.Log().Errorf(ctx, "建表 SQL 执行失败: %s, err=%v", sql, err)
			return err
		}
	}
	// 旧版本库表结构迁移：幂等补齐 v3 新增列（契约 7.2）
	if err := migrateChatMessagesReasoning(ctx); err != nil {
		return err
	}
	g.Log().Info(ctx, "数据库表结构初始化完成")
	return nil
}

// migrateChatMessagesReasoning 幂等补列：仅当 chat_messages 缺少 reasoning_content 时执行 ALTER ADD COLUMN，
// 已存在则跳过，兼容老库平滑升级。SQLite 用 PRAGMA table_info 探测；MySQL 查 information_schema.COLUMNS。
func migrateChatMessagesReasoning(ctx context.Context) error {
	const column = "reasoning_content"
	exists := false

	if driverName(ctx) == "mysql" {
		cnt, err := g.DB().Model("information_schema.columns").
			Where("table_schema = DATABASE()").
			Where("table_name", "chat_messages").
			Where("column_name", column).
			Count()
		if err != nil {
			return err
		}
		exists = cnt > 0
	} else {
		// SQLite：PRAGMA table_info 返回 name/type 等列，逐行比对
		records, err := g.DB().GetAll(ctx, "PRAGMA table_info(chat_messages)")
		if err != nil {
			return err
		}
		for _, r := range records {
			if r["name"].String() == column {
				exists = true
				break
			}
		}
	}

	if exists {
		return nil
	}
	var alterSQL string
	if driverName(ctx) == "mysql" {
		// MySQL TEXT/BLOB 不支持字面默认值，使用 VARCHAR(10000) 并给默认空串
		alterSQL = "ALTER TABLE chat_messages ADD COLUMN reasoning_content VARCHAR(10000) NOT NULL DEFAULT ''"
	} else {
		alterSQL = "ALTER TABLE chat_messages ADD COLUMN reasoning_content TEXT NOT NULL DEFAULT ''"
	}
	if _, err := g.DB().Exec(ctx, alterSQL); err != nil {
		g.Log().Errorf(ctx, "旧库补列失败: %s, err=%v", alterSQL, err)
		return err
	}
	g.Log().Info(ctx, "已为旧库 chat_messages 补充 reasoning_content 列")
	return nil
}

// ddlMap 两套方言 DDL；类型选择遵循“SQLite 宽松类型 + MySQL 精确类型”原则。
var ddlMap = map[string][]string{
	"sqlite": {
		`CREATE TABLE IF NOT EXISTS providers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			api_format TEXT NOT NULL DEFAULT 'openai',
			base_url TEXT NOT NULL,
			platform_url TEXT DEFAULT '',
			balance REAL NOT NULL DEFAULT 0,
			max_concurrency INTEGER NOT NULL DEFAULT 10,
			timeout_seconds INTEGER NOT NULL DEFAULT 60,
			status INTEGER NOT NULL DEFAULT 1,
			remark TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS models (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			provider_id INTEGER NOT NULL,
			input_price_cached REAL NOT NULL DEFAULT 0,
			input_price REAL NOT NULL DEFAULT 0,
			output_price REAL NOT NULL DEFAULT 0,
			context_length INTEGER NOT NULL DEFAULT 0,
			support_thinking INTEGER NOT NULL DEFAULT 0,
			support_multimodal INTEGER NOT NULL DEFAULT 0,
			status INTEGER NOT NULL DEFAULT 1,
			remark TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_models_provider ON models(provider_id)`,
		`CREATE TABLE IF NOT EXISTS api_keys (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			provider_id INTEGER NOT NULL,
			api_key TEXT NOT NULL,
			expire_date TEXT DEFAULT '',
			max_quota REAL NOT NULL DEFAULT 0,
			used_quota REAL NOT NULL DEFAULT 0,
			status INTEGER NOT NULL DEFAULT 1,
			remark TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_keys_provider ON api_keys(provider_id)`,
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'user',
			display_name TEXT DEFAULT '',
			status INTEGER NOT NULL DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS key_user (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			key_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			UNIQUE(key_id, user_id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_keyuser_key ON key_user(key_id)`,
		`CREATE INDEX IF NOT EXISTS idx_keyuser_user ON key_user(user_id)`,
		`CREATE TABLE IF NOT EXISTS chat_sessions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			key_id INTEGER NOT NULL,
			model TEXT NOT NULL DEFAULT '',
			title TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_user ON chat_sessions(user_id)`,
		`CREATE TABLE IF NOT EXISTS chat_messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id INTEGER NOT NULL,
			role TEXT NOT NULL,
			content TEXT NOT NULL DEFAULT '',
			reasoning_content TEXT NOT NULL DEFAULT '',
			prompt_tokens INTEGER NOT NULL DEFAULT 0,
			completion_tokens INTEGER NOT NULL DEFAULT 0,
			cached_tokens INTEGER NOT NULL DEFAULT 0,
			total_tokens INTEGER NOT NULL DEFAULT 0,
			cost REAL NOT NULL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_messages_session ON chat_messages(session_id)`,
	},
	"mysql": {
		`CREATE TABLE IF NOT EXISTS providers (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(128) NOT NULL UNIQUE,
			api_format VARCHAR(16) NOT NULL DEFAULT 'openai',
			base_url VARCHAR(512) NOT NULL,
			platform_url VARCHAR(512) DEFAULT '',
			balance DOUBLE NOT NULL DEFAULT 0,
			max_concurrency INT NOT NULL DEFAULT 10,
			timeout_seconds INT NOT NULL DEFAULT 60,
			status TINYINT NOT NULL DEFAULT 1,
			remark VARCHAR(1024) DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS models (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(128) NOT NULL,
			provider_id BIGINT UNSIGNED NOT NULL,
			input_price_cached DOUBLE NOT NULL DEFAULT 0,
			input_price DOUBLE NOT NULL DEFAULT 0,
			output_price DOUBLE NOT NULL DEFAULT 0,
			context_length INT NOT NULL DEFAULT 0,
			support_thinking TINYINT NOT NULL DEFAULT 0,
			support_multimodal TINYINT NOT NULL DEFAULT 0,
			status TINYINT NOT NULL DEFAULT 1,
			remark VARCHAR(1024) DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			KEY idx_models_provider (provider_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS api_keys (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(128) NOT NULL,
			provider_id BIGINT UNSIGNED NOT NULL,
			api_key TEXT NOT NULL,
			expire_date VARCHAR(20) DEFAULT '',
			max_quota DOUBLE NOT NULL DEFAULT 0,
			used_quota DOUBLE NOT NULL DEFAULT 0,
			status TINYINT NOT NULL DEFAULT 1,
			remark VARCHAR(1024) DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			KEY idx_keys_provider (provider_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS users (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
			username VARCHAR(64) NOT NULL UNIQUE,
			password_hash VARCHAR(255) NOT NULL,
			role VARCHAR(16) NOT NULL DEFAULT 'user',
			display_name VARCHAR(64) DEFAULT '',
			status TINYINT NOT NULL DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS key_user (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
			key_id BIGINT UNSIGNED NOT NULL,
			user_id BIGINT UNSIGNED NOT NULL,
			UNIQUE KEY uk_key_user (key_id, user_id),
			KEY idx_keyuser_key (key_id),
			KEY idx_keyuser_user (user_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS chat_sessions (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
			user_id BIGINT UNSIGNED NOT NULL,
			key_id BIGINT UNSIGNED NOT NULL,
			model VARCHAR(128) NOT NULL DEFAULT '',
			title VARCHAR(255) DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			KEY idx_sessions_user (user_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS chat_messages (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
			session_id BIGINT UNSIGNED NOT NULL,
			role VARCHAR(16) NOT NULL,
			content MEDIUMTEXT,
			reasoning_content VARCHAR(10000) NOT NULL DEFAULT '',
			prompt_tokens INT NOT NULL DEFAULT 0,
			completion_tokens INT NOT NULL DEFAULT 0,
			cached_tokens INT NOT NULL DEFAULT 0,
			total_tokens INT NOT NULL DEFAULT 0,
			cost DOUBLE NOT NULL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			KEY idx_messages_session (session_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	},
}
