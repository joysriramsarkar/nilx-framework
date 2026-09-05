package orm

import (
	"strings"
	"testing"
)

func TestQueryBuilder(t *testing.T) {
	sql, args := Table("users").
		Select("id", "email").
		Where("status", "=", "active").
		Limit(10).
		ToSQL()

	if !strings.Contains(sql, "SELECT id, email FROM users WHERE status = $1 LIMIT 10") {
		t.Errorf("unexpected SQL: %s", sql)
	}
	if len(args) != 1 || args[0] != "active" {
		t.Errorf("unexpected args: %v", args)
	}
}

func TestDBPool(t *testing.T) {
	pool := NewDBPool()
	err := pool.Transaction(func(tx *Tx) error {
		return nil
	})
	if err != nil {
		t.Errorf("transaction failed: %v", err)
	}
}
