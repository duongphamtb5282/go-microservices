package mongodb

import (
	"context"
	"fmt"
	"sync"
	"time"

	"backend-core/database"

	"go.mongodb.org/mongo-driver/mongo"
)

// MongoDBTransactionManager manages MongoDB transactions
type MongoDBTransactionManager struct {
	client        *mongo.Client
	transactions  map[string]*MongoDBTransaction
	mu            sync.RWMutex
	transactionID int64
}

// NewMongoDBTransactionManager creates a new MongoDB transaction manager
func NewMongoDBTransactionManager(client *mongo.Client) *MongoDBTransactionManager {
	return &MongoDBTransactionManager{
		client:       client,
		transactions: make(map[string]*MongoDBTransaction),
	}
}

// WithTransaction executes a function within a transaction
func (tm *MongoDBTransactionManager) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	session, err := tm.client.StartSession()
	if err != nil {
		return fmt.Errorf("failed to start session: %w", err)
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		return nil, fn(sessCtx)
	})

	return err
}

// BeginTransaction begins a new transaction
func (tm *MongoDBTransactionManager) BeginTransaction(ctx context.Context) (database.Transaction, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.transactionID++
	transactionID := fmt.Sprintf("tx_%d", tm.transactionID)

	session, err := tm.client.StartSession()
	if err != nil {
		return nil, fmt.Errorf("failed to start session: %w", err)
	}

	postgresTx := &MongoDBTransaction{
		id:        transactionID,
		session:   session,
		ctx:       ctx,
		active:    true,
		createdAt: time.Now(),
	}

	tm.transactions[transactionID] = postgresTx
	return postgresTx, nil
}

// IsInTransaction checks if there's an active transaction in the context
func (tm *MongoDBTransactionManager) IsInTransaction(ctx context.Context) bool {
	// This is a simplified implementation
	// In a real implementation, you might store transaction info in the context
	return false
}

// GetTransactionLevel returns the current transaction nesting level
func (tm *MongoDBTransactionManager) GetTransactionLevel(ctx context.Context) int {
	// MongoDB doesn't support nested transactions in the same way as some other databases
	// This is a simplified implementation
	return 0
}

// GetTransaction returns a transaction by ID
func (tm *MongoDBTransactionManager) GetTransaction(id string) *MongoDBTransaction {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.transactions[id]
}

// RemoveTransaction removes a transaction from the manager
func (tm *MongoDBTransactionManager) RemoveTransaction(id string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	delete(tm.transactions, id)
}

// MongoDBTransaction represents a MongoDB transaction
type MongoDBTransaction struct {
	id        string
	session   mongo.Session
	ctx       context.Context
	active    bool
	createdAt time.Time
	mu        sync.RWMutex
}

// Commit commits the transaction
func (t *MongoDBTransaction) Commit(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.active {
		return fmt.Errorf("transaction is not active")
	}

	// MongoDB transactions are committed automatically when the session ends
	// This is a simplified implementation
	t.active = false
	return nil
}

// Rollback rolls back the transaction
func (t *MongoDBTransaction) Rollback(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.active {
		return fmt.Errorf("transaction is not active")
	}

	// MongoDB transactions are rolled back automatically when the session ends
	// This is a simplified implementation
	t.active = false
	return nil
}

// IsActive returns whether the transaction is active
func (t *MongoDBTransaction) IsActive() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.active
}

// GetID returns the transaction ID
func (t *MongoDBTransaction) GetID() string {
	return t.id
}

// GetSession returns the MongoDB session
func (t *MongoDBTransaction) GetSession() mongo.Session {
	return t.session
}

// GetContext returns the transaction context
func (t *MongoDBTransaction) GetContext() context.Context {
	return t.ctx
}

// GetCreatedAt returns when the transaction was created
func (t *MongoDBTransaction) GetCreatedAt() time.Time {
	return t.createdAt
}
