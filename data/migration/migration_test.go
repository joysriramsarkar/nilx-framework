package migration

import (
	"testing"
)

func TestMigrationRunner(t *testing.T) {
	runner := NewRunner()
	runner.Register(Migration{
		Version: 1,
		Name:    "init_users",
		UpSQL:   "CREATE TABLE users (id INT);",
	})

	applied := runner.Up()
	if len(applied) != 1 || applied[0] != 1 {
		t.Errorf("expected migration 1 applied, got: %v", applied)
	}

	// Re-run Up should apply 0
	appliedAgain := runner.Up()
	if len(appliedAgain) != 0 {
		t.Errorf("expected 0 new migrations, got: %v", appliedAgain)
	}
}
