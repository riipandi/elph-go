package datastore

func init() {
	// Register core migrations
	Register(MigrationSet{
		Source: "core",
		Migrations: []Migration{
			{
				Version: 1,
				Name:    "create_sessions_table",
				Up: `CREATE TABLE IF NOT EXISTS sessions (
					id TEXT PRIMARY KEY,
					created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
					updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
					work_dir TEXT,
					provider_id TEXT,
					model_id TEXT,
					agent_mode TEXT DEFAULT 'build',
					system_prompt TEXT,
					metadata TEXT
				);

				CREATE INDEX IF NOT EXISTS idx_sessions_created_at
				ON sessions(created_at);

				CREATE INDEX IF NOT EXISTS idx_sessions_work_dir
				ON sessions(work_dir);`,
			},
			{
				Version: 2,
				Name:    "create_messages_table",
				Up: `CREATE TABLE IF NOT EXISTS messages (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					session_id TEXT NOT NULL,
					role TEXT NOT NULL,
					content TEXT,
					tool_call_id TEXT,
					tool_calls TEXT,
					created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
					FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
				);

				CREATE INDEX IF NOT EXISTS idx_messages_session_id
				ON messages(session_id);

				CREATE INDEX IF NOT EXISTS idx_messages_created_at
				ON messages(created_at);`,
			},
			{
				Version: 3,
				Name:    "create_todos_table",
				Up: `CREATE TABLE IF NOT EXISTS todos (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					session_id TEXT NOT NULL,
					content TEXT NOT NULL,
					completed BOOLEAN NOT NULL DEFAULT 0,
					position INTEGER NOT NULL DEFAULT 0,
					created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
					updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
					FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
				);

				CREATE INDEX IF NOT EXISTS idx_todos_session_id
				ON todos(session_id);

				CREATE INDEX IF NOT EXISTS idx_todos_position
				ON todos(session_id, position);`,
			},
			{
				Version: 4,
				Name:    "create_goals_table",
				Up: `CREATE TABLE IF NOT EXISTS goals (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					session_id TEXT NOT NULL,
					objective TEXT NOT NULL,
					completion_criterion TEXT,
					status TEXT NOT NULL DEFAULT 'active',
					turns_used INTEGER NOT NULL DEFAULT 0,
					tokens_used INTEGER NOT NULL DEFAULT 0,
					wall_clock_ms INTEGER NOT NULL DEFAULT 0,
					wall_clock_budget_ms INTEGER NOT NULL DEFAULT 0,
					turn_budget INTEGER NOT NULL DEFAULT 0,
					token_budget INTEGER NOT NULL DEFAULT 0,
					created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
					completed_at DATETIME,
					FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
				);

				CREATE INDEX IF NOT EXISTS idx_goals_session_id
				ON goals(session_id);

				CREATE INDEX IF NOT EXISTS idx_goals_status
				ON goals(status);`,
			},
			{
				Version: 5,
				Name:    "create_skill_cache_table",
				Up: `CREATE TABLE IF NOT EXISTS skill_cache (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					skill_name TEXT NOT NULL,
					skill_hash TEXT NOT NULL,
					content TEXT NOT NULL,
					created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
					expires_at DATETIME,
					UNIQUE(skill_name, skill_hash)
				);

				CREATE INDEX IF NOT EXISTS idx_skill_cache_name
				ON skill_cache(skill_name);

				CREATE INDEX IF NOT EXISTS idx_skill_cache_expires
				ON skill_cache(expires_at);`,
			},
		},
	})
}
