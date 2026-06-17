package snip

import (
	"github.com/riipandi/elph/internal/datastore"
)

func init() {
	// Register snip migrations
	datastore.Register(datastore.MigrationSet{
		Source: "snip",
		Migrations: []datastore.Migration{
			{
				Version: 6,
				Name:    "create_snip_commands_table",
				Up: `CREATE TABLE IF NOT EXISTS snip_commands (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					timestamp DATETIME DEFAULT (datetime('now')),
					original_cmd TEXT NOT NULL,
					snip_cmd TEXT NOT NULL,
					input_tokens INTEGER NOT NULL,
					output_tokens INTEGER NOT NULL,
					saved_tokens INTEGER NOT NULL,
					savings_pct REAL NOT NULL,
					exec_time_ms INTEGER NOT NULL
				);

				CREATE INDEX IF NOT EXISTS idx_snip_commands_timestamp
				ON snip_commands(timestamp);

				CREATE INDEX IF NOT EXISTS idx_snip_commands_original_cmd
				ON snip_commands(original_cmd);`,
			},
		},
	})
}
