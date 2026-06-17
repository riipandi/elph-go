package snip

import (
	"testing"

	"github.com/riipandi/elph/internal/datastore"
	"github.com/stretchr/testify/require"
)

func TestSnipMigrationsRegistered(t *testing.T) {
	sets := datastore.GetAllMigrations()

	// Find snip migrations
	var found bool
	for _, set := range sets {
		if set.Source == "snip" {
			found = true
			require.Len(t, set.Migrations, 1)
			require.Equal(t, int64(6), set.Migrations[0].Version)
			require.Equal(t, "create_snip_commands_table", set.Migrations[0].Name)
			break
		}
	}
	require.True(t, found, "snip migrations not registered")
}
