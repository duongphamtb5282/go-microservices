package reload

import (
	"context"
	"fmt"
	"time"
)

// MockDataSource is a mock data source for testing
type MockDataSource struct {
	data map[string]interface{}
}

// NewMockDataSource creates a new mock data source
func NewMockDataSource() *MockDataSource {
	return &MockDataSource{
		data: make(map[string]interface{}),
	}
}

// SetData sets data in the mock source
func (m *MockDataSource) SetData(key string, value interface{}) {
	m.data[key] = value
}

// LoadData loads data from the mock source
func (m *MockDataSource) LoadData(ctx context.Context, key string) (interface{}, error) {
	// Simulate some processing time
	time.Sleep(10 * time.Millisecond)

	if value, exists := m.data[key]; exists {
		return value, nil
	}

	return nil, fmt.Errorf("key %s not found", key)
}

// LoadDataBatch loads multiple data items from the mock source
func (m *MockDataSource) LoadDataBatch(ctx context.Context, keys []string) (map[string]interface{}, error) {
	// Simulate some processing time
	time.Sleep(50 * time.Millisecond)

	result := make(map[string]interface{})
	for _, key := range keys {
		if value, exists := m.data[key]; exists {
			result[key] = value
		}
	}

	return result, nil
}

// LoadAllData loads all data from the mock source
func (m *MockDataSource) LoadAllData(ctx context.Context) (map[string]interface{}, error) {
	// Simulate some processing time
	time.Sleep(100 * time.Millisecond)

	result := make(map[string]interface{})
	for key, value := range m.data {
		result[key] = value
	}

	return result, nil
}

// GetDataKeys returns all available data keys
func (m *MockDataSource) GetDataKeys(ctx context.Context) ([]string, error) {
	keys := make([]string, 0, len(m.data))
	for key := range m.data {
		keys = append(keys, key)
	}
	return keys, nil
}

// ValidateData validates loaded data
func (m *MockDataSource) ValidateData(data interface{}) error {
	if data == nil {
		return fmt.Errorf("data cannot be nil")
	}
	return nil
}

// DatabaseDataSource simulates a database data source
type DatabaseDataSource struct {
	// In a real implementation, this would contain database connection
	// For now, we'll use a mock data store
	data map[string]interface{}
}

// NewDatabaseDataSource creates a new database data source
func NewDatabaseDataSource() *DatabaseDataSource {
	return &DatabaseDataSource{
		data: make(map[string]interface{}),
	}
}

// SetData sets data in the database source
func (d *DatabaseDataSource) SetData(key string, value interface{}) {
	d.data[key] = value
}

// LoadData loads data from the database
func (d *DatabaseDataSource) LoadData(ctx context.Context, key string) (interface{}, error) {
	// Simulate database query time
	time.Sleep(20 * time.Millisecond)

	if value, exists := d.data[key]; exists {
		return value, nil
	}

	return nil, fmt.Errorf("record with key %s not found", key)
}

// LoadDataBatch loads multiple data items from the database
func (d *DatabaseDataSource) LoadDataBatch(ctx context.Context, keys []string) (map[string]interface{}, error) {
	// Simulate batch database query time
	time.Sleep(100 * time.Millisecond)

	result := make(map[string]interface{})
	for _, key := range keys {
		if value, exists := d.data[key]; exists {
			result[key] = value
		}
	}

	return result, nil
}

// LoadAllData loads all data from the database
func (d *DatabaseDataSource) LoadAllData(ctx context.Context) (map[string]interface{}, error) {
	// Simulate full table scan time
	time.Sleep(200 * time.Millisecond)

	result := make(map[string]interface{})
	for key, value := range d.data {
		result[key] = value
	}

	return result, nil
}

// GetDataKeys returns all available data keys
func (d *DatabaseDataSource) GetDataKeys(ctx context.Context) ([]string, error) {
	keys := make([]string, 0, len(d.data))
	for key := range d.data {
		keys = append(keys, key)
	}
	return keys, nil
}

// ValidateData validates loaded data
func (d *DatabaseDataSource) ValidateData(data interface{}) error {
	if data == nil {
		return fmt.Errorf("data cannot be nil")
	}

	// Additional validation for database data
	if dataMap, ok := data.(map[string]interface{}); ok {
		if _, hasID := dataMap["id"]; !hasID {
			return fmt.Errorf("data must have an 'id' field")
		}
	}

	return nil
}

// APIDataSource simulates an external API data source
type APIDataSource struct {
	// In a real implementation, this would contain HTTP client
	// For now, we'll use a mock data store
	data map[string]interface{}
}

// NewAPIDataSource creates a new API data source
func NewAPIDataSource() *APIDataSource {
	return &APIDataSource{
		data: make(map[string]interface{}),
	}
}

// SetData sets data in the API source
func (a *APIDataSource) SetData(key string, value interface{}) {
	a.data[key] = value
}

// LoadData loads data from the API
func (a *APIDataSource) LoadData(ctx context.Context, key string) (interface{}, error) {
	// Simulate API call time
	time.Sleep(100 * time.Millisecond)

	if value, exists := a.data[key]; exists {
		return value, nil
	}

	return nil, fmt.Errorf("API resource with key %s not found", key)
}

// LoadDataBatch loads multiple data items from the API
func (a *APIDataSource) LoadDataBatch(ctx context.Context, keys []string) (map[string]interface{}, error) {
	// Simulate batch API call time
	time.Sleep(300 * time.Millisecond)

	result := make(map[string]interface{})
	for _, key := range keys {
		if value, exists := a.data[key]; exists {
			result[key] = value
		}
	}

	return result, nil
}

// LoadAllData loads all data from the API
func (a *APIDataSource) LoadAllData(ctx context.Context) (map[string]interface{}, error) {
	// Simulate full API call time
	time.Sleep(500 * time.Millisecond)

	result := make(map[string]interface{})
	for key, value := range a.data {
		result[key] = value
	}

	return result, nil
}

// GetDataKeys returns all available data keys
func (a *APIDataSource) GetDataKeys(ctx context.Context) ([]string, error) {
	keys := make([]string, 0, len(a.data))
	for key := range a.data {
		keys = append(keys, key)
	}
	return keys, nil
}

// ValidateData validates loaded data
func (a *APIDataSource) ValidateData(data interface{}) error {
	if data == nil {
		return fmt.Errorf("data cannot be nil")
	}

	// Additional validation for API data
	if dataMap, ok := data.(map[string]interface{}); ok {
		if _, hasStatus := dataMap["status"]; !hasStatus {
			return fmt.Errorf("API data must have a 'status' field")
		}
	}

	return nil
}
