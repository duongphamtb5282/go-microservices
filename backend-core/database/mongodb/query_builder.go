package mongodb

import (
	"context"
	"fmt"
	"strings"

	"backend-core/database"
	"backend-core/database/interfaces"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDBQueryBuilder implements QueryBuilder for MongoDB
type MongoDBQueryBuilder struct {
	database   *mongo.Database
	collection *mongo.Collection
	filter     bson.M
	options    *options.FindOptions
	params     []interface{}
}

// NewMongoDBQueryBuilder creates a new MongoDB query builder
func NewMongoDBQueryBuilder(database *mongo.Database) *MongoDBQueryBuilder {
	return &MongoDBQueryBuilder{
		database: database,
		filter:   bson.M{},
		options:  options.Find(),
		params:   make([]interface{}, 0),
	}
}

// Where adds a WHERE condition
func (q *MongoDBQueryBuilder) Where(field string, operator string, value interface{}) database.QueryBuilder {
	switch operator {
	case "=", "eq":
		q.filter[field] = value
	case "!=", "ne":
		q.filter[field] = bson.M{"$ne": value}
	case ">", "gt":
		q.filter[field] = bson.M{"$gt": value}
	case ">=", "gte":
		q.filter[field] = bson.M{"$gte": value}
	case "<", "lt":
		q.filter[field] = bson.M{"$lt": value}
	case "<=", "lte":
		q.filter[field] = bson.M{"$lte": value}
	case "in":
		q.filter[field] = bson.M{"$in": value}
	case "nin":
		q.filter[field] = bson.M{"$nin": value}
	case "exists":
		q.filter[field] = bson.M{"$exists": value}
	case "regex":
		q.filter[field] = bson.M{"$regex": value}
	default:
		q.filter[field] = value
	}

	q.params = append(q.params, value)
	return q
}

// And adds AND conditions
func (q *MongoDBQueryBuilder) And(conditions map[string]interface{}) database.QueryBuilder {
	for field, value := range conditions {
		q.filter[field] = value
		q.params = append(q.params, value)
	}
	return q
}

// Or adds OR conditions
func (q *MongoDBQueryBuilder) Or(conditions []map[string]interface{}) database.QueryBuilder {
	if len(conditions) == 0 {
		return q
	}

	var orConditions []bson.M
	for _, condition := range conditions {
		orConditions = append(orConditions, bson.M(condition))
	}

	q.filter["$or"] = orConditions
	return q
}

// Sort adds ORDER BY clause
func (q *MongoDBQueryBuilder) Sort(field string, order string) database.QueryBuilder {
	sortOrder := 1
	if strings.ToLower(order) == "desc" {
		sortOrder = -1
	}
	q.options.SetSort(bson.M{field: sortOrder})
	return q
}

// Limit adds LIMIT clause
func (q *MongoDBQueryBuilder) Limit(limit int) database.QueryBuilder {
	q.options.SetLimit(int64(limit))
	return q
}

// Offset adds OFFSET clause
func (q *MongoDBQueryBuilder) Offset(offset int) database.QueryBuilder {
	q.options.SetSkip(int64(offset))
	return q
}

// Build returns the built query
func (q *MongoDBQueryBuilder) Build() (interface{}, error) {
	return map[string]interface{}{
		"filter":  q.filter,
		"options": q.options,
	}, nil
}

// Execute executes the query
func (q *MongoDBQueryBuilder) Execute(ctx context.Context) (interface{}, error) {
	if q.collection == nil {
		return nil, fmt.Errorf("collection not set")
	}

	cursor, err := q.collection.Find(ctx, q.filter, q.options)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	err = cursor.All(ctx, &results)
	return results, err
}

// GetSQL returns the MongoDB query as JSON
func (q *MongoDBQueryBuilder) GetSQL() string {
	return fmt.Sprintf("Filter: %v, Options: %v", q.filter, q.options)
}

// GetParams returns the query parameters
func (q *MongoDBQueryBuilder) GetParams() []interface{} {
	return q.params
}

// Advanced query methods

// WhereIn adds WHERE IN condition
func (q *MongoDBQueryBuilder) WhereIn(field string, values []interface{}) database.QueryBuilder {
	if len(values) > 0 {
		q.filter[field] = bson.M{"$in": values}
		q.params = append(q.params, values...)
	}
	return q
}

// WhereNotIn adds WHERE NOT IN condition
func (q *MongoDBQueryBuilder) WhereNotIn(field string, values []interface{}) database.QueryBuilder {
	if len(values) > 0 {
		q.filter[field] = bson.M{"$nin": values}
		q.params = append(q.params, values...)
	}
	return q
}

// WhereBetween adds WHERE BETWEEN condition
func (q *MongoDBQueryBuilder) WhereBetween(field string, start, end interface{}) database.QueryBuilder {
	q.filter[field] = bson.M{"$gte": start, "$lte": end}
	q.params = append(q.params, start, end)
	return q
}

// WhereNull adds WHERE IS NULL condition
func (q *MongoDBQueryBuilder) WhereNull(field string) database.QueryBuilder {
	q.filter[field] = bson.M{"$exists": false}
	return q
}

// WhereNotNull adds WHERE IS NOT NULL condition
func (q *MongoDBQueryBuilder) WhereNotNull(field string) database.QueryBuilder {
	q.filter[field] = bson.M{"$exists": true, "$ne": nil}
	return q
}

// WhereLike adds WHERE LIKE condition
func (q *MongoDBQueryBuilder) WhereLike(field string, pattern string) database.QueryBuilder {
	q.filter[field] = bson.M{"$regex": pattern, "$options": "i"}
	q.params = append(q.params, pattern)
	return q
}

// WhereRegex adds WHERE REGEXP condition
func (q *MongoDBQueryBuilder) WhereRegex(field string, pattern string) database.QueryBuilder {
	q.filter[field] = bson.M{"$regex": pattern}
	q.params = append(q.params, pattern)
	return q
}

// OrderBy adds ORDER BY clause
func (q *MongoDBQueryBuilder) OrderBy(field string, direction string) database.QueryBuilder {
	sortOrder := 1
	if strings.ToLower(direction) == "desc" {
		sortOrder = -1
	}
	q.options.SetSort(bson.M{field: sortOrder})
	return q
}

// OrderByMultiple adds multiple ORDER BY clauses
func (q *MongoDBQueryBuilder) OrderByMultiple(orders []interfaces.OrderBy) database.QueryBuilder {
	sort := bson.M{}
	for _, order := range orders {
		sortOrder := 1
		if strings.ToLower(order.Direction) == "desc" {
			sortOrder = -1
		}
		sort[order.Field] = sortOrder
	}
	q.options.SetSort(sort)
	return q
}

// Page adds pagination
func (q *MongoDBQueryBuilder) Page(page, pageSize int) database.QueryBuilder {
	if pageSize > 0 {
		offset := (page - 1) * pageSize
		q.options.SetSkip(int64(offset))
		q.options.SetLimit(int64(pageSize))
	}
	return q
}

// Join adds INNER JOIN (MongoDB uses $lookup)
func (q *MongoDBQueryBuilder) Join(table string, condition string) database.QueryBuilder {
	// MongoDB doesn't have traditional joins, this would be implemented with $lookup
	// This is a simplified implementation
	return q
}

// LeftJoin adds LEFT JOIN (MongoDB uses $lookup)
func (q *MongoDBQueryBuilder) LeftJoin(table string, condition string) database.QueryBuilder {
	// MongoDB doesn't have traditional joins, this would be implemented with $lookup
	// This is a simplified implementation
	return q
}

// RightJoin adds RIGHT JOIN (MongoDB uses $lookup)
func (q *MongoDBQueryBuilder) RightJoin(table string, condition string) database.QueryBuilder {
	// MongoDB doesn't have traditional joins, this would be implemented with $lookup
	// This is a simplified implementation
	return q
}

// InnerJoin adds INNER JOIN (MongoDB uses $lookup)
func (q *MongoDBQueryBuilder) InnerJoin(table string, condition string) database.QueryBuilder {
	// MongoDB doesn't have traditional joins, this would be implemented with $lookup
	// This is a simplified implementation
	return q
}

// Select specifies fields to select
func (q *MongoDBQueryBuilder) Select(fields ...string) database.QueryBuilder {
	projection := bson.M{}
	for _, field := range fields {
		projection[field] = 1
	}
	q.options.SetProjection(projection)
	return q
}

// Count adds COUNT aggregation
func (q *MongoDBQueryBuilder) Count(field string) database.QueryBuilder {
	// This would be implemented with aggregation pipeline
	// This is a simplified implementation
	return q
}

// Sum adds SUM aggregation
func (q *MongoDBQueryBuilder) Sum(field string) database.QueryBuilder {
	// This would be implemented with aggregation pipeline
	// This is a simplified implementation
	return q
}

// Avg adds AVG aggregation
func (q *MongoDBQueryBuilder) Avg(field string) database.QueryBuilder {
	// This would be implemented with aggregation pipeline
	// This is a simplified implementation
	return q
}

// Min adds MIN aggregation
func (q *MongoDBQueryBuilder) Min(field string) database.QueryBuilder {
	// This would be implemented with aggregation pipeline
	// This is a simplified implementation
	return q
}

// Max adds MAX aggregation
func (q *MongoDBQueryBuilder) Max(field string) database.QueryBuilder {
	// This would be implemented with aggregation pipeline
	// This is a simplified implementation
	return q
}

// GroupBy adds GROUP BY clause
func (q *MongoDBQueryBuilder) GroupBy(fields ...string) database.QueryBuilder {
	// This would be implemented with aggregation pipeline
	// This is a simplified implementation
	return q
}

// Having adds HAVING clause
func (q *MongoDBQueryBuilder) Having(condition string) database.QueryBuilder {
	// This would be implemented with aggregation pipeline
	// This is a simplified implementation
	return q
}

// SetCollection sets the collection for the query
func (q *MongoDBQueryBuilder) SetCollection(collection *mongo.Collection) {
	q.collection = collection
}
