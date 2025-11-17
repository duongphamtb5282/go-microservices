package masking

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	auditconfig "backend-core/audit/config"
)

// FileAuditor implements file-based audit logging
type FileAuditor struct {
	config     auditconfig.Config
	auditLogs  []MaskingAudit
	mutex      sync.RWMutex
	logFile    *os.File
	lastFlush  time.Time
	flushMutex sync.Mutex
}

// NewFileAuditor creates a new file auditor
func NewFileAuditor(config auditconfig.Config) *FileAuditor {
	auditor := &FileAuditor{
		config:    config,
		auditLogs: make([]MaskingAudit, 0),
	}

	// Open log file if specified
	if config.LogLevel != "" {
		logPath := filepath.Join("logs", "masking_audit.log")
		if err := os.MkdirAll(filepath.Dir(logPath), 0755); err == nil {
			if file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
				auditor.logFile = file
			}
		}
	}

	return auditor
}

// LogAudit logs a masking operation
func (a *FileAuditor) LogAudit(audit MaskingAudit) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	a.auditLogs = append(a.auditLogs, audit)

	// Write to file if configured
	if a.logFile != nil {
		a.writeToFile(audit)
	}

	// Flush periodically
	a.flushMutex.Lock()
	if time.Since(a.lastFlush) > 5*time.Minute {
		a.flushToFile()
		a.lastFlush = time.Now()
	}
	a.flushMutex.Unlock()
}

// GetAuditLogs retrieves audit logs
func (a *FileAuditor) GetAuditLogs(filter AuditFilter) ([]MaskingAudit, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	var filteredLogs []MaskingAudit

	for _, audit := range a.auditLogs {
		if a.matchesFilter(audit, filter) {
			filteredLogs = append(filteredLogs, audit)
		}
	}

	// Sort by timestamp (newest first)
	sort.Slice(filteredLogs, func(i, j int) bool {
		return filteredLogs[i].Timestamp.After(filteredLogs[j].Timestamp)
	})

	// Apply limit and offset
	start := filter.Offset
	if start >= len(filteredLogs) {
		return []MaskingAudit{}, nil
	}

	end := start + filter.Limit
	if end > len(filteredLogs) {
		end = len(filteredLogs)
	}

	if filter.Limit <= 0 {
		end = len(filteredLogs)
	}

	return filteredLogs[start:end], nil
}

// ExportAuditLogs exports audit logs
func (a *FileAuditor) ExportAuditLogs(format string) ([]byte, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	switch format {
	case "json":
		return json.MarshalIndent(a.auditLogs, "", "  ")
	case "csv":
		return a.exportCSV(), nil
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// matchesFilter checks if an audit entry matches the filter
func (a *FileAuditor) matchesFilter(audit MaskingAudit, filter AuditFilter) bool {
	if filter.Field != "" && audit.Field != filter.Field {
		return false
	}
	if filter.User != "" && audit.User != filter.User {
		return false
	}
	if filter.Environment != "" && audit.Environment != filter.Environment {
		return false
	}
	if !filter.StartTime.IsZero() && audit.Timestamp.Before(filter.StartTime) {
		return false
	}
	if !filter.EndTime.IsZero() && audit.Timestamp.After(filter.EndTime) {
		return false
	}
	return true
}

// writeToFile writes a single audit entry to file
func (a *FileAuditor) writeToFile(audit MaskingAudit) {
	if a.logFile == nil {
		return
	}

	entry := fmt.Sprintf("[%s] Field: %s, Method: %s, Environment: %s\n",
		audit.Timestamp.Format(time.RFC3339),
		audit.Field,
		audit.Method,
		audit.Environment,
	)

	a.logFile.WriteString(entry)
}

// flushToFile flushes all audit logs to file
func (a *FileAuditor) flushToFile() {
	if a.logFile == nil {
		return
	}

	a.logFile.Sync()
}

// exportCSV exports audit logs as CSV
func (a *FileAuditor) exportCSV() []byte {
	csv := "timestamp,field,original_length,masked_length,method,user,environment,rule\n"

	for _, audit := range a.auditLogs {
		csv += fmt.Sprintf("%s,%s,%d,%d,%s,%s,%s,%s\n",
			audit.Timestamp.Format(time.RFC3339),
			audit.Field,
			audit.OriginalLen,
			audit.MaskedLen,
			audit.Method,
			audit.User,
			audit.Environment,
			audit.Rule,
		)
	}

	return []byte(csv)
}

// Close closes the auditor
func (a *FileAuditor) Close() error {
	if a.logFile != nil {
		return a.logFile.Close()
	}
	return nil
}

// NoOpAuditor implements a no-operation auditor
type NoOpAuditor struct{}

// NewNoOpAuditor creates a new no-operation auditor
func NewNoOpAuditor() *NoOpAuditor {
	return &NoOpAuditor{}
}

// LogAudit does nothing
func (a *NoOpAuditor) LogAudit(audit MaskingAudit) {
	// No-op
}

// GetAuditLogs returns empty slice
func (a *NoOpAuditor) GetAuditLogs(filter AuditFilter) ([]MaskingAudit, error) {
	return []MaskingAudit{}, nil
}

// ExportAuditLogs returns empty data
func (a *NoOpAuditor) ExportAuditLogs(format string) ([]byte, error) {
	return []byte{}, nil
}
