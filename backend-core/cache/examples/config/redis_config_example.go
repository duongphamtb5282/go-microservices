package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"backend-core/cache"
	"backend-core/config"
)

func main() {
	// Example 1: Standalone Redis configuration
	fmt.Println("=== Standalone Redis Configuration ===")
	standaloneConfig := &config.RedisConfig{
		Name:       "microservices-redis-standalone",
		Addr:       "localhost:6379",
		Password:   "",
		DB:         0,
		UseCluster: false,
		PoolSize:   10,
	}

	// Validate configuration
	if err := standaloneConfig.Validate(); err != nil {
		log.Fatalf("Standalone config validation failed: %v", err)
	}

	// Create Redis cache instance
	redisCache := cache.NewRedisCache(standaloneConfig)
	defer redisCache.Close()

	// Test connection
	ctx := context.Background()
	if err := redisCache.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping Redis: %v", err)
	}

	fmt.Printf("✅ Connected to standalone Redis at %s\n", standaloneConfig.GetRedisAddr())
	fmt.Printf("   Instance name: %s\n", standaloneConfig.Name)
	fmt.Printf("   Database: %d\n", standaloneConfig.DB)
	fmt.Printf("   Pool size: %d\n", standaloneConfig.PoolSize)
	fmt.Printf("   Cluster mode: %t\n", standaloneConfig.IsCluster())

	// Test basic operations
	testBasicOperations(ctx, redisCache)

	// Example 2: Cluster Redis configuration
	fmt.Println("\n=== Cluster Redis Configuration ===")
	clusterConfig := &config.RedisConfig{
		Name:       "microservices-redis-cluster",
		Addr:       "", // Not used in cluster mode
		Password:   "",
		DB:         0, // Must be 0 in cluster mode
		UseCluster: true,
		ClusterAddrs: []string{
			"localhost:7000",
			"localhost:7001",
			"localhost:7002",
			"localhost:7003",
			"localhost:7004",
			"localhost:7005",
		},
		PoolSize: 20,
	}

	// Validate configuration
	if err := clusterConfig.Validate(); err != nil {
		log.Fatalf("Cluster config validation failed: %v", err)
	}

	// Create Redis cluster cache instance
	clusterCache := cache.NewRedisCache(clusterConfig)
	defer clusterCache.Close()

	// Test connection
	if err := clusterCache.Ping(ctx); err != nil {
		log.Printf("⚠️  Failed to ping Redis cluster (cluster might not be running): %v", err)
	} else {
		fmt.Printf("✅ Connected to Redis cluster with %d nodes\n", len(clusterConfig.GetClusterAddrs()))
		fmt.Printf("   Instance name: %s\n", clusterConfig.Name)
		fmt.Printf("   Cluster addresses: %v\n", clusterConfig.GetClusterAddrs())
		fmt.Printf("   Pool size: %d\n", clusterConfig.PoolSize)
		fmt.Printf("   Cluster mode: %t\n", clusterConfig.IsCluster())

		// Test basic operations
		testBasicOperations(ctx, clusterCache)
	}

	// Example 3: Simple standalone Redis (convenience function)
	fmt.Println("\n=== Simple Standalone Redis ===")
	simpleRedis := cache.NewStandaloneRedisCache("localhost:6379", "", 0)
	defer simpleRedis.Close()

	if err := simpleRedis.Ping(ctx); err != nil {
		log.Printf("⚠️  Failed to ping simple Redis: %v", err)
	} else {
		fmt.Println("✅ Connected to simple standalone Redis")
		fmt.Printf("   Address: localhost:6379\n")
		fmt.Printf("   Database: 0\n")
		fmt.Printf("   Pool size: 10 (default)\n")
	}

	// Example 4: Configuration from YAML
	fmt.Println("\n=== Configuration from YAML ===")
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Printf("Redis config from YAML:\n")
	fmt.Printf("   Name: %s\n", cfg.Redis.Name)
	fmt.Printf("   Address: %s\n", cfg.Redis.Addr)
	fmt.Printf("   Database: %d\n", cfg.Redis.DB)
	fmt.Printf("   Use cluster: %t\n", cfg.Redis.UseCluster)
	fmt.Printf("   Pool size: %d\n", cfg.Redis.PoolSize)

	if cfg.Redis.UseCluster {
		fmt.Printf("   Cluster addresses: %v\n", cfg.Redis.ClusterAddrs)
	}
}

func testBasicOperations(ctx context.Context, redisCache *cache.RedisCache) {
	fmt.Println("\n--- Testing Basic Operations ---")

	// Test Set
	key := "test:key"
	value := map[string]interface{}{
		"message":   "Hello Redis!",
		"timestamp": time.Now().Unix(),
		"mode":      "standalone",
	}

	if err := redisCache.Set(ctx, key, value, 5*time.Minute); err != nil {
		log.Printf("❌ Set failed: %v", err)
		return
	}
	fmt.Println("✅ Set operation successful")

	// Test Get
	var retrievedValue map[string]interface{}
	if err := redisCache.Get(ctx, key, &retrievedValue); err != nil {
		log.Printf("❌ Get failed: %v", err)
		return
	}
	fmt.Printf("✅ Get operation successful: %+v\n", retrievedValue)

	// Test Exists
	exists, err := redisCache.Exists(ctx, key)
	if err != nil {
		log.Printf("❌ Exists check failed: %v", err)
		return
	}
	fmt.Printf("✅ Key exists: %t\n", exists)

	// Test TTL
	ttl, err := redisCache.TTL(ctx, key)
	if err != nil {
		log.Printf("❌ TTL check failed: %v", err)
		return
	}
	fmt.Printf("✅ TTL: %v\n", ttl)

	// Test Delete
	if err := redisCache.Delete(ctx, key); err != nil {
		log.Printf("❌ Delete failed: %v", err)
		return
	}
	fmt.Println("✅ Delete operation successful")

	// Test List operations
	fmt.Println("\n--- Testing List Operations ---")
	listKey := "test:list"

	// Push items to list
	items := []string{"item1", "item2", "item3"}
	for _, item := range items {
		if err := redisCache.ListPush(ctx, listKey, item); err != nil {
			log.Printf("❌ ListPush failed: %v", err)
			return
		}
	}
	fmt.Println("✅ ListPush operations successful")

	// Get list length
	length, err := redisCache.ListLength(ctx, listKey)
	if err != nil {
		log.Printf("❌ ListLength failed: %v", err)
		return
	}
	fmt.Printf("✅ List length: %d\n", length)

	// Pop items from list
	for i := 0; i < len(items); i++ {
		var poppedItem string
		if err := redisCache.ListPop(ctx, listKey, &poppedItem); err != nil {
			log.Printf("❌ ListPop failed: %v", err)
			return
		}
		fmt.Printf("✅ Popped item: %s\n", poppedItem)
	}

	// Test Set operations
	fmt.Println("\n--- Testing Set Operations ---")
	setKey := "test:set"

	// Add members to set
	setMembers := []string{"member1", "member2", "member3"}
	for _, member := range setMembers {
		if err := redisCache.SetAdd(ctx, setKey, member); err != nil {
			log.Printf("❌ SetAdd failed: %v", err)
			return
		}
	}
	fmt.Println("✅ SetAdd operations successful")

	// Get all members
	members, err := redisCache.SetMembers(ctx, setKey)
	if err != nil {
		log.Printf("❌ SetMembers failed: %v", err)
		return
	}
	fmt.Printf("✅ Set members: %v\n", members)

	// Test Hash operations
	fmt.Println("\n--- Testing Hash Operations ---")
	hashKey := "test:hash"

	// Set hash fields
	hashFields := map[string]interface{}{
		"field1": "value1",
		"field2": "value2",
		"field3": "value3",
	}

	for field, value := range hashFields {
		if err := redisCache.HashSet(ctx, hashKey, field, value); err != nil {
			log.Printf("❌ HashSet failed: %v", err)
			return
		}
	}
	fmt.Println("✅ HashSet operations successful")

	// Get all hash fields
	allFields, err := redisCache.HashGetAll(ctx, hashKey)
	if err != nil {
		log.Printf("❌ HashGetAll failed: %v", err)
		return
	}
	fmt.Printf("✅ Hash fields: %+v\n", allFields)

	// Cleanup
	redisCache.Delete(ctx, listKey)
	redisCache.Delete(ctx, setKey)
	redisCache.Delete(ctx, hashKey)
	fmt.Println("✅ Cleanup completed")
}
