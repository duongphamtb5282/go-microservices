package interfaces

import "context"

// TransactionManager defines the transaction management interface
type TransactionManager interface {
	// Transaction operations
	BeginTransaction(ctx context.Context) (Transaction, error)
	WithTransaction(ctx context.Context, fn func(Transaction) error) error
}

// Transaction defines the transaction interface
type Transaction interface {
	// Transaction control
	Commit() error
	Rollback() error
	IsActive() bool

	// Repository operations within transaction
	GetRepository() interface{}
}
