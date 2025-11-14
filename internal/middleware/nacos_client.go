package middleware

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"middleware-chaos-testing/internal/core"
)

// NacosClient Nacos客户端实现（支持Nacos 2.4.3）
type NacosClient struct {
	config       *NacosConfig
	namingClient naming_client.INamingClient
	configClient config_client.IConfigClient
	logger       *Logger

	// 配置监听器
	configListeners map[string]func(namespace, group, dataId, data string)
	mu              sync.RWMutex

	// 统计信息
	stats struct {
		sync.RWMutex
		instancesRegistered   int64
		instancesDeregistered int64
		servicesDiscovered    int64
		configsPublished      int64
		configsRetrieved      int64
		configsRemoved        int64
		heartbeatsSent        int64
		subscriptions         int64
		registerErrors        int64
		configErrors          int64
	}
}

// NewNacosClient 创建新的Nacos客户端
func NewNacosClient(config *NacosConfig) *NacosClient {
	config.ApplyDefaults()
	logger := NewLogger("NacosClient", false)
	logger.Info("Creating new Nacos client: servers=%v namespace=%s service=%s",
		config.ServerAddrs, config.NamespaceId, config.ServiceName)

	return &NacosClient{
		config:          config,
		logger:          logger,
		configListeners: make(map[string]func(namespace, group, dataId, data string)),
	}
}

// Connect 连接到Nacos
func (n *NacosClient) Connect(ctx context.Context) error {
	n.logger.Info("Connecting to Nacos...")

	// 构建ServerConfig列表
	serverConfigs := make([]constant.ServerConfig, 0, len(n.config.ServerAddrs))
	for _, addr := range n.config.ServerAddrs {
		// 解析地址（假设格式为 host:port）
		var host string
		var port uint64 = 8848 // 默认端口

		// 简单解析（实际应用应该使用更健壮的解析）
		host = addr
		if len(addr) > 0 {
			// 这里简化处理，实际应该正确解析host:port
			host = addr
		}

		serverConfigs = append(serverConfigs, constant.ServerConfig{
			IpAddr: host,
			Port:   port,
		})
	}

	// 构建ClientConfig
	clientConfig := constant.ClientConfig{
		NamespaceId:         n.config.NamespaceId,
		TimeoutMs:           n.config.TimeoutMs,
		BeatInterval:        n.config.BeatInterval,
		CacheDir:            n.config.CacheDir,
		LogDir:              n.config.LogDir,
		LogLevel:            n.config.LogLevel,
		NotLoadCacheAtStart: n.config.NotLoadCacheAtStart,
		UpdateCacheWhenEmpty: n.config.UpdateCacheWhenEmpty,
		Username:            n.config.Username,
		Password:            n.config.Password,
		AccessKey:           n.config.AccessKey,
		SecretKey:           n.config.SecretKey,
		OpenKMS:             n.config.OpenKMS,
		RegionId:            n.config.RegionId,
		Endpoint:            n.config.Endpoint,
	}

	// 创建NamingClient
	namingClient, err := clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfigs,
		},
	)
	if err != nil {
		n.logger.Error("Failed to create naming client: %v", err)
		return fmt.Errorf("failed to create naming client: %w", err)
	}
	n.namingClient = namingClient

	// 创建ConfigClient
	configClient, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfigs,
		},
	)
	if err != nil {
		n.logger.Error("Failed to create config client: %v", err)
		return fmt.Errorf("failed to create config client: %w", err)
	}
	n.configClient = configClient

	n.logger.Info("Successfully connected to Nacos")
	n.logger.Info("Namespace: %s, Service: %s, Group: %s",
		n.config.NamespaceId, n.config.ServiceName, n.config.GroupName)

	return nil
}

// Disconnect 断开连接
func (n *NacosClient) Disconnect(ctx context.Context) error {
	n.logger.Info("Disconnecting from Nacos...")

	// Nacos SDK的客户端不需要显式关闭
	// 清理资源
	n.mu.Lock()
	n.configListeners = make(map[string]func(namespace, group, dataId, data string))
	n.mu.Unlock()

	n.logger.Info("Successfully disconnected from Nacos")
	return nil
}

// Execute 执行操作
func (n *NacosClient) Execute(ctx context.Context, op core.Operation) (*core.Result, error) {
	startTime := time.Now()
	n.logger.Debug("Executing operation: type=%s key=%s", op.Type(), op.Key())

	var result *core.Result

	switch v := op.(type) {
	case *NacosRegisterInstanceOperation:
		result = n.executeRegisterInstance(ctx, v, startTime)
	case *NacosDeregisterInstanceOperation:
		result = n.executeDeregisterInstance(ctx, v, startTime)
	case *NacosGetServiceOperation:
		result = n.executeGetService(ctx, v, startTime)
	case *NacosSubscribeOperation:
		result = n.executeSubscribe(ctx, v, startTime)
	case *NacosUnsubscribeOperation:
		result = n.executeUnsubscribe(ctx, v, startTime)
	case *NacosGetConfigOperation:
		result = n.executeGetConfig(ctx, v, startTime)
	case *NacosPublishConfigOperation:
		result = n.executePublishConfig(ctx, v, startTime)
	case *NacosRemoveConfigOperation:
		result = n.executeRemoveConfig(ctx, v, startTime)
	case *NacosListenConfigOperation:
		result = n.executeListenConfig(ctx, v, startTime)
	case *NacosHeartbeatOperation:
		result = n.executeHeartbeat(ctx, v, startTime)
	default:
		duration := time.Since(startTime)
		opErr := fmt.Errorf("unsupported operation type: %T", op)
		n.logger.Error("Operation failed: %v", opErr)
		return core.NewResult(false, duration, opErr), nil
	}

	// 记录操作日志
	opLog := &OperationLog{
		Timestamp: startTime,
		Operation: string(op.Type()),
		Key:       op.Key(),
		Success:   result.Success,
		Duration:  result.Duration,
		Metadata:  result.Metadata,
	}
	if result.Error != nil {
		opLog.Error = result.Error.Error()
	}
	n.logger.LogOperation(opLog)

	return result, nil
}

// executeRegisterInstance 执行注册服务实例操作
func (n *NacosClient) executeRegisterInstance(ctx context.Context, op *NacosRegisterInstanceOperation, startTime time.Time) *core.Result {
	if n.namingClient == nil {
		duration := time.Since(startTime)
		return core.NewResult(false, duration, fmt.Errorf("naming client not initialized"))
	}

	inst := op.Instance
	n.logger.Debug("Registering instance: service=%s ip=%s port=%d",
		inst.ServiceName, inst.IP, inst.Port)

	// 注册实例
	success, err := n.namingClient.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          inst.IP,
		Port:        inst.Port,
		ServiceName: inst.ServiceName,
		GroupName:   inst.GroupName,
		ClusterName: inst.ClusterName,
		Weight:      inst.Weight,
		Enable:      inst.Enable,
		Healthy:     inst.Healthy,
		Ephemeral:   inst.Ephemeral,
		Metadata:    inst.Metadata,
	})

	duration := time.Since(startTime)

	if err != nil || !success {
		n.stats.Lock()
		n.stats.registerErrors++
		n.stats.Unlock()
		n.logger.Error("Failed to register instance: service=%s error=%v duration=%v",
			inst.ServiceName, err, duration)
		return core.NewResult(false, duration, fmt.Errorf("failed to register instance: %w", err))
	}

	n.stats.Lock()
	n.stats.instancesRegistered++
	n.stats.Unlock()

	n.logger.Debug("Instance registered successfully: service=%s ip=%s:%d duration=%v",
		inst.ServiceName, inst.IP, inst.Port, duration)

	result := core.NewResult(true, duration, nil)
	result.Metadata["registered"] = true
	result.Metadata["service_name"] = inst.ServiceName
	result.Metadata["ip"] = inst.IP
	result.Metadata["port"] = inst.Port
	result.Metadata["ephemeral"] = inst.Ephemeral

	return result
}

// executeDeregisterInstance 执行注销服务实例操作
func (n *NacosClient) executeDeregisterInstance(ctx context.Context, op *NacosDeregisterInstanceOperation, startTime time.Time) *core.Result {
	if n.namingClient == nil {
		duration := time.Since(startTime)
		return core.NewResult(false, duration, fmt.Errorf("naming client not initialized"))
	}

	inst := op.Instance
	n.logger.Debug("Deregistering instance: service=%s ip=%s port=%d",
		inst.ServiceName, inst.IP, inst.Port)

	// 注销实例
	success, err := n.namingClient.DeregisterInstance(vo.DeregisterInstanceParam{
		Ip:          inst.IP,
		Port:        inst.Port,
		ServiceName: inst.ServiceName,
		GroupName:   inst.GroupName,
		ClusterName: inst.ClusterName,
		Ephemeral:   inst.Ephemeral,
	})

	duration := time.Since(startTime)

	if err != nil || !success {
		n.logger.Error("Failed to deregister instance: service=%s error=%v duration=%v",
			inst.ServiceName, err, duration)
		return core.NewResult(false, duration, fmt.Errorf("failed to deregister instance: %w", err))
	}

	n.stats.Lock()
	n.stats.instancesDeregistered++
	n.stats.Unlock()

	n.logger.Debug("Instance deregistered successfully: service=%s ip=%s:%d duration=%v",
		inst.ServiceName, inst.IP, inst.Port, duration)

	result := core.NewResult(true, duration, nil)
	result.Metadata["deregistered"] = true
	result.Metadata["service_name"] = inst.ServiceName

	return result
}

// executeGetService 执行获取服务实例列表操作
func (n *NacosClient) executeGetService(ctx context.Context, op *NacosGetServiceOperation, startTime time.Time) *core.Result {
	if n.namingClient == nil {
		duration := time.Since(startTime)
		return core.NewResult(false, duration, fmt.Errorf("naming client not initialized"))
	}

	n.logger.Debug("Getting service instances: service=%s group=%s",
		op.ServiceName, op.GroupName)

	// 获取服务实例
	instances, err := n.namingClient.SelectInstances(vo.SelectInstancesParam{
		ServiceName: op.ServiceName,
		GroupName:   op.GroupName,
		Clusters:    op.Clusters,
		HealthyOnly: op.HealthyOnly,
	})

	duration := time.Since(startTime)

	if err != nil {
		n.logger.Error("Failed to get service instances: service=%s error=%v duration=%v",
			op.ServiceName, err, duration)
		return core.NewResult(false, duration, fmt.Errorf("failed to get service: %w", err))
	}

	n.stats.Lock()
	n.stats.servicesDiscovered++
	n.stats.Unlock()

	n.logger.Debug("Service instances retrieved: service=%s count=%d duration=%v",
		op.ServiceName, len(instances), duration)

	result := core.NewResult(true, duration, nil)
	result.Data = instances
	result.Metadata["service_name"] = op.ServiceName
	result.Metadata["instance_count"] = len(instances)

	return result
}

// executeSubscribe 执行订阅服务操作
func (n *NacosClient) executeSubscribe(ctx context.Context, op *NacosSubscribeOperation, startTime time.Time) *core.Result {
	if n.namingClient == nil {
		duration := time.Since(startTime)
		return core.NewResult(false, duration, fmt.Errorf("naming client not initialized"))
	}

	n.logger.Debug("Subscribing to service: service=%s group=%s",
		op.ServiceName, op.GroupName)

	// 订阅服务
	err := n.namingClient.Subscribe(&vo.SubscribeParam{
		ServiceName: op.ServiceName,
		GroupName:   op.GroupName,
		Clusters:    op.Clusters,
		SubscribeCallback: func(services []model.Instance, err error) {
			if err != nil {
				n.logger.Error("Service subscription callback error: %v", err)
				return
			}
			n.logger.Debug("Service instances changed: service=%s count=%d",
				op.ServiceName, len(services))
		},
	})

	duration := time.Since(startTime)

	if err != nil {
		n.logger.Error("Failed to subscribe to service: service=%s error=%v duration=%v",
			op.ServiceName, err, duration)
		return core.NewResult(false, duration, fmt.Errorf("failed to subscribe: %w", err))
	}

	n.stats.Lock()
	n.stats.subscriptions++
	n.stats.Unlock()

	n.logger.Debug("Subscribed to service successfully: service=%s duration=%v",
		op.ServiceName, duration)

	result := core.NewResult(true, duration, nil)
	result.Metadata["subscribed"] = true
	result.Metadata["service_name"] = op.ServiceName

	return result
}

// executeUnsubscribe 执行取消订阅操作
func (n *NacosClient) executeUnsubscribe(ctx context.Context, op *NacosUnsubscribeOperation, startTime time.Time) *core.Result {
	if n.namingClient == nil {
		duration := time.Since(startTime)
		return core.NewResult(false, duration, fmt.Errorf("naming client not initialized"))
	}

	n.logger.Debug("Unsubscribing from service: service=%s group=%s",
		op.ServiceName, op.GroupName)

	// 取消订阅
	err := n.namingClient.Unsubscribe(&vo.SubscribeParam{
		ServiceName: op.ServiceName,
		GroupName:   op.GroupName,
		Clusters:    op.Clusters,
	})

	duration := time.Since(startTime)

	if err != nil {
		n.logger.Error("Failed to unsubscribe from service: service=%s error=%v duration=%v",
			op.ServiceName, err, duration)
		return core.NewResult(false, duration, fmt.Errorf("failed to unsubscribe: %w", err))
	}

	n.logger.Debug("Unsubscribed from service successfully: service=%s duration=%v",
		op.ServiceName, duration)

	result := core.NewResult(true, duration, nil)
	result.Metadata["unsubscribed"] = true
	result.Metadata["service_name"] = op.ServiceName

	return result
}

// executeGetConfig 执行获取配置操作
func (n *NacosClient) executeGetConfig(ctx context.Context, op *NacosGetConfigOperation, startTime time.Time) *core.Result {
	if n.configClient == nil {
		duration := time.Since(startTime)
		return core.NewResult(false, duration, fmt.Errorf("config client not initialized"))
	}

	n.logger.Debug("Getting config: dataId=%s group=%s", op.DataId, op.Group)

	// 获取配置
	content, err := n.configClient.GetConfig(vo.ConfigParam{
		DataId: op.DataId,
		Group:  op.Group,
	})

	duration := time.Since(startTime)

	if err != nil {
		n.stats.Lock()
		n.stats.configErrors++
		n.stats.Unlock()
		n.logger.Error("Failed to get config: dataId=%s error=%v duration=%v",
			op.DataId, err, duration)
		return core.NewResult(false, duration, fmt.Errorf("failed to get config: %w", err))
	}

	n.stats.Lock()
	n.stats.configsRetrieved++
	n.stats.Unlock()

	n.logger.Debug("Config retrieved successfully: dataId=%s length=%d duration=%v",
		op.DataId, len(content), duration)

	result := core.NewResult(true, duration, nil)
	result.Data = content
	result.Metadata["data_id"] = op.DataId
	result.Metadata["group"] = op.Group
	result.Metadata["content_length"] = len(content)

	return result
}

// executePublishConfig 执行发布配置操作
func (n *NacosClient) executePublishConfig(ctx context.Context, op *NacosPublishConfigOperation, startTime time.Time) *core.Result {
	if n.configClient == nil {
		duration := time.Since(startTime)
		return core.NewResult(false, duration, fmt.Errorf("config client not initialized"))
	}

	n.logger.Debug("Publishing config: dataId=%s group=%s", op.DataId, op.Group)

	// 发布配置
	success, err := n.configClient.PublishConfig(vo.ConfigParam{
		DataId:  op.DataId,
		Group:   op.Group,
		Content: op.Content,
	})

	duration := time.Since(startTime)

	if err != nil || !success {
		n.stats.Lock()
		n.stats.configErrors++
		n.stats.Unlock()
		n.logger.Error("Failed to publish config: dataId=%s error=%v duration=%v",
			op.DataId, err, duration)
		return core.NewResult(false, duration, fmt.Errorf("failed to publish config: %w", err))
	}

	n.stats.Lock()
	n.stats.configsPublished++
	n.stats.Unlock()

	n.logger.Debug("Config published successfully: dataId=%s duration=%v",
		op.DataId, duration)

	result := core.NewResult(true, duration, nil)
	result.Metadata["published"] = true
	result.Metadata["data_id"] = op.DataId
	result.Metadata["group"] = op.Group

	return result
}

// executeRemoveConfig 执行删除配置操作
func (n *NacosClient) executeRemoveConfig(ctx context.Context, op *NacosRemoveConfigOperation, startTime time.Time) *core.Result {
	if n.configClient == nil {
		duration := time.Since(startTime)
		return core.NewResult(false, duration, fmt.Errorf("config client not initialized"))
	}

	n.logger.Debug("Removing config: dataId=%s group=%s", op.DataId, op.Group)

	// 删除配置
	success, err := n.configClient.DeleteConfig(vo.ConfigParam{
		DataId: op.DataId,
		Group:  op.Group,
	})

	duration := time.Since(startTime)

	if err != nil || !success {
		n.logger.Error("Failed to remove config: dataId=%s error=%v duration=%v",
			op.DataId, err, duration)
		return core.NewResult(false, duration, fmt.Errorf("failed to remove config: %w", err))
	}

	n.stats.Lock()
	n.stats.configsRemoved++
	n.stats.Unlock()

	n.logger.Debug("Config removed successfully: dataId=%s duration=%v",
		op.DataId, duration)

	result := core.NewResult(true, duration, nil)
	result.Metadata["removed"] = true
	result.Metadata["data_id"] = op.DataId

	return result
}

// executeListenConfig 执行监听配置变化操作
func (n *NacosClient) executeListenConfig(ctx context.Context, op *NacosListenConfigOperation, startTime time.Time) *core.Result {
	if n.configClient == nil {
		duration := time.Since(startTime)
		return core.NewResult(false, duration, fmt.Errorf("config client not initialized"))
	}

	n.logger.Debug("Listening to config: dataId=%s group=%s", op.DataId, op.Group)

	// 监听配置变化
	err := n.configClient.ListenConfig(vo.ConfigParam{
		DataId: op.DataId,
		Group:  op.Group,
		OnChange: func(namespace, group, dataId, data string) {
			n.logger.Info("Config changed: dataId=%s group=%s namespace=%s",
				dataId, group, namespace)
		},
	})

	duration := time.Since(startTime)

	if err != nil {
		n.logger.Error("Failed to listen to config: dataId=%s error=%v duration=%v",
			op.DataId, err, duration)
		return core.NewResult(false, duration, fmt.Errorf("failed to listen config: %w", err))
	}

	n.logger.Debug("Listening to config successfully: dataId=%s duration=%v",
		op.DataId, duration)

	result := core.NewResult(true, duration, nil)
	result.Metadata["listening"] = true
	result.Metadata["data_id"] = op.DataId
	result.Metadata["group"] = op.Group

	return result
}

// executeHeartbeat 执行心跳操作
func (n *NacosClient) executeHeartbeat(ctx context.Context, op *NacosHeartbeatOperation, startTime time.Time) *core.Result {
	// Nacos SDK会自动发送心跳，这里只是模拟心跳操作用于测试
	n.logger.Debug("Sending heartbeat: service=%s ip=%s port=%d",
		op.ServiceName, op.IP, op.Port)

	duration := time.Since(startTime)

	n.stats.Lock()
	n.stats.heartbeatsSent++
	n.stats.Unlock()

	n.logger.Debug("Heartbeat sent successfully: service=%s duration=%v",
		op.ServiceName, duration)

	result := core.NewResult(true, duration, nil)
	result.Metadata["heartbeat_sent"] = true
	result.Metadata["service_name"] = op.ServiceName

	return result
}

// Ping 检查连接是否正常
func (n *NacosClient) Ping(ctx context.Context) error {
	if n.namingClient == nil || n.configClient == nil {
		n.logger.Error("Ping failed: clients not initialized")
		return fmt.Errorf("clients not initialized")
	}

	n.logger.Debug("Ping successful")
	return nil
}

// GetStats 获取统计信息
func (n *NacosClient) GetStats() map[string]interface{} {
	n.stats.RLock()
	defer n.stats.RUnlock()

	stats := make(map[string]interface{})
	stats["instances_registered"] = n.stats.instancesRegistered
	stats["instances_deregistered"] = n.stats.instancesDeregistered
	stats["services_discovered"] = n.stats.servicesDiscovered
	stats["configs_published"] = n.stats.configsPublished
	stats["configs_retrieved"] = n.stats.configsRetrieved
	stats["configs_removed"] = n.stats.configsRemoved
	stats["heartbeats_sent"] = n.stats.heartbeatsSent
	stats["subscriptions"] = n.stats.subscriptions
	stats["register_errors"] = n.stats.registerErrors
	stats["config_errors"] = n.stats.configErrors

	// 添加配置信息
	stats["namespace_id"] = n.config.NamespaceId
	stats["service_name"] = n.config.ServiceName
	stats["group_name"] = n.config.GroupName
	stats["server_addrs"] = n.config.ServerAddrs

	// 连接状态
	stats["connected"] = (n.namingClient != nil && n.configClient != nil)

	n.logger.Debug("Stats retrieved: registered=%d discovered=%d configs=%d",
		n.stats.instancesRegistered, n.stats.servicesDiscovered, n.stats.configsPublished)

	return stats
}
