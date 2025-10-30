package core

// OperationType 操作类型
type OperationType string

const (
	// OpTypeRead 读操作
	OpTypeRead OperationType = "read"
	// OpTypeWrite 写操作
	OpTypeWrite OperationType = "write"
	// OpTypeDelete 删除操作
	OpTypeDelete OperationType = "delete"
	// OpTypeCustom 自定义操作
	OpTypeCustom OperationType = "custom"
)

// Operation 操作接口
type Operation interface {
	// Type 返回操作类型
	Type() OperationType

	// Key 返回操作的键（如果适用）
	Key() string

	// Value 返回操作的值（如果适用）
	Value() []byte

	// Metadata 返回操作的元数据
	Metadata() map[string]interface{}
}
