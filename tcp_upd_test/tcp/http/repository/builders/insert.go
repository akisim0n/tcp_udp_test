package builders

import (
	"fmt"
	"strings"
)

type InsertBuilder struct {
	columns []string
	values  []interface{}
	table   string
	retCols []string
}

func NewInsertBuilder(table string) *InsertBuilder {
	return &InsertBuilder{table: table}
}

func (b *InsertBuilder) Set(col string, val interface{}) *InsertBuilder {
	b.columns = append(b.columns, col)
	b.values = append(b.values, val)
	return b
}

func (b *InsertBuilder) Build() (string, []interface{}) {
	valuePlaceholders := make([]string, len(b.columns))

	for i := range b.columns {
		valuePlaceholders[i] = fmt.Sprintf("$%d", i+1)
	}

	queryText := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		b.table,
		strings.Join(b.columns, ","),
		strings.Join(valuePlaceholders, ","))

	if b.retCols != nil && len(b.retCols) > 0 {
		queryText += " RETURNING " + strings.Join(b.retCols, ",")
	}

	return queryText, b.values
}

func (b *InsertBuilder) SetReturning(cols ...string) *InsertBuilder {
	b.retCols = cols
	return b
}
