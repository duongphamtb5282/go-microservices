package postgresql

import (
	"backend-core/database/interfaces"
	"context"
	"strings"

	"gorm.io/gorm"
)

// PostgreSQLQueryBuilder implements QueryBuilder for PostgreSQL
type PostgreSQLQueryBuilder struct {
	gormDB *gorm.DB
	query  *gorm.DB
	params []interface{}
}

// NewPostgreSQLQueryBuilder creates a new PostgreSQL query builder
func NewPostgreSQLQueryBuilder(gormDB *gorm.DB) *PostgreSQLQueryBuilder {
	return &PostgreSQLQueryBuilder{
		gormDB: gormDB,
		query:  gormDB,
		params: make([]interface{}, 0),
	}
}

// Where adds a WHERE condition
func (q *PostgreSQLQueryBuilder) Where(condition string, args ...interface{}) interfaces.QueryBuilder {
	q.query = q.query.Where(condition, args...)
	q.params = append(q.params, args...)
	return q
}

// And adds AND conditions
func (q *PostgreSQLQueryBuilder) And(conditions map[string]interface{}) interfaces.QueryBuilder {
	for field, value := range conditions {
		q.query = q.query.Where(field+" = ?", value)
		q.params = append(q.params, value)
	}
	return q
}

// Or adds OR conditions
func (q *PostgreSQLQueryBuilder) Or(conditions []map[string]interface{}) interfaces.QueryBuilder {
	if len(conditions) == 0 {
		return q
	}

	var orConditions []string
	var orParams []interface{}

	for _, condition := range conditions {
		var parts []string
		for field, value := range condition {
			parts = append(parts, field+" = ?")
			orParams = append(orParams, value)
		}
		if len(parts) > 0 {
			orConditions = append(orConditions, "("+strings.Join(parts, " AND ")+")")
		}
	}

	if len(orConditions) > 0 {
		q.query = q.query.Where(strings.Join(orConditions, " OR "), orParams...)
		q.params = append(q.params, orParams...)
	}

	return q
}

// Sort adds ORDER BY clause
func (q *PostgreSQLQueryBuilder) Sort(field string, order string) interfaces.QueryBuilder {
	q.query = q.query.Order(field + " " + strings.ToUpper(order))
	return q
}

// Limit adds LIMIT clause
func (q *PostgreSQLQueryBuilder) Limit(limit int) interfaces.QueryBuilder {
	q.query = q.query.Limit(limit)
	return q
}

// Offset adds OFFSET clause
func (q *PostgreSQLQueryBuilder) Offset(offset int) interfaces.QueryBuilder {
	q.query = q.query.Offset(offset)
	return q
}

// Build returns the built query
func (q *PostgreSQLQueryBuilder) Build() interfaces.Query {
	return interfaces.Query{
		SQL:  q.query.ToSQL(func(tx *gorm.DB) *gorm.DB { return tx }),
		Args: q.params,
	}
}

// Execute executes the query
func (q *PostgreSQLQueryBuilder) Execute(ctx context.Context) (interface{}, error) {
	var results []map[string]interface{}
	err := q.query.WithContext(ctx).Find(&results).Error
	return results, err
}

// GetSQL returns the SQL string
func (q *PostgreSQLQueryBuilder) GetSQL() string {
	return q.query.ToSQL(func(tx *gorm.DB) *gorm.DB { return tx })
}

// GetParams returns the query parameters
func (q *PostgreSQLQueryBuilder) GetParams() []interface{} {
	return q.params
}

// Advanced query methods

// WhereIn adds WHERE IN condition
func (q *PostgreSQLQueryBuilder) WhereIn(field string, values []interface{}) interfaces.QueryBuilder {
	if len(values) > 0 {
		q.query = q.query.Where(field+" IN ?", values)
		q.params = append(q.params, values...)
	}
	return q
}

// WhereNotIn adds WHERE NOT IN condition
func (q *PostgreSQLQueryBuilder) WhereNotIn(field string, values []interface{}) interfaces.QueryBuilder {
	if len(values) > 0 {
		q.query = q.query.Where(field+" NOT IN ?", values)
		q.params = append(q.params, values...)
	}
	return q
}

// WhereBetween adds WHERE BETWEEN condition
func (q *PostgreSQLQueryBuilder) WhereBetween(field string, start, end interface{}) interfaces.QueryBuilder {
	q.query = q.query.Where(field+" BETWEEN ? AND ?", start, end)
	q.params = append(q.params, start, end)
	return q
}

// WhereNull adds WHERE IS NULL condition
func (q *PostgreSQLQueryBuilder) WhereNull(field string) interfaces.QueryBuilder {
	q.query = q.query.Where(field + " IS NULL")
	return q
}

// WhereNotNull adds WHERE IS NOT NULL condition
func (q *PostgreSQLQueryBuilder) WhereNotNull(field string) interfaces.QueryBuilder {
	q.query = q.query.Where(field + " IS NOT NULL")
	return q
}

// WhereLike adds WHERE LIKE condition
func (q *PostgreSQLQueryBuilder) WhereLike(field string, pattern string) interfaces.QueryBuilder {
	q.query = q.query.Where(field+" LIKE ?", pattern)
	q.params = append(q.params, pattern)
	return q
}

// WhereRegex adds WHERE REGEXP condition
func (q *PostgreSQLQueryBuilder) WhereRegex(field string, pattern string) interfaces.QueryBuilder {
	q.query = q.query.Where(field+" ~ ?", pattern)
	q.params = append(q.params, pattern)
	return q
}

// OrderBy adds ORDER BY clause
func (q *PostgreSQLQueryBuilder) OrderBy(field string, direction string) interfaces.QueryBuilder {
	q.query = q.query.Order(field + " " + strings.ToUpper(direction))
	return q
}

// OrderByMultiple adds multiple ORDER BY clauses
func (q *PostgreSQLQueryBuilder) OrderByMultiple(orders []interfaces.OrderBy) interfaces.QueryBuilder {
	for _, order := range orders {
		q.query = q.query.Order(order.Field + " " + strings.ToUpper(order.Direction))
	}
	return q
}

// Page adds pagination
func (q *PostgreSQLQueryBuilder) Page(page, pageSize int) interfaces.QueryBuilder {
	if pageSize > 0 {
		offset := (page - 1) * pageSize
		q.query = q.query.Offset(offset).Limit(pageSize)
	}
	return q
}

// Join adds INNER JOIN
func (q *PostgreSQLQueryBuilder) Join(table string, condition string) interfaces.QueryBuilder {
	q.query = q.query.Joins("INNER JOIN " + table + " ON " + condition)
	return q
}

// LeftJoin adds LEFT JOIN
func (q *PostgreSQLQueryBuilder) LeftJoin(table string, condition string) interfaces.QueryBuilder {
	q.query = q.query.Joins("LEFT JOIN " + table + " ON " + condition)
	return q
}

// RightJoin adds RIGHT JOIN
func (q *PostgreSQLQueryBuilder) RightJoin(table string, condition string) interfaces.QueryBuilder {
	q.query = q.query.Joins("RIGHT JOIN " + table + " ON " + condition)
	return q
}

// InnerJoin adds INNER JOIN
func (q *PostgreSQLQueryBuilder) InnerJoin(table string, condition string) interfaces.QueryBuilder {
	q.query = q.query.Joins("INNER JOIN " + table + " ON " + condition)
	return q
}

// Select specifies columns to select
func (q *PostgreSQLQueryBuilder) Select(fields ...string) interfaces.QueryBuilder {
	q.query = q.query.Select(fields)
	return q
}

// Count adds COUNT aggregation
func (q *PostgreSQLQueryBuilder) Count(field string) interfaces.QueryBuilder {
	q.query = q.query.Select("COUNT(" + field + ")")
	return q
}

// Sum adds SUM aggregation
func (q *PostgreSQLQueryBuilder) Sum(field string) interfaces.QueryBuilder {
	q.query = q.query.Select("SUM(" + field + ")")
	return q
}

// Avg adds AVG aggregation
func (q *PostgreSQLQueryBuilder) Avg(field string) interfaces.QueryBuilder {
	q.query = q.query.Select("AVG(" + field + ")")
	return q
}

// Min adds MIN aggregation
func (q *PostgreSQLQueryBuilder) Min(field string) interfaces.QueryBuilder {
	q.query = q.query.Select("MIN(" + field + ")")
	return q
}

// Max adds MAX aggregation
func (q *PostgreSQLQueryBuilder) Max(field string) interfaces.QueryBuilder {
	q.query = q.query.Select("MAX(" + field + ")")
	return q
}

// GroupBy adds GROUP BY clause
func (q *PostgreSQLQueryBuilder) GroupBy(fields ...string) interfaces.QueryBuilder {
	q.query = q.query.Group(strings.Join(fields, ", "))
	return q
}

// Having adds HAVING clause
func (q *PostgreSQLQueryBuilder) Having(condition string, args ...interface{}) interfaces.QueryBuilder {
	q.query = q.query.Having(condition, args...)
	return q
}

// From specifies the table to query from
func (q *PostgreSQLQueryBuilder) From(table string) interfaces.QueryBuilder {
	q.query = q.query.Table(table)
	return q
}
