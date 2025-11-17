package mongodb

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"backend-core/database"
	"backend-core/logging"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

// MongoDBRepository implements the Repository interface for MongoDB
type MongoDBRepository[T any] struct {
	collection *mongo.Collection
	logger     *logging.Logger
	client     *mongo.Client
}

// ValidateEntity validates an entity before operations
func (r *MongoDBRepository[T]) ValidateEntity(entity *T) error {
	// Basic validation - can be extended as needed
	if entity == nil {
		return fmt.Errorf("entity cannot be nil")
	}
	return nil
}

// LogOperation logs repository operations
func (r *MongoDBRepository[T]) LogOperation(operation string, err error, fields ...zap.Field) {
	if err != nil {
		r.logger.Error("Repository operation failed",
			logging.String("operation", operation),
			logging.Error(err))
	} else {
		r.logger.Info("Repository operation completed",
			logging.String("operation", operation))
	}
}

// LogQuery logs database queries
func (r *MongoDBRepository[T]) LogQuery(query string, entity interface{}, duration time.Duration, err error) {
	if err != nil {
		r.logger.Error("Database query failed",
			logging.String("query", query),
			logging.Duration("duration", duration),
			logging.Error(err))
	} else {
		r.logger.Info("Database query completed",
			logging.String("query", query),
			logging.Duration("duration", duration))
	}
}

// ValidateID validates an ID before operations
func (r *MongoDBRepository[T]) ValidateID(id interface{}) error {
	// Basic validation - can be extended as needed
	if id == nil {
		return fmt.Errorf("id cannot be nil")
	}
	return nil
}

// NewMongoDBRepository creates a new MongoDB repository
func NewMongoDBRepository[T any](db *MongoDBDatabase, collectionName string) *MongoDBRepository[T] {
	return &MongoDBRepository[T]{
		collection: db.database.Collection(collectionName),
		client:     db.client,
		logger:     db.logger,
	}
}

// Create creates a new entity
func (r *MongoDBRepository[T]) Create(ctx context.Context, entity *T) error {
	start := time.Now()

	if err := r.ValidateEntity(entity); err != nil {
		r.LogOperation("create", err, zap.String("error", "validation_failed"))
		return err
	}

	_, err := r.collection.InsertOne(ctx, entity)
	duration := time.Since(start)

	r.LogQuery("INSERT_ONE", entity, duration, err)
	r.LogOperation("create", err, zap.Duration("duration", duration))

	return err
}

// CreateBatch creates multiple entities
func (r *MongoDBRepository[T]) CreateBatch(ctx context.Context, entities []*T) error {
	start := time.Now()

	if len(entities) == 0 {
		return fmt.Errorf("no entities to create")
	}

	// Convert to []interface{} for MongoDB
	docs := make([]interface{}, len(entities))
	for i, entity := range entities {
		docs[i] = entity
	}

	_, err := r.collection.InsertMany(ctx, docs)
	duration := time.Since(start)

	r.LogQuery("INSERT_MANY", entities, duration, err)
	r.LogOperation("create_batch", err,
		zap.Int("count", len(entities)),
		zap.Duration("duration", duration),
	)

	return err
}

// GetByID retrieves an entity by ID
func (r *MongoDBRepository[T]) GetByID(ctx context.Context, id interface{}) (*T, error) {
	start := time.Now()

	if err := r.ValidateID(id); err != nil {
		r.LogOperation("get_by_id", err, zap.String("error", "validation_failed"))
		return nil, err
	}

	// Convert ID to ObjectID if it's a string
	objectID, err := r.convertToObjectID(id)
	if err != nil {
		return nil, fmt.Errorf("invalid ID format: %w", err)
	}

	var entity T
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&entity)
	duration := time.Since(start)

	r.LogQuery("FIND_ONE", map[string]interface{}{"id": id}, duration, err)
	r.LogOperation("get_by_id", err,
		zap.Any("id", id),
		zap.Duration("duration", duration),
	)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("entity not found")
		}
		return nil, err
	}

	return &entity, nil
}

// GetByField retrieves an entity by field value
func (r *MongoDBRepository[T]) GetByField(ctx context.Context, field string, value interface{}) (*T, error) {
	start := time.Now()

	var entity T
	err := r.collection.FindOne(ctx, bson.M{field: value}).Decode(&entity)
	duration := time.Since(start)

	r.LogQuery("FIND_ONE", map[string]interface{}{"field": field, "value": value}, duration, err)
	r.LogOperation("get_by_field", err,
		zap.String("field", field),
		zap.Any("value", value),
		zap.Duration("duration", duration),
	)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("entity not found")
		}
		return nil, err
	}

	return &entity, nil
}

// GetAll retrieves entities with filter and pagination
func (r *MongoDBRepository[T]) GetAll(ctx context.Context, filter database.Filter, pagination database.Pagination) ([]*T, error) {
	start := time.Now()

	// Convert filter to MongoDB filter
	mongoFilter := r.convertFilter(filter)

	// Build options
	opts := options.Find()

	// Apply pagination
	if pagination.PageSize > 0 {
		offset := (pagination.Page - 1) * pagination.PageSize
		opts.SetSkip(int64(offset))
		opts.SetLimit(int64(pagination.PageSize))
	}

	cursor, err := r.collection.Find(ctx, mongoFilter, opts)
	if err != nil {
		duration := time.Since(start)
		r.LogQuery("FIND", map[string]interface{}{"filter": filter, "pagination": pagination}, duration, err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var entities []*T
	err = cursor.All(ctx, &entities)
	duration := time.Since(start)

	r.LogQuery("FIND", map[string]interface{}{"filter": filter, "pagination": pagination}, duration, err)
	r.LogOperation("get_all", err,
		zap.Any("filter", filter),
		zap.Any("pagination", pagination),
		zap.Int("count", len(entities)),
		zap.Duration("duration", duration),
	)

	return entities, err
}

// Update updates an entity
func (r *MongoDBRepository[T]) Update(ctx context.Context, entity *T) error {
	start := time.Now()

	if err := r.ValidateEntity(entity); err != nil {
		r.LogOperation("update", err, zap.String("error", "validation_failed"))
		return err
	}

	// Extract ID from entity
	id, err := r.extractID(entity)
	if err != nil {
		return fmt.Errorf("failed to extract ID from entity: %w", err)
	}

	objectID, err := r.convertToObjectID(id)
	if err != nil {
		return fmt.Errorf("invalid ID format: %w", err)
	}

	_, err = r.collection.ReplaceOne(ctx, bson.M{"_id": objectID}, entity)
	duration := time.Since(start)

	r.LogQuery("REPLACE_ONE", entity, duration, err)
	r.LogOperation("update", err, zap.Duration("duration", duration))

	return err
}

// UpdateField updates a specific field of an entity
func (r *MongoDBRepository[T]) UpdateField(ctx context.Context, id interface{}, field string, value interface{}) error {
	start := time.Now()

	if err := r.ValidateID(id); err != nil {
		r.LogOperation("update_field", err, zap.String("error", "validation_failed"))
		return err
	}

	objectID, err := r.convertToObjectID(id)
	if err != nil {
		return fmt.Errorf("invalid ID format: %w", err)
	}

	_, err = r.collection.UpdateOne(ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": bson.M{field: value}},
	)
	duration := time.Since(start)

	r.LogQuery("UPDATE_ONE", map[string]interface{}{"id": id, "field": field, "value": value}, duration, err)
	r.LogOperation("update_field", err,
		zap.Any("id", id),
		zap.String("field", field),
		zap.Any("value", value),
		zap.Duration("duration", duration),
	)

	return err
}

// Upsert creates or updates an entity
func (r *MongoDBRepository[T]) Upsert(ctx context.Context, filter database.Filter, entity *T) error {
	start := time.Now()

	if err := r.ValidateEntity(entity); err != nil {
		r.LogOperation("upsert", err, zap.String("error", "validation_failed"))
		return err
	}

	mongoFilter := r.convertFilter(filter)
	opts := options.Replace().SetUpsert(true)

	_, err := r.collection.ReplaceOne(ctx, mongoFilter, entity, opts)
	duration := time.Since(start)

	r.LogQuery("REPLACE_ONE_UPSERT", map[string]interface{}{"filter": filter, "entity": entity}, duration, err)
	r.LogOperation("upsert", err,
		zap.Any("filter", filter),
		zap.Duration("duration", duration),
	)

	return err
}

// Delete deletes an entity by ID
func (r *MongoDBRepository[T]) Delete(ctx context.Context, id interface{}) error {
	start := time.Now()

	if err := r.ValidateID(id); err != nil {
		r.LogOperation("delete", err, zap.String("error", "validation_failed"))
		return err
	}

	objectID, err := r.convertToObjectID(id)
	if err != nil {
		return fmt.Errorf("invalid ID format: %w", err)
	}

	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	duration := time.Since(start)

	r.LogQuery("DELETE_ONE", map[string]interface{}{"id": id}, duration, err)
	r.LogOperation("delete", err,
		zap.Any("id", id),
		zap.Duration("duration", duration),
	)

	return err
}

// DeleteBatch deletes multiple entities by IDs
func (r *MongoDBRepository[T]) DeleteBatch(ctx context.Context, ids []interface{}) error {
	start := time.Now()

	if len(ids) == 0 {
		return fmt.Errorf("no IDs provided for batch delete")
	}

	// Convert IDs to ObjectIDs
	objectIDs := make([]primitive.ObjectID, len(ids))
	for i, id := range ids {
		objectID, err := r.convertToObjectID(id)
		if err != nil {
			return fmt.Errorf("invalid ID format at index %d: %w", i, err)
		}
		objectIDs[i] = objectID
	}

	_, err := r.collection.DeleteMany(ctx, bson.M{"_id": bson.M{"$in": objectIDs}})
	duration := time.Since(start)

	r.LogQuery("DELETE_MANY", map[string]interface{}{"ids": ids}, duration, err)
	r.LogOperation("delete_batch", err,
		zap.Int("count", len(ids)),
		zap.Duration("duration", duration),
	)

	return err
}

// Count counts entities matching the filter
func (r *MongoDBRepository[T]) Count(ctx context.Context, filter database.Filter) (int64, error) {
	start := time.Now()

	mongoFilter := r.convertFilter(filter)

	count, err := r.collection.CountDocuments(ctx, mongoFilter)
	duration := time.Since(start)

	r.LogQuery("COUNT", map[string]interface{}{"filter": filter}, duration, err)
	r.LogOperation("count", err,
		zap.Any("filter", filter),
		zap.Int64("count", count),
		zap.Duration("duration", duration),
	)

	return count, err
}

// Exists checks if entities exist matching the filter
func (r *MongoDBRepository[T]) Exists(ctx context.Context, filter database.Filter) (bool, error) {
	count, err := r.Count(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Find executes a complex query
func (r *MongoDBRepository[T]) Find(ctx context.Context, query database.Query) ([]*T, error) {
	start := time.Now()

	mongoFilter := r.convertFilter(query.Filter)

	// Build options
	opts := options.Find()

	// Apply ordering
	if query.OrderBy != "" {
		order := 1
		if query.Order == "desc" {
			order = -1
		}
		opts.SetSort(bson.M{query.OrderBy: order})
	}

	// Apply pagination
	if query.Pagination.PageSize > 0 {
		offset := (query.Pagination.Page - 1) * query.Pagination.PageSize
		opts.SetSkip(int64(offset))
		opts.SetLimit(int64(query.Pagination.PageSize))
	}

	cursor, err := r.collection.Find(ctx, mongoFilter, opts)
	if err != nil {
		duration := time.Since(start)
		r.LogQuery("FIND", query, duration, err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var entities []*T
	err = cursor.All(ctx, &entities)
	duration := time.Since(start)

	r.LogQuery("FIND", query, duration, err)
	r.LogOperation("find", err,
		zap.Any("query", query),
		zap.Int("count", len(entities)),
		zap.Duration("duration", duration),
	)

	return entities, err
}

// WithTransaction executes a function within a transaction
func (r *MongoDBRepository[T]) WithTransaction(ctx context.Context, fn func(database.Repository[T]) error) error {
	session, err := r.client.StartSession()
	if err != nil {
		return fmt.Errorf("failed to start session: %w", err)
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		// Create a new repository with the session context
		txRepo := &MongoDBRepository[T]{
			collection: r.collection,
			client:     r.client,
			logger:     r.logger,
		}
		return nil, fn(txRepo)
	})

	return err
}

// GetQueryBuilder returns the MongoDB query builder
func (r *MongoDBRepository[T]) GetQueryBuilder() database.QueryBuilder {
	return NewMongoDBQueryBuilder(r.collection.Database())
}

// GetTransactionManager returns the transaction manager
func (r *MongoDBRepository[T]) GetTransactionManager() database.TransactionManager {
	return NewMongoDBTransactionManager(r.client)
}

// Helper methods

func (r *MongoDBRepository[T]) convertFilter(filter database.Filter) bson.M {
	mongoFilter := bson.M{}
	for field, value := range filter {
		mongoFilter[field] = value
	}
	return mongoFilter
}

func (r *MongoDBRepository[T]) convertToObjectID(id interface{}) (primitive.ObjectID, error) {
	switch v := id.(type) {
	case string:
		return primitive.ObjectIDFromHex(v)
	case primitive.ObjectID:
		return v, nil
	default:
		return primitive.NilObjectID, fmt.Errorf("unsupported ID type: %T", id)
	}
}

func (r *MongoDBRepository[T]) extractID(entity *T) (interface{}, error) {
	// This is a simplified implementation
	// In a real implementation, you would use reflection to find the ID field
	// or use struct tags to identify the ID field
	val := reflect.ValueOf(entity).Elem()

	// Look for common ID field names
	idFields := []string{"ID", "Id", "id", "_id"}
	for _, fieldName := range idFields {
		field := val.FieldByName(fieldName)
		if field.IsValid() && !field.IsZero() {
			return field.Interface(), nil
		}
	}

	return nil, fmt.Errorf("no ID field found in entity")
}

func getCollectionNameFromType[T any]() string {
	var t T
	rt := reflect.TypeOf(t)
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}
	return rt.Name()
}
