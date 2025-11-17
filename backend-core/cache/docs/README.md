# Redis Cache Architecture

## 🎯 Single Responsibility Principle (SRP) Implementation

The Redis cache implementation has been refactored to follow the Single Responsibility Principle, with each file having a single, well-defined responsibility.

## 📁 File Structure

```
cache/
├── interfaces.go          # Cache interface definitions
├── redis.go              # Main RedisCache facade (delegates to specialized handlers)
├── redis_client.go       # Redis client creation and configuration
├── redis_operations.go   # Basic Redis operations (CRUD, TTL, etc.)
├── redis_counters.go     # Counter operations (increment, decrement)
├── redis_lists.go        # List operations (push, pop, length)
├── redis_sets.go         # Set operations (add, members, intersection)
├── redis_hashes.go       # Hash operations (set, get, delete fields)
└── README.md            # This documentation
```

## 🏗️ Architecture Overview

### **1. RedisClientFactory** (`redis_client.go`)

**Responsibility**: Create Redis clients based on configuration

- ✅ Creates standalone Redis clients
- ✅ Creates cluster Redis clients
- ✅ Handles client configuration
- ✅ Provides convenience methods

### **2. RedisOperations** (`redis_operations.go`)

**Responsibility**: Basic Redis operations

- ✅ Set/Get/Delete operations
- ✅ Key existence checks
- ✅ TTL management
- ✅ Pattern-based deletion
- ✅ Connection testing (Ping)

### **3. RedisCounters** (`redis_counters.go`)

**Responsibility**: Counter operations

- ✅ Increment/Decrement operations
- ✅ Increment/Decrement by specific amounts
- ✅ Atomic counter operations

### **4. RedisLists** (`redis_lists.go`)

**Responsibility**: List operations

- ✅ Push/Pop operations (left and right)
- ✅ List length operations
- ✅ List range operations
- ✅ Queue/Stack functionality

### **5. RedisSets** (`redis_sets.go`)

**Responsibility**: Set operations

- ✅ Add/Remove members
- ✅ Member existence checks
- ✅ Set operations (union, intersection)
- ✅ Set cardinality

### **6. RedisHashes** (`redis_hashes.go`)

**Responsibility**: Hash operations

- ✅ Field operations (set, get, delete)
- ✅ Hash-wide operations (get all, keys, values)
- ✅ Field existence checks
- ✅ Hash length operations

### **7. RedisCache** (`redis.go`)

**Responsibility**: Facade that composes specialized handlers

- ✅ Implements the main Cache interface
- ✅ Delegates operations to specialized handlers
- ✅ Manages connection lifecycle
- ✅ Provides configuration access

## 🔄 Design Patterns Used

### **1. Facade Pattern**

- `RedisCache` acts as a facade, providing a simple interface to complex Redis operations
- Hides the complexity of multiple specialized handlers

### **2. Composition Pattern**

- `RedisCache` composes specialized handlers instead of inheriting
- Each handler is responsible for a specific domain

### **3. Factory Pattern**

- `RedisClientFactory` creates appropriate Redis clients based on configuration
- Supports both standalone and cluster modes

### **4. Delegation Pattern**

- All operations are delegated to appropriate specialized handlers
- Clean separation of concerns

## 🎯 Benefits of This Architecture

### **1. Single Responsibility Principle (SRP)**

- ✅ Each file has one reason to change
- ✅ Clear separation of concerns
- ✅ Easy to understand and maintain

### **2. Open/Closed Principle (OCP)**

- ✅ Easy to add new Redis data types without modifying existing code
- ✅ Can extend functionality by adding new handlers

### **3. Interface Segregation Principle (ISP)**

- ✅ Clients only depend on interfaces they use
- ✅ Specialized handlers can be used independently

### **4. Dependency Inversion Principle (DIP)**

- ✅ High-level modules don't depend on low-level modules
- ✅ Both depend on abstractions (interfaces)

## 🚀 Usage Examples

### **Basic Usage**

```go
// Create Redis cache
cfg := &config.RedisConfig{
    Name: "my-redis",
    Addr: "localhost:6379",
    // ... other config
}
redisCache := cache.NewRedisCache(cfg)
defer redisCache.Close()

// Use basic operations
redisCache.Set(ctx, "key", "value", time.Hour)
redisCache.Get(ctx, "key", &result)
```

### **Specialized Operations**

```go
// Access specialized handlers directly (if needed)
ops := cache.NewRedisOperations(redisCache.GetClient())
counters := cache.NewRedisCounters(redisCache.GetClient())
lists := cache.NewRedisLists(redisCache.GetClient())
sets := cache.NewRedisSets(redisCache.GetClient())
hashes := cache.NewRedisHashes(redisCache.GetClient())
```

## 🔧 Maintenance Benefits

1. **Easy Testing**: Each handler can be tested independently
2. **Easy Debugging**: Issues are isolated to specific handlers
3. **Easy Extension**: Add new Redis data types without touching existing code
4. **Easy Refactoring**: Changes to one handler don't affect others
5. **Clear Documentation**: Each file's purpose is immediately clear

## 📈 Performance Benefits

1. **Reduced Memory**: Only load handlers you need
2. **Better Caching**: Specialized handlers can implement their own optimizations
3. **Parallel Development**: Multiple developers can work on different handlers
4. **Selective Updates**: Update only the handlers that need changes

This architecture makes the Redis cache implementation much more maintainable, testable, and extensible while following SOLID principles! 🎉
