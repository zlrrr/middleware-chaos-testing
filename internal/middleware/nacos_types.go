package middleware

import (
	"time"

	"middleware-chaos-testing/internal/core"
)

// NacosConfig Nacos配置（Nacos 2.4.3）
type NacosConfig struct {
	ServerAddrs []string // Nacos服务器地址列表，如: ["127.0.0.1:8848"]
	NamespaceId string   // 命名空间ID，默认: "public"

	// 服务注册配置
	ServiceName string  // 服务名
	GroupName   string  // 分组名，默认: DEFAULT_GROUP
	ClusterName string  // 集群名，默认: DEFAULT
	IP          string  // 服务IP
	Port        uint64  // 服务端口
	Weight      float64 // 权重，默认: 1.0
	Enable      bool    // 是否启用，默认: true
	Healthy     bool    // 健康状态，默认: true
	Ephemeral   bool    // 是否临时实例，默认: true（临时实例需要心跳）
	Metadata    map[string]string // 元数据

	// 配置中心配置
	DataId      string // 配置ID
	ConfigGroup string // 配置分组，默认: DEFAULT_GROUP

	// 客户端配置
	TimeoutMs    uint64 // 超时时间（毫秒），默认: 10000
	BeatInterval int64  // 心跳间隔（毫秒），默认: 5000
	CacheDir     string // 缓存目录，默认: /tmp/nacos/cache
	LogDir       string // 日志目录，默认: /tmp/nacos/log
	LogLevel     string // 日志级别: debug/info/warn/error，默认: info

	// Nacos 2.x gRPC配置
	NotLoadCacheAtStart bool          // 启动时不加载缓存，默认: false
	UpdateCacheWhenEmpty bool          // 缓存为空时更新，默认: false
	Username            string        // 用户名（Nacos 2.x鉴权）
	Password            string        // 密码（Nacos 2.x鉴权）
	AccessKey           string        // AccessKey
	SecretKey           string        // SecretKey
	OpenKMS             bool          // 是否开启KMS，默认: false
	RegionId            string        // RegionId（阿里云KMS）
	Endpoint            string        // Endpoint
}

// ApplyDefaults 应用默认配置
func (c *NacosConfig) ApplyDefaults() {
	if c.NamespaceId == "" {
		c.NamespaceId = "public"
	}
	if c.GroupName == "" {
		c.GroupName = "DEFAULT_GROUP"
	}
	if c.ClusterName == "" {
		c.ClusterName = "DEFAULT"
	}
	if c.ConfigGroup == "" {
		c.ConfigGroup = "DEFAULT_GROUP"
	}
	if c.Weight == 0 {
		c.Weight = 1.0
	}
	if c.TimeoutMs == 0 {
		c.TimeoutMs = 10000 // 10秒
	}
	if c.BeatInterval == 0 {
		c.BeatInterval = 5000 // 5秒
	}
	if c.CacheDir == "" {
		c.CacheDir = "/tmp/nacos/cache"
	}
	if c.LogDir == "" {
		c.LogDir = "/tmp/nacos/log"
	}
	if c.LogLevel == "" {
		c.LogLevel = "info"
	}
}

// Nacos服务实例结构

// NacosInstance Nacos服务实例
type NacosInstance struct {
	ServiceName string            // 服务名
	GroupName   string            // 分组名
	ClusterName string            // 集群名
	IP          string            // IP地址
	Port        uint64            // 端口
	Weight      float64           // 权重
	Enable      bool              // 是否启用
	Healthy     bool              // 健康状态
	Ephemeral   bool              // 是否临时实例
	Metadata    map[string]string // 元数据
}

// Nacos操作类型

// NacosRegisterInstanceOperation 注册服务实例操作
type NacosRegisterInstanceOperation struct {
	core.BaseOperation
	Instance *NacosInstance // 要注册的服务实例
}

// NacosDeregisterInstanceOperation 注销服务实例操作
type NacosDeregisterInstanceOperation struct {
	core.BaseOperation
	Instance *NacosInstance // 要注销的服务实例
}

// NacosGetServiceOperation 获取服务实例列表操作
type NacosGetServiceOperation struct {
	core.BaseOperation
	ServiceName string   // 服务名
	GroupName   string   // 分组名
	Clusters    []string // 集群列表
	HealthyOnly bool     // 是否只获取健康实例
}

// NacosSubscribeOperation 订阅服务操作
type NacosSubscribeOperation struct {
	core.BaseOperation
	ServiceName string   // 服务名
	GroupName   string   // 分组名
	Clusters    []string // 集群列表
}

// NacosUnsubscribeOperation 取消订阅服务操作
type NacosUnsubscribeOperation struct {
	core.BaseOperation
	ServiceName string   // 服务名
	GroupName   string   // 分组名
	Clusters    []string // 集群列表
}

// NacosGetConfigOperation 获取配置操作
type NacosGetConfigOperation struct {
	core.BaseOperation
	DataId string // 配置ID
	Group  string // 配置分组
}

// NacosPublishConfigOperation 发布配置操作
type NacosPublishConfigOperation struct {
	core.BaseOperation
	DataId  string // 配置ID
	Group   string // 配置分组
	Content string // 配置内容
}

// NacosRemoveConfigOperation 删除配置操作
type NacosRemoveConfigOperation struct {
	core.BaseOperation
	DataId string // 配置ID
	Group  string // 配置分组
}

// NacosListenConfigOperation 监听配置变化操作
type NacosListenConfigOperation struct {
	core.BaseOperation
	DataId string // 配置ID
	Group  string // 配置分组
}

// NacosHeartbeatOperation 发送心跳操作
type NacosHeartbeatOperation struct {
	core.BaseOperation
	ServiceName string // 服务名
	GroupName   string // 分组名
	IP          string // IP地址
	Port        uint64 // 端口
}

// Nacos常量定义
const (
	// 默认分组
	DefaultGroup = "DEFAULT_GROUP"

	// 默认集群
	DefaultCluster = "DEFAULT"

	// 公共命名空间
	PublicNamespace = "public"

	// 心跳间隔
	DefaultBeatInterval = 5 * time.Second

	// 超时时间
	DefaultTimeout = 10 * time.Second
)

// Nacos服务实例健康状态
const (
	InstanceHealthy   = true  // 健康
	InstanceUnhealthy = false // 不健康
)

// Nacos实例类型
const (
	InstanceEphemeral   = true  // 临时实例（需要心跳保活）
	InstancePersistent  = false // 持久实例（不需要心跳）
)
