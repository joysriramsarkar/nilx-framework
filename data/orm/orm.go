package orm

import (
	"fmt"
	"strings"
	"sync"
)

type QueryBuilder struct {
	table    string
	selected []string
	wheres   []string
	args     []interface{}
	limit    int
}

func Table(name string) *QueryBuilder {
	return &QueryBuilder{
		table:    name,
		selected: []string{"*"},
		wheres:   make([]string, 0),
		args:     make([]interface{}, 0),
	}
}

func (q *QueryBuilder) Select(cols ...string) *QueryBuilder {
	q.selected = cols
	return q
}

func (q *QueryBuilder) Where(col, op string, val interface{}) *QueryBuilder {
	argIdx := len(q.args) + 1
	q.wheres = append(q.wheres, fmt.Sprintf("%s %s $%d", col, op, argIdx))
	q.args = append(q.args, val)
	return q
}

func (q *QueryBuilder) Limit(limit int) *QueryBuilder {
	q.limit = limit
	return q
}

func (q *QueryBuilder) ToSQL() (string, []interface{}) {
	sql := fmt.Sprintf("SELECT %s FROM %s", strings.Join(q.selected, ", "), q.table)
	if len(q.wheres) > 0 {
		sql += " WHERE " + strings.Join(q.wheres, " AND ")
	}
	if q.limit > 0 {
		sql += fmt.Sprintf(" LIMIT %d", q.limit)
	}
	return sql, q.args
}

type Tx struct {
	committed bool
}

func (tx *Tx) Commit() error {
	tx.committed = true
	return nil
}

type DBPool struct {
	mu sync.Mutex
}

func NewDBPool() *DBPool {
	return &DBPool{}
}

func (p *DBPool) Transaction(fn func(tx *Tx) error) error {
	tx := &Tx{}
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}
