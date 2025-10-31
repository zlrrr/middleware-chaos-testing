package middleware

import (
	"time"

	"middleware-chaos-testing/internal/core"
)

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string        // 主机地址
	Port     int           // 端口
	Password string        // 密码
	DB       int           // 数据库编号
	Timeout  time.Duration // 超时时间
}

// RedisClient 的完整实现在 redis_client.go 中

// RedisSetOperation SET操作
type RedisSetOperation struct {
	OpKey   string
	OpValue []byte
}

func (r *RedisSetOperation) Type() core.OperationType {
	return core.OpTypeWrite
}

func (r *RedisSetOperation) Key() string {
	return r.OpKey
}

func (r *RedisSetOperation) Value() []byte {
	return r.OpValue
}

func (r *RedisSetOperation) Metadata() map[string]interface{} {
	return make(map[string]interface{})
}

// RedisGetOperation GET操作
type RedisGetOperation struct {
	OpKey string
}

func (r *RedisGetOperation) Type() core.OperationType {
	return core.OpTypeRead
}

func (r *RedisGetOperation) Key() string {
	return r.OpKey
}

func (r *RedisGetOperation) Value() []byte {
	return nil
}

func (r *RedisGetOperation) Metadata() map[string]interface{} {
	return make(map[string]interface{})
}

// RedisDeleteOperation DELETE操作
type RedisDeleteOperation struct {
	OpKey string
}

func (r *RedisDeleteOperation) Type() core.OperationType {
	return core.OpTypeDelete
}

func (r *RedisDeleteOperation) Key() string {
	return r.OpKey
}

func (r *RedisDeleteOperation) Value() []byte {
	return nil
}

func (r *RedisDeleteOperation) Metadata() map[string]interface{} {
	return make(map[string]interface{})
}
