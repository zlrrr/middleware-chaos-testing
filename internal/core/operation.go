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

// BaseOperation 提供Operation接口的默认实现
// 可以被嵌入到具体的操作类型中，避免重复代码
type BaseOperation struct {
	OpType     OperationType
	OpKey      string
	OpValue    []byte
	OpMetadata map[string]interface{}
}

// Type 返回操作类型
func (b *BaseOperation) Type() OperationType {
	if b.OpType == "" {
		return OpTypeCustom
	}
	return b.OpType
}

// Key 返回操作的键
func (b *BaseOperation) Key() string {
	return b.OpKey
}

// Value 返回操作的值
func (b *BaseOperation) Value() []byte {
	return b.OpValue
}

// Metadata 返回操作的元数据
func (b *BaseOperation) Metadata() map[string]interface{} {
	if b.OpMetadata == nil {
		return make(map[string]interface{})
	}
	return b.OpMetadata
}

// NewBaseOperation 创建一个新的BaseOperation
func NewBaseOperation(opType OperationType, key string, value []byte) *BaseOperation {
	return &BaseOperation{
		OpType:     opType,
		OpKey:      key,
		OpValue:    value,
		OpMetadata: make(map[string]interface{}),
	}
}
