package config

import (
	"fmt"
	"strings"
)

// MongoDBConfig holds MongoDB-specific configuration
type MongoDBConfig struct {
	Coll             string       `json:"coll" yaml:"coll" mapstructure:"coll"`                                           // collection name
	Options          string       `json:"options" yaml:"options" mapstructure:"options"`                                  // mongodb options
	Database         string       `json:"database" yaml:"database" mapstructure:"database"`                               // database name
	Username         string       `json:"username" yaml:"username" mapstructure:"username"`                               // 用户名
	Password         string       `json:"password" yaml:"password" mapstructure:"password"`                               // 密码
	AuthSource       string       `json:"auth-source" yaml:"auth-source" mapstructure:"auth-source"`                      // 验证数据库
	MinPoolSize      uint64       `json:"min-pool-size" yaml:"min-pool-size" mapstructure:"min-pool-size"`                // 最小连接池
	MaxPoolSize      uint64       `json:"max-pool-size" yaml:"max-pool-size" mapstructure:"max-pool-size"`                // 最大连接池
	SocketTimeoutMs  int64        `json:"socket-timeout-ms" yaml:"socket-timeout-ms" mapstructure:"socket-timeout-ms"`    // socket超时时间
	ConnectTimeoutMs int64        `json:"connect-timeout-ms" yaml:"connect-timeout-ms" mapstructure:"connect-timeout-ms"` // 连接超时时间
	IsZap            bool         `json:"is-zap" yaml:"is-zap" mapstructure:"is-zap"`                                     // 是否开启zap日志
	Hosts            []*MongoHost `json:"hosts" yaml:"hosts" mapstructure:"hosts"`                                        // 主机列表
}

// MongoHost represents a MongoDB host configuration
type MongoHost struct {
	Host string `json:"host" yaml:"host" mapstructure:"host"` // ip地址
	Port string `json:"port" yaml:"port" mapstructure:"port"` // 端口
}

// NewMongoDBConfig creates a new MongoDB configuration with defaults
func NewMongoDBConfig() *MongoDBConfig {
	return &MongoDBConfig{
		Database:         "microservices",
		Username:         "",
		Password:         "",
		AuthSource:       "admin",
		MinPoolSize:      10,
		MaxPoolSize:      100,
		SocketTimeoutMs:  30000,
		ConnectTimeoutMs: 10000,
		IsZap:            true,
		Hosts: []*MongoHost{
			{
				Host: "localhost",
				Port: "27017",
			},
		},
	}
}

// Dsn returns MongoDB connection string
func (m *MongoDBConfig) Dsn() string {
	return m.Uri()
}

// Uri returns MongoDB URI
func (m *MongoDBConfig) Uri() string {
	length := len(m.Hosts)
	hosts := make([]string, 0, length)
	for i := 0; i < length; i++ {
		if m.Hosts[i].Host != "" && m.Hosts[i].Port != "" {
			hosts = append(hosts, m.Hosts[i].Host+":"+m.Hosts[i].Port)
		}
	}

	if len(hosts) == 0 {
		return ""
	}

	// Build URI with authentication if provided
	var uri string
	if m.Username != "" && m.Password != "" {
		uri = fmt.Sprintf("mongodb://%s:%s@%s/%s",
			m.Username,
			m.Password,
			strings.Join(hosts, ","),
			m.Database,
		)
	} else {
		uri = fmt.Sprintf("mongodb://%s/%s",
			strings.Join(hosts, ","),
			m.Database,
		)
	}

	// Add options if provided
	if m.Options != "" {
		uri += "?" + m.Options
	}

	return uri
}

// AddHost adds a host to the MongoDB configuration
func (m *MongoDBConfig) AddHost(host, port string) {
	if m.Hosts == nil {
		m.Hosts = make([]*MongoHost, 0)
	}
	m.Hosts = append(m.Hosts, &MongoHost{
		Host: host,
		Port: port,
	})
}

// SetCredentials sets username and password
func (m *MongoDBConfig) SetCredentials(username, password string) {
	m.Username = username
	m.Password = password
}

// SetDatabase sets the database name
func (m *MongoDBConfig) SetDatabase(database string) {
	m.Database = database
}

// SetAuthSource sets the authentication source
func (m *MongoDBConfig) SetAuthSource(authSource string) {
	m.AuthSource = authSource
}

// SetPoolSize sets the connection pool size
func (m *MongoDBConfig) SetPoolSize(min, max uint64) {
	m.MinPoolSize = min
	m.MaxPoolSize = max
}

// SetTimeouts sets the connection and socket timeouts
func (m *MongoDBConfig) SetTimeouts(connectTimeout, socketTimeout int64) {
	m.ConnectTimeoutMs = connectTimeout
	m.SocketTimeoutMs = socketTimeout
}

// SetOptions sets MongoDB options
func (m *MongoDBConfig) SetOptions(options string) {
	m.Options = options
}

// AddOption adds a single option to the options string
func (m *MongoDBConfig) AddOption(key, value string) {
	if m.Options == "" {
		m.Options = fmt.Sprintf("%s=%s", key, value)
	} else {
		m.Options += fmt.Sprintf("&%s=%s", key, value)
	}
}

// SetZapLogging enables or disables Zap logging
func (m *MongoDBConfig) SetZapLogging(enabled bool) {
	m.IsZap = enabled
}

// SetCollection sets the default collection name
func (m *MongoDBConfig) SetCollection(collection string) {
	m.Coll = collection
}

// GetHosts returns the list of hosts
func (m *MongoDBConfig) GetHosts() []*MongoHost {
	return m.Hosts
}

// GetPrimaryHost returns the first host (primary)
func (m *MongoDBConfig) GetPrimaryHost() *MongoHost {
	if len(m.Hosts) > 0 {
		return m.Hosts[0]
	}
	return nil
}

// IsReplicaSet returns true if multiple hosts are configured
func (m *MongoDBConfig) IsReplicaSet() bool {
	return len(m.Hosts) > 1
}

// GetConnectionString returns a formatted connection string for logging
func (m *MongoDBConfig) GetConnectionString() string {
	hosts := make([]string, 0, len(m.Hosts))
	for _, host := range m.Hosts {
		hosts = append(hosts, host.Host+":"+host.Port)
	}

	auth := ""
	if m.Username != "" {
		auth = "***:***@"
	}

	return fmt.Sprintf("mongodb://%s%s/%s", auth, strings.Join(hosts, ","), m.Database)
}
