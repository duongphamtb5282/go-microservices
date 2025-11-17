package global

import (
	"fmt"
	"sync"

	"backend-core/database/gorm"
)

var (
	// Global database registry
	GVA_DB     gorm.Database
	GVA_DBList map[string]gorm.Database
	lock       sync.RWMutex
)

// GetGlobalDBByDBName gets a database by name
func GetGlobalDBByDBName(dbname string) gorm.Database {
	lock.RLock()
	defer lock.RUnlock()
	return GVA_DBList[dbname]
}

// MustGetGlobalDBByDBName gets a database by name, panics if not found
func MustGetGlobalDBByDBName(dbname string) gorm.Database {
	lock.RLock()
	defer lock.RUnlock()
	db, ok := GVA_DBList[dbname]
	if !ok || db == nil {
		panic(fmt.Sprintf("database %s not found", dbname))
	}
	return db
}

// SetGlobalDB sets the global database
func SetGlobalDB(db gorm.Database) {
	lock.Lock()
	defer lock.Unlock()
	GVA_DB = db
}

// SetGlobalDBList sets the global database list
func SetGlobalDBList(dbList map[string]gorm.Database) {
	lock.Lock()
	defer lock.Unlock()
	GVA_DBList = dbList
}

// GetGlobalDBList returns the global database list
func GetGlobalDBList() map[string]gorm.Database {
	lock.RLock()
	defer lock.RUnlock()
	return GVA_DBList
}

// GetDefaultDB returns the default database (system)
func GetDefaultDB() gorm.Database {
	return GetGlobalDBByDBName("system")
}
