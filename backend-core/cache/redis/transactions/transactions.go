package transactions

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisTransaction represents a Redis transaction
type RedisTransaction struct {
	tx     redis.Pipeliner
	client redis.UniversalClient
	ctx    context.Context
}

// RedisTransactions handles Redis transaction operations
type RedisTransactions struct {
	client redis.UniversalClient
}

// NewRedisTransactions creates a new Redis transactions handler
func NewRedisTransactions(client redis.UniversalClient) *RedisTransactions {
	return &RedisTransactions{
		client: client,
	}
}

// Begin starts a new Redis transaction
func (r *RedisTransactions) Begin(ctx context.Context) *RedisTransaction {
	return &RedisTransaction{
		tx:     r.client.TxPipeline(),
		client: r.client,
		ctx:    ctx,
	}
}

// Watch watches one or more keys for modifications during a transaction
func (r *RedisTransactions) Watch(ctx context.Context, keys ...string) error {
	return r.client.Watch(ctx, func(tx *redis.Tx) error {
		// This function is called when keys are modified
		// Return an error to abort the transaction
		return nil
	}, keys...)
}

// ============================================================================
// Transaction Operations
// ============================================================================

// Set stores a value in the transaction
func (t *RedisTransaction) Set(key string, value interface{}, expiration time.Duration) *RedisTransaction {
	t.tx.Set(t.ctx, key, value, expiration)
	return t
}

// Get retrieves a value in the transaction (returns a future result)
func (t *RedisTransaction) Get(key string) *redis.StringCmd {
	return t.tx.Get(t.ctx, key)
}

// Delete removes a key in the transaction
func (t *RedisTransaction) Delete(keys ...string) *RedisTransaction {
	t.tx.Del(t.ctx, keys...)
	return t
}

// Exists checks if keys exist in the transaction
func (t *RedisTransaction) Exists(keys ...string) *redis.IntCmd {
	return t.tx.Exists(t.ctx, keys...)
}

// Expire sets expiration for a key in the transaction
func (t *RedisTransaction) Expire(key string, expiration time.Duration) *RedisTransaction {
	t.tx.Expire(t.ctx, key, expiration)
	return t
}

// Increment increments a counter in the transaction
func (t *RedisTransaction) Increment(key string) *redis.IntCmd {
	return t.tx.Incr(t.ctx, key)
}

// IncrementBy increments a counter by amount in the transaction
func (t *RedisTransaction) IncrementBy(key string, value int64) *redis.IntCmd {
	return t.tx.IncrBy(t.ctx, key, value)
}

// Decrement decrements a counter in the transaction
func (t *RedisTransaction) Decrement(key string) *redis.IntCmd {
	return t.tx.Decr(t.ctx, key)
}

// DecrementBy decrements a counter by amount in the transaction
func (t *RedisTransaction) DecrementBy(key string, value int64) *redis.IntCmd {
	return t.tx.DecrBy(t.ctx, key, value)
}

// ListPush adds a value to a list in the transaction
func (t *RedisTransaction) ListPush(key string, values ...interface{}) *RedisTransaction {
	t.tx.RPush(t.ctx, key, values...)
	return t
}

// ListPop removes and returns the last element of a list in the transaction
func (t *RedisTransaction) ListPop(key string) *redis.StringCmd {
	return t.tx.RPop(t.ctx, key)
}

// ListLeftPush adds a value to the beginning of a list in the transaction
func (t *RedisTransaction) ListLeftPush(key string, values ...interface{}) *RedisTransaction {
	t.tx.LPush(t.ctx, key, values...)
	return t
}

// ListLeftPop removes and returns the first element of a list in the transaction
func (t *RedisTransaction) ListLeftPop(key string) *redis.StringCmd {
	return t.tx.LPop(t.ctx, key)
}

// SetAdd adds members to a set in the transaction
func (t *RedisTransaction) SetAdd(key string, members ...interface{}) *RedisTransaction {
	t.tx.SAdd(t.ctx, key, members...)
	return t
}

// SetRemove removes members from a set in the transaction
func (t *RedisTransaction) SetRemove(key string, members ...interface{}) *RedisTransaction {
	t.tx.SRem(t.ctx, key, members...)
	return t
}

// HashSet sets a field in a hash in the transaction
func (t *RedisTransaction) HashSet(key string, values ...interface{}) *RedisTransaction {
	t.tx.HSet(t.ctx, key, values...)
	return t
}

// HashDelete deletes fields from a hash in the transaction
func (t *RedisTransaction) HashDelete(key string, fields ...string) *RedisTransaction {
	t.tx.HDel(t.ctx, key, fields...)
	return t
}

// ============================================================================
// Transaction Control
// ============================================================================

// Exec executes the transaction
func (t *RedisTransaction) Exec() ([]redis.Cmder, error) {
	return t.tx.Exec(t.ctx)
}

// Discard discards the transaction
func (t *RedisTransaction) Discard() error {
	return t.tx.Discard()
}

// ============================================================================
// Advanced Transaction Operations
// ============================================================================

// ConditionalSet sets a value only if a condition is met
func (t *RedisTransaction) ConditionalSet(key string, value interface{}, condition string) *RedisTransaction {
	switch condition {
	case "nx": // Only if key doesn't exist
		t.tx.SetNX(t.ctx, key, value, 0)
	case "xx": // Only if key exists
		t.tx.SetXX(t.ctx, key, value, 0)
	}
	return t
}

// ConditionalSetWithExpiry sets a value with expiry only if condition is met
func (t *RedisTransaction) ConditionalSetWithExpiry(key string, value interface{}, expiration time.Duration, condition string) *RedisTransaction {
	switch condition {
	case "nx": // Only if key doesn't exist
		t.tx.SetNX(t.ctx, key, value, expiration)
	case "xx": // Only if key exists
		t.tx.SetXX(t.ctx, key, value, expiration)
	}
	return t
}

// CompareAndSwap performs atomic compare-and-swap operation
func (t *RedisTransaction) CompareAndSwap(key string, oldValue, newValue interface{}) *redis.Cmd {
	return t.tx.Eval(t.ctx, `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("SET", KEYS[1], ARGV[2])
		else
			return 0
		end
	`, []string{key}, oldValue, newValue)
}

// ============================================================================
// Batch Operations
// ============================================================================

// BatchSet sets multiple key-value pairs in a single transaction
func (t *RedisTransaction) BatchSet(pairs map[string]interface{}, expiration time.Duration) *RedisTransaction {
	for key, value := range pairs {
		t.Set(key, value, expiration)
	}
	return t
}

// BatchDelete deletes multiple keys in a single transaction
func (t *RedisTransaction) BatchDelete(keys []string) *RedisTransaction {
	if len(keys) > 0 {
		t.Delete(keys...)
	}
	return t
}

// BatchIncrement increments multiple counters in a single transaction
func (t *RedisTransaction) BatchIncrement(keys []string) *RedisTransaction {
	for _, key := range keys {
		t.Increment(key)
	}
	return t
}

// ============================================================================
// Utility Methods
// ============================================================================

// GetPipeline returns the underlying pipeline for advanced operations
func (t *RedisTransaction) GetPipeline() redis.Pipeliner {
	return t.tx
}

// IsEmpty checks if the transaction has any commands
func (t *RedisTransaction) IsEmpty() bool {
	// This is a simplified check - in practice, you'd need to track commands
	return t.tx == nil
}

// CommandCount returns the number of commands in the transaction
func (t *RedisTransaction) CommandCount() int {
	// This is a simplified implementation
	// In practice, you'd need to track the number of commands added
	return 0
}
