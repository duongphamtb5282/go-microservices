package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"backend-core/cache"
	"backend-core/cache/redis/reload"
	"backend-core/config"
)

func main() {
	// Create Redis configuration
	cfg := &config.RedisConfig{
		Name:       "reload-example",
		Addr:       "localhost:6379",
		Password:   "",
		DB:         0,
		UseCluster: false,
		PoolSize:   10,
	}

	// Create Redis cache instance
	redisCache := cache.NewRedisCache(cfg)
	defer redisCache.Close()

	ctx := context.Background()

	// Test connection
	if err := redisCache.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping Redis: %v", err)
	}

	fmt.Println("✅ Connected to Redis")

	// Example 1: Basic Cache Reloading
	fmt.Println("\n=== Example 1: Basic Cache Reloading ===")
	basicReloadExample(ctx, redisCache)

	// Example 2: Cache Invalidation
	fmt.Println("\n=== Example 2: Cache Invalidation ===")
	invalidationExample(ctx, redisCache)

	// Example 3: Cache Warming
	fmt.Println("\n=== Example 3: Cache Warming ===")
	warmingExample(ctx, redisCache)

	// Example 4: Scheduled Reloading
	fmt.Println("\n=== Example 4: Scheduled Reloading ===")
	scheduledReloadExample(ctx, redisCache)

	// Example 5: Different Data Sources
	fmt.Println("\n=== Example 5: Different Data Sources ===")
	dataSourceExample(ctx, redisCache)

	// Example 6: Cache Reloading Strategies
	fmt.Println("\n=== Example 6: Cache Reloading Strategies ===")
	strategyExample(ctx, redisCache)
}

func basicReloadExample(ctx context.Context, redisCache *cache.RedisCache) {
	// Create mock data source
	dataSource := reload.NewMockDataSource()
	dataSource.SetData("user:1", map[string]interface{}{
		"id":    1,
		"name":  "Alice",
		"email": "alice@example.com",
	})
	dataSource.SetData("user:2", map[string]interface{}{
		"id":    2,
		"name":  "Bob",
		"email": "bob@example.com",
	})

	// Create reload configuration
	reloadConfig := &reload.CacheReloadConfig{
		Strategy:          reload.StrategyRefresh,
		TTL:               time.Hour,
		BatchSize:         10,
		MaxRetries:        3,
		RetryDelay:        time.Second,
		EnableLazyLoading: false,
	}

	// Create reloader
	reloader := reload.NewRedisCacheReloader(redisCache.GetClient(), reloadConfig, dataSource)
	redisCache.SetReloader(reloader)

	// Create invalidator
	invalidator := reload.NewRedisCacheInvalidator(redisCache.GetClient(), reloader)
	redisCache.SetInvalidator(invalidator)

	// Reload specific keys
	fmt.Println("🔄 Reloading specific keys...")
	if err := redisCache.Reload(ctx, "user:1"); err != nil {
		log.Printf("❌ Failed to reload user:1: %v", err)
		return
	}

	// Verify reloaded data
	var user1 map[string]interface{}
	if err := redisCache.Get(ctx, "user:1", &user1); err == nil {
		fmt.Printf("   ✅ Reloaded user:1: %+v\n", user1)
	}

	// Reload batch of keys
	fmt.Println("🔄 Reloading batch of keys...")
	keys := []string{"user:1", "user:2"}
	if err := redisCache.ReloadBatch(ctx, keys); err != nil {
		log.Printf("❌ Failed to reload batch: %v", err)
		return
	}

	// Verify batch reloaded data
	for _, key := range keys {
		var user map[string]interface{}
		if err := redisCache.Get(ctx, key, &user); err == nil {
			fmt.Printf("   ✅ Reloaded %s: %+v\n", key, user)
		}
	}
}

func invalidationExample(ctx context.Context, redisCache *cache.RedisCache) {
	// Set some test data
	redisCache.Set(ctx, "test:key1", "value1", time.Hour)
	redisCache.Set(ctx, "test:key2", "value2", time.Hour)
	redisCache.Set(ctx, "test:key3", "value3", time.Hour)

	fmt.Println("📝 Set test data")

	// Invalidate single key
	fmt.Println("🗑️ Invalidating single key...")
	if err := redisCache.Invalidate(ctx, "test:key1"); err != nil {
		log.Printf("❌ Failed to invalidate test:key1: %v", err)
		return
	}

	// Check if key is invalidated
	exists, err := redisCache.Exists(ctx, "test:key1")
	if err != nil {
		log.Printf("❌ Failed to check existence: %v", err)
		return
	}
	fmt.Printf("   test:key1 exists: %v\n", exists)

	// Invalidate pattern
	fmt.Println("🗑️ Invalidating pattern test:key*...")
	if err := redisCache.InvalidatePattern(ctx, "test:key*"); err != nil {
		log.Printf("❌ Failed to invalidate pattern: %v", err)
		return
	}

	// Check remaining keys
	for i := 1; i <= 3; i++ {
		exists, err := redisCache.Exists(ctx, fmt.Sprintf("test:key%d", i))
		if err != nil {
			log.Printf("❌ Failed to check existence: %v", err)
			return
		}
		fmt.Printf("   test:key%d exists: %v\n", i, exists)
	}
}

func warmingExample(ctx context.Context, redisCache *cache.RedisCache) {
	// Create mock data source with more data
	dataSource := reload.NewMockDataSource()
	for i := 1; i <= 5; i++ {
		dataSource.SetData(fmt.Sprintf("warm:key%d", i), map[string]interface{}{
			"id":   i,
			"data": fmt.Sprintf("warm_data_%d", i),
		})
	}

	// Create reload configuration
	reloadConfig := &reload.CacheReloadConfig{
		Strategy:     reload.StrategyRefresh,
		TTL:          time.Hour,
		BatchSize:    10,
		EnableWarmUp: true,
		WarmUpKeys:   []string{"warm:key1", "warm:key2", "warm:key3"},
	}

	// Create reloader and warmer
	reloader := reload.NewRedisCacheReloader(redisCache.GetClient(), reloadConfig, dataSource)
	warmer := reload.NewRedisCacheWarmer(redisCache.GetClient(), reloader, reloadConfig)

	redisCache.SetReloader(reloader)
	redisCache.SetWarmer(warmer)

	// Check if cache is warmed up
	isWarmedUp, err := redisCache.IsWarmedUp(ctx)
	if err != nil {
		log.Printf("❌ Failed to check warm-up status: %v", err)
		return
	}
	fmt.Printf("Cache warmed up: %v\n", isWarmedUp)

	// Warm up cache
	fmt.Println("🔥 Warming up cache...")
	if err := redisCache.WarmUp(ctx); err != nil {
		log.Printf("❌ Failed to warm up cache: %v", err)
		return
	}

	// Check if cache is warmed up
	isWarmedUp, err = redisCache.IsWarmedUp(ctx)
	if err != nil {
		log.Printf("❌ Failed to check warm-up status: %v", err)
		return
	}
	fmt.Printf("Cache warmed up: %v\n", isWarmedUp)

	// Verify warmed up data
	for i := 1; i <= 3; i++ {
		key := fmt.Sprintf("warm:key%d", i)
		var data map[string]interface{}
		if err := redisCache.Get(ctx, key, &data); err == nil {
			fmt.Printf("   ✅ Warmed up %s: %+v\n", key, data)
		}
	}
}

func scheduledReloadExample(ctx context.Context, redisCache *cache.RedisCache) {
	// Create mock data source
	dataSource := reload.NewMockDataSource()
	dataSource.SetData("scheduled:key", map[string]interface{}{
		"id":        1,
		"data":      "scheduled_data",
		"timestamp": time.Now().Unix(),
	})

	// Create reload configuration with scheduled reloading
	reloadConfig := &reload.CacheReloadConfig{
		Strategy:              reload.StrategyRefresh,
		TTL:                   time.Hour,
		BatchSize:             10,
		EnableScheduledReload: true,
		ReloadInterval:        5 * time.Second,
	}

	// Create reloader
	reloader := reload.NewRedisCacheReloader(redisCache.GetClient(), reloadConfig, dataSource)
	redisCache.SetReloader(reloader)

	fmt.Println("⏰ Starting scheduled reloading (5 second intervals)...")
	fmt.Println("   Press Ctrl+C to stop")

	// Let it run for a while
	time.Sleep(15 * time.Second)

	// Stop the scheduled reloader
	reloader.Stop()
	fmt.Println("🛑 Stopped scheduled reloading")
}

func dataSourceExample(ctx context.Context, redisCache *cache.RedisCache) {
	// Test with different data sources
	fmt.Println("🗄️ Testing Database Data Source...")
	databaseSource := reload.NewMockDataSource()
	databaseSource.SetData("db:user:1", map[string]interface{}{
		"id":    1,
		"name":  "Database User",
		"email": "db@example.com",
	})

	reloadConfig := &reload.CacheReloadConfig{
		Strategy:  reload.StrategyRefresh,
		TTL:       time.Hour,
		BatchSize: 10,
	}

	reloader := reload.NewRedisCacheReloader(redisCache.GetClient(), reloadConfig, databaseSource)
	redisCache.SetReloader(reloader)

	// Reload from database source
	if err := redisCache.Reload(ctx, "db:user:1"); err != nil {
		log.Printf("❌ Failed to reload from database: %v", err)
		return
	}

	var user map[string]interface{}
	if err := redisCache.Get(ctx, "db:user:1", &user); err == nil {
		fmt.Printf("   ✅ Database user: %+v\n", user)
	}

	fmt.Println("🌐 Testing API Data Source...")
	apiSource := reload.NewMockDataSource()
	apiSource.SetData("api:user:1", map[string]interface{}{
		"id":     1,
		"name":   "API User",
		"email":  "api@example.com",
		"status": "active",
	})

	reloader.SetDataSource(apiSource)

	// Reload from API source
	if err := redisCache.Reload(ctx, "api:user:1"); err != nil {
		log.Printf("❌ Failed to reload from API: %v", err)
		return
	}

	if err := redisCache.Get(ctx, "api:user:1", &user); err == nil {
		fmt.Printf("   ✅ API user: %+v\n", user)
	}
}

func strategyExample(ctx context.Context, redisCache *cache.RedisCache) {
	// Test different reload strategies
	strategies := []reload.CacheReloadStrategy{
		reload.StrategyRefresh,
		reload.StrategyReplace,
		reload.StrategyLazy,
	}

	for _, strategy := range strategies {
		fmt.Printf("🔄 Testing %s strategy...\n", strategy)

		// Create data source
		dataSource := reload.NewMockDataSource()
		dataSource.SetData("strategy:key", map[string]interface{}{
			"strategy":  string(strategy),
			"timestamp": time.Now().Unix(),
		})

		// Create reload configuration
		reloadConfig := &reload.CacheReloadConfig{
			Strategy:  strategy,
			TTL:       time.Hour,
			BatchSize: 10,
		}

		// Create reloader
		reloader := reload.NewRedisCacheReloader(redisCache.GetClient(), reloadConfig, dataSource)
		redisCache.SetReloader(reloader)

		// Reload data
		if err := redisCache.Reload(ctx, "strategy:key"); err != nil {
			log.Printf("❌ Failed to reload with %s strategy: %v", strategy, err)
			continue
		}

		// Verify data
		var data map[string]interface{}
		if err := redisCache.Get(ctx, "strategy:key", &data); err == nil {
			fmt.Printf("   ✅ %s strategy data: %+v\n", strategy, data)
		}

		// Clean up
		redisCache.Delete(ctx, "strategy:key")
	}
}
