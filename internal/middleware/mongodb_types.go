package middleware

import (
	"time"

	"middleware-chaos-testing/internal/core"
)

// MongoDBConfig MongoDB配置
type MongoDBConfig struct {
	URI        string        // mongodb://user:pass@host:port/db
	Database   string        // 数据库名
	Collection string        // 集合名
	Timeout    time.Duration // 超时时间

	// 连接池配置
	MaxPoolSize int           // 最大连接池大小，默认: 100
	MinPoolSize int           // 最小连接池大小，默认: 10
	MaxIdleTime time.Duration // 最大空闲时间，默认: 10分钟

	// 副本集配置
	ReplicaSet     string // 副本集名称
	ReadPreference string // 读偏好: primary, secondary, nearest
	WriteConcern   string // 写关注: majority, w1, w2

	// 性能优化
	Compressors []string // 压缩算法: snappy, zlib, zstd
}

// ApplyDefaults 应用默认配置
func (c *MongoDBConfig) ApplyDefaults() {
	if c.Timeout == 0 {
		c.Timeout = 5 * time.Second
	}
	if c.MaxPoolSize == 0 {
		c.MaxPoolSize = 100
	}
	if c.MinPoolSize == 0 {
		c.MinPoolSize = 10
	}
	if c.MaxIdleTime == 0 {
		c.MaxIdleTime = 10 * time.Minute
	}
	if c.ReadPreference == "" {
		c.ReadPreference = "primary"
	}
	if c.WriteConcern == "" {
		c.WriteConcern = "majority"
	}
}

// MongoDB操作类型

// MongoDBInsertOperation 插入操作
type MongoDBInsertOperation struct {
	core.BaseOperation
	Document map[string]interface{} // 要插入的文档
}

// MongoDBFindOperation 查询操作
type MongoDBFindOperation struct {
	core.BaseOperation
	Filter map[string]interface{} // 查询过滤器
	Limit  int64                  // 限制返回数量
}

// MongoDBUpdateOperation 更新操作
type MongoDBUpdateOperation struct {
	core.BaseOperation
	Filter map[string]interface{} // 查询过滤器
	Update map[string]interface{} // 更新内容
}

// MongoDBDeleteOperation 删除操作
type MongoDBDeleteOperation struct {
	core.BaseOperation
	Filter map[string]interface{} // 删除过滤器
}

// MongoDBAggregateOperation 聚合操作
type MongoDBAggregateOperation struct {
	core.BaseOperation
	Pipeline []map[string]interface{} // 聚合管道
}
