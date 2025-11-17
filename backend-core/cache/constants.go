package cache

// Cache Profile Constants
const (
	// ProfileDefault represents the default cache profile
	ProfileDefault = "default"
	
	// ProfileHighPerformance represents the high-performance cache profile
	ProfileHighPerformance = "high-performance"
	
	// ProfileConsistency represents the consistency-optimized cache profile
	ProfileConsistency = "consistency"
	
	// ProfileReadHeavy represents the read-heavy workload cache profile
	ProfileReadHeavy = "read-heavy"
	
	// ProfileWriteHeavy represents the write-heavy workload cache profile
	ProfileWriteHeavy = "write-heavy"
	
	// ProfileBalanced represents the balanced workload cache profile
	ProfileBalanced = "balanced"
	
	// ProfileLowCache represents the low-cache profile
	ProfileLowCache = "low-cache"
	
	// ProfileNoCache represents the no-cache profile
	ProfileNoCache = "no-cache"
)

// Cache Strategy Constants
const (
	// StrategyReadThrough represents the read-through cache strategy
	StrategyReadThrough = "read-through"
	
	// StrategyWriteThrough represents the write-through cache strategy
	StrategyWriteThrough = "write-through"
	
	// StrategyWriteBehind represents the write-behind cache strategy
	StrategyWriteBehind = "write-behind"
	
	// StrategyCacheAside represents the cache-aside strategy
	StrategyCacheAside = "cache-aside"
)

// Cache Key Prefix Constants
const (
	// KeyPrefixDefault represents the default key prefix
	KeyPrefixDefault = "default"
	
	// KeyPrefixPerformance represents the performance key prefix
	KeyPrefixPerformance = "perf"
	
	// KeyPrefixConsistency represents the consistency key prefix
	KeyPrefixConsistency = "consistency"
	
	// KeyPrefixRead represents the read-heavy key prefix
	KeyPrefixRead = "read"
	
	// KeyPrefixWrite represents the write-heavy key prefix
	KeyPrefixWrite = "write"
)

// Cache TTL Constants (in minutes)
const (
	// TTLDefault represents the default TTL in minutes
	TTLDefault = 5
	
	// TTLShort represents a short TTL in minutes
	TTLShort = 2
	
	// TTLMedium represents a medium TTL in minutes
	TTLMedium = 10
	
	// TTLLong represents a long TTL in minutes
	TTLLong = 30
	
	// TTLVeryLong represents a very long TTL in minutes
	TTLVeryLong = 60
	
	// TTLUltraLong represents an ultra long TTL in minutes
	TTLUltraLong = 90
)

// Cache Profile Descriptions
var ProfileDescriptions = map[string]string{
	ProfileDefault:          "Balanced configuration with moderate TTL and write-through strategy",
	ProfileHighPerformance: "Optimized for high performance with longer TTL and cache-aside strategy",
	ProfileConsistency:      "Optimized for data consistency with shorter TTL and write-through strategy",
	ProfileReadHeavy:        "Optimized for read-heavy workloads with long TTL and read-through strategy",
	ProfileWriteHeavy:       "Optimized for write-heavy workloads with short TTL and write-behind strategy",
	ProfileBalanced:         "Balanced configuration for mixed read/write workloads",
	ProfileLowCache:         "Minimal caching with short TTLs",
	ProfileNoCache:          "No caching enabled",
}

// Cache Strategy Descriptions
var StrategyDescriptions = map[string]string{
	StrategyReadThrough:  "Read-through: Cache is populated on read miss, data is loaded from source",
	StrategyWriteThrough: "Write-through: Data is written to both cache and source simultaneously",
	StrategyWriteBehind:  "Write-behind: Data is written to cache first, then asynchronously to source",
	StrategyCacheAside:   "Cache-aside: Application manages cache directly, cache and source are separate",
}

// IsValidProfile checks if a profile is valid
func IsValidProfile(profile string) bool {
	_, exists := ProfileDescriptions[profile]
	return exists
}

// IsValidStrategy checks if a strategy is valid
func IsValidStrategy(strategy string) bool {
	_, exists := StrategyDescriptions[strategy]
	return exists
}

// GetAvailableProfiles returns a list of available cache profiles
func GetAvailableProfiles() []string {
	profiles := make([]string, 0, len(ProfileDescriptions))
	for profile := range ProfileDescriptions {
		profiles = append(profiles, profile)
	}
	return profiles
}

// GetAvailableStrategies returns a list of available cache strategies
func GetAvailableStrategies() []string {
	strategies := make([]string, 0, len(StrategyDescriptions))
	for strategy := range StrategyDescriptions {
		strategies = append(strategies, strategy)
	}
	return strategies
}
