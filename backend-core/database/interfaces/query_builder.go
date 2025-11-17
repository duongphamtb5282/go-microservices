package interfaces

import "context"

// OrderBy represents an ordering clause
type OrderBy struct {
	Field     string
	Direction string
}

// QueryBuilder defines the query builder interface
type QueryBuilder interface {
	// Basic query building
	Select(fields ...string) QueryBuilder
	From(table string) QueryBuilder
	Where(condition string, args ...interface{}) QueryBuilder
	OrderBy(field string, direction string) QueryBuilder
	GroupBy(fields ...string) QueryBuilder
	Having(condition string, args ...interface{}) QueryBuilder

	// Joins
	Join(table string, condition string) QueryBuilder
	LeftJoin(table string, condition string) QueryBuilder
	RightJoin(table string, condition string) QueryBuilder
	InnerJoin(table string, condition string) QueryBuilder

	// Pagination
	Page(page, pageSize int) QueryBuilder
	Limit(limit int) QueryBuilder
	Offset(offset int) QueryBuilder

	// Execution
	Build() Query
	Execute(ctx context.Context) (interface{}, error)
}
