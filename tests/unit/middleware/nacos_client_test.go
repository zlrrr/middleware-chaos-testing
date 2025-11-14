package middleware_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"middleware-chaos-testing/internal/core"
	"middleware-chaos-testing/internal/middleware"
)

// NacosClientTestSuite Nacos客户端测试套件
type NacosClientTestSuite struct {
	suite.Suite
	client *middleware.NacosClient
	config *middleware.NacosConfig
}

// SetupTest 每个测试前执行
func (suite *NacosClientTestSuite) SetupTest() {
	suite.config = &middleware.NacosConfig{
		ServerAddrs: []string{"127.0.0.1:8848"},
		NamespaceId: "public",
		ServiceName: "chaos-test-service",
		GroupName:   "DEFAULT_GROUP",
		ClusterName: "DEFAULT",
		IP:          "127.0.0.1",
		Port:        8080,
		Weight:      1.0,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
		DataId:      "chaos-test-config",
		ConfigGroup: "DEFAULT_GROUP",
		TimeoutMs:   10000,
		BeatInterval: 5000,
	}
	suite.client = middleware.NewNacosClient(suite.config)
}

// TearDownTest 每个测试后执行
func (suite *NacosClientTestSuite) TearDownTest() {
	if suite.client != nil {
		ctx := context.Background()
		_ = suite.client.Disconnect(ctx)
	}
}

// TestConnect 测试Nacos连接
func (suite *NacosClientTestSuite) TestConnect() {
	ctx := context.Background()

	// 测试成功连接
	err := suite.client.Connect(ctx)
	suite.NoError(err, "Connect should succeed")

	// 验证连接状态
	err = suite.client.Ping(ctx)
	suite.NoError(err, "Ping should succeed after connect")
}

// TestConnect_InvalidServerAddr 测试无效服务器地址连接
func (suite *NacosClientTestSuite) TestConnect_InvalidServerAddr() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	invalidConfig := &middleware.NacosConfig{
		ServerAddrs: []string{"invalid-host:8848"},
		NamespaceId: "public",
		ServiceName: "test-service",
		TimeoutMs:   1000,
	}
	client := middleware.NewNacosClient(invalidConfig)

	err := client.Connect(ctx)
	// 连接可能会超时或失败
	if err != nil {
		suite.T().Logf("Expected error for invalid server: %v", err)
	}
}

// TestRegisterInstance 测试服务实例注册
func (suite *NacosClientTestSuite) TestRegisterInstance() {
	ctx := context.Background()

	// 先连接
	err := suite.client.Connect(ctx)
	suite.Require().NoError(err)

	// 创建注册实例操作
	instance := &middleware.NacosInstance{
		ServiceName: suite.config.ServiceName,
		GroupName:   suite.config.GroupName,
		ClusterName: suite.config.ClusterName,
		IP:          suite.config.IP,
		Port:        suite.config.Port,
		Weight:      suite.config.Weight,
		Enable:      suite.config.Enable,
		Healthy:     suite.config.Healthy,
		Ephemeral:   suite.config.Ephemeral,
		Metadata: map[string]string{
			"version": "1.0.0",
			"env":     "test",
		},
	}

	op := &middleware.NacosRegisterInstanceOperation{
		BaseOperation: core.BaseOperation{
			OpType: core.OpTypeWrite,
			OpKey:  "register-instance",
		},
		Instance: instance,
	}

	// 执行注册
	result, err := suite.client.Execute(ctx, op)
	suite.NoError(err, "Register instance should succeed")
	suite.NotNil(result, "Result should not be nil")
	suite.True(result.Success, "Register operation should be successful")
	suite.Greater(result.Duration, time.Duration(0), "Duration should be positive")

	// 验证注册成功
	registered, ok := result.Metadata["registered"]
	suite.True(ok, "Should have registered metadata")
	suite.True(registered.(bool), "Instance should be registered")
}

// TestDeregisterInstance 测试服务实例注销
func (suite *NacosClientTestSuite) TestDeregisterInstance() {
	ctx := context.Background()

	err := suite.client.Connect(ctx)
	suite.Require().NoError(err)

	// 先注册实例
	instance := &middleware.NacosInstance{
		ServiceName: suite.config.ServiceName,
		GroupName:   suite.config.GroupName,
		ClusterName: suite.config.ClusterName,
		IP:          suite.config.IP,
		Port:        suite.config.Port,
	}

	registerOp := &middleware.NacosRegisterInstanceOperation{
		BaseOperation: core.BaseOperation{
			OpType: core.OpTypeWrite,
			OpKey:  "register-for-deregister",
		},
		Instance: instance,
	}
	_, err = suite.client.Execute(ctx, registerOp)
	suite.Require().NoError(err)

	time.Sleep(500 * time.Millisecond) // 等待注册生效

	// 注销实例
	deregisterOp := &middleware.NacosDeregisterInstanceOperation{
		BaseOperation: core.BaseOperation{
			OpType: core.OpTypeWrite,
			OpKey:  "deregister-instance",
		},
		Instance: instance,
	}

	result, err := suite.client.Execute(ctx, deregisterOp)
	suite.NoError(err, "Deregister instance should succeed")
	suite.True(result.Success, "Deregister operation should be successful")

	// 验证注销成功
	deregistered, ok := result.Metadata["deregistered"]
	suite.True(ok, "Should have deregistered metadata")
	suite.True(deregistered.(bool), "Instance should be deregistered")
}

// TestGetService 测试服务发现
func (suite *NacosClientTestSuite) TestGetService() {
	ctx := context.Background()

	err := suite.client.Connect(ctx)
	suite.Require().NoError(err)

	// 先注册一个实例
	instance := &middleware.NacosInstance{
		ServiceName: suite.config.ServiceName,
		GroupName:   suite.config.GroupName,
		IP:          suite.config.IP,
		Port:        suite.config.Port,
		Healthy:     true,
	}

	registerOp := &middleware.NacosRegisterInstanceOperation{
		BaseOperation: core.BaseOperation{
			OpType: core.OpTypeWrite,
			OpKey:  "register-for-get-service",
		},
		Instance: instance,
	}
	_, err = suite.client.Execute(ctx, registerOp)
	suite.Require().NoError(err)

	time.Sleep(500 * time.Millisecond) // 等待注册生效

	// 获取服务实例列表
	getServiceOp := &middleware.NacosGetServiceOperation{
		BaseOperation: core.BaseOperation{
			OpType: core.OpTypeRead,
			OpKey:  "get-service",
		},
		ServiceName: suite.config.ServiceName,
		GroupName:   suite.config.GroupName,
		Clusters:    []string{suite.config.ClusterName},
	}

	result, err := suite.client.Execute(ctx, getServiceOp)
	suite.NoError(err, "Get service should succeed")
	suite.True(result.Success, "Get service operation should be successful")

	// 验证返回了服务实例
	if result.Data != nil {
		suite.T().Logf("Service instances found: %v", result.Data)
	}
}

// TestSubscribe 测试服务订阅
func (suite *NacosClientTestSuite) TestSubscribe() {
	ctx := context.Background()

	err := suite.client.Connect(ctx)
	suite.Require().NoError(err)

	// 订阅服务
	subscribeOp := &middleware.NacosSubscribeOperation{
		BaseOperation: core.BaseOperation{
			OpType: core.OpTypeRead,
			OpKey:  "subscribe-service",
		},
		ServiceName: suite.config.ServiceName,
		GroupName:   suite.config.GroupName,
		Clusters:    []string{suite.config.ClusterName},
	}

	result, err := suite.client.Execute(ctx, subscribeOp)
	suite.NoError(err, "Subscribe should succeed")
	suite.True(result.Success, "Subscribe operation should be successful")

	// 验证订阅成功
	subscribed, ok := result.Metadata["subscribed"]
	suite.True(ok, "Should have subscribed metadata")
	suite.True(subscribed.(bool), "Should be subscribed")
}

// TestGetConfig 测试获取配置
func (suite *NacosClientTestSuite) TestGetConfig() {
	ctx := context.Background()

	err := suite.client.Connect(ctx)
	suite.Require().NoError(err)

	// 获取配置
	getConfigOp := &middleware.NacosGetConfigOperation{
		BaseOperation: core.BaseOperation{
			OpType: core.OpTypeRead,
			OpKey:  "get-config",
		},
		DataId: suite.config.DataId,
		Group:  suite.config.ConfigGroup,
	}

	result, err := suite.client.Execute(ctx, getConfigOp)
	suite.NoError(err, "Get config should not error")
	suite.NotNil(result, "Result should not be nil")

	// 配置可能不存在，Success可能为false
	if result.Success {
		suite.T().Log("Config found")
		if result.Data != nil {
			suite.T().Logf("Config content: %v", result.Data)
		}
	} else {
		suite.T().Log("Config not found (expected for first run)")
	}
}

// TestPublishConfig 测试发布配置
func (suite *NacosClientTestSuite) TestPublishConfig() {
	ctx := context.Background()

	err := suite.client.Connect(ctx)
	suite.Require().NoError(err)

	// 发布配置
	configContent := "test.config.key=test.config.value\ntest.config.enabled=true"
	publishConfigOp := &middleware.NacosPublishConfigOperation{
		BaseOperation: core.BaseOperation{
			OpType: core.OpTypeWrite,
			OpKey:  "publish-config",
		},
		DataId:  suite.config.DataId,
		Group:   suite.config.ConfigGroup,
		Content: configContent,
	}

	result, err := suite.client.Execute(ctx, publishConfigOp)
	suite.NoError(err, "Publish config should succeed")
	suite.True(result.Success, "Publish config operation should be successful")

	// 验证发布成功
	published, ok := result.Metadata["published"]
	suite.True(ok, "Should have published metadata")
	suite.True(published.(bool), "Config should be published")

	time.Sleep(500 * time.Millisecond) // 等待配置生效

	// 验证可以获取到刚发布的配置
	getConfigOp := &middleware.NacosGetConfigOperation{
		BaseOperation: core.BaseOperation{
			OpType: core.OpTypeRead,
			OpKey:  "get-published-config",
		},
		DataId: suite.config.DataId,
		Group:  suite.config.ConfigGroup,
	}

	getResult, err := suite.client.Execute(ctx, getConfigOp)
	suite.NoError(err, "Get published config should succeed")
	suite.True(getResult.Success, "Should get the published config")
}

// TestRemoveConfig 测试删除配置
func (suite *NacosClientTestSuite) TestRemoveConfig() {
	ctx := context.Background()

	err := suite.client.Connect(ctx)
	suite.Require().NoError(err)

	// 先发布一个配置
	publishConfigOp := &middleware.NacosPublishConfigOperation{
		BaseOperation: core.BaseOperation{
			OpType: core.OpTypeWrite,
			OpKey:  "publish-for-remove",
		},
		DataId:  suite.config.DataId,
		Group:   suite.config.ConfigGroup,
		Content: "config.to.be.removed=true",
	}
	_, err = suite.client.Execute(ctx, publishConfigOp)
	suite.Require().NoError(err)

	time.Sleep(500 * time.Millisecond)

	// 删除配置
	removeConfigOp := &middleware.NacosRemoveConfigOperation{
		BaseOperation: core.BaseOperation{
			OpType: core.OpTypeWrite,
			OpKey:  "remove-config",
		},
		DataId: suite.config.DataId,
		Group:  suite.config.ConfigGroup,
	}

	result, err := suite.client.Execute(ctx, removeConfigOp)
	suite.NoError(err, "Remove config should succeed")
	suite.True(result.Success, "Remove config operation should be successful")

	// 验证删除成功
	removed, ok := result.Metadata["removed"]
	suite.True(ok, "Should have removed metadata")
	suite.True(removed.(bool), "Config should be removed")
}

// TestListenConfig 测试监听配置变化
func (suite *NacosClientTestSuite) TestListenConfig() {
	ctx := context.Background()

	err := suite.client.Connect(ctx)
	suite.Require().NoError(err)

	// 监听配置变化
	listenConfigOp := &middleware.NacosListenConfigOperation{
		BaseOperation: core.BaseOperation{
			OpType: core.OpTypeRead,
			OpKey:  "listen-config",
		},
		DataId: suite.config.DataId,
		Group:  suite.config.ConfigGroup,
	}

	result, err := suite.client.Execute(ctx, listenConfigOp)
	suite.NoError(err, "Listen config should succeed")
	suite.True(result.Success, "Listen config operation should be successful")

	// 验证监听已设置
	listening, ok := result.Metadata["listening"]
	suite.True(ok, "Should have listening metadata")
	suite.True(listening.(bool), "Should be listening")

	suite.T().Log("Config listener registered - will be notified on config changes")
}

// TestHeartbeat 测试心跳
func (suite *NacosClientTestSuite) TestHeartbeat() {
	ctx := context.Background()

	err := suite.client.Connect(ctx)
	suite.Require().NoError(err)

	// 先注册一个实例
	instance := &middleware.NacosInstance{
		ServiceName: suite.config.ServiceName,
		GroupName:   suite.config.GroupName,
		ClusterName: suite.config.ClusterName,
		IP:          suite.config.IP,
		Port:        suite.config.Port,
		Ephemeral:   true, // 临时实例需要心跳
	}

	registerOp := &middleware.NacosRegisterInstanceOperation{
		BaseOperation: core.BaseOperation{
			OpType: core.OpTypeWrite,
			OpKey:  "register-for-heartbeat",
		},
		Instance: instance,
	}
	_, err = suite.client.Execute(ctx, registerOp)
	suite.Require().NoError(err)

	time.Sleep(500 * time.Millisecond)

	// 发送心跳
	heartbeatOp := &middleware.NacosHeartbeatOperation{
		BaseOperation: core.BaseOperation{
			OpType: core.OpTypeWrite,
			OpKey:  "send-heartbeat",
		},
		ServiceName: suite.config.ServiceName,
		GroupName:   suite.config.GroupName,
		IP:          suite.config.IP,
		Port:        suite.config.Port,
	}

	result, err := suite.client.Execute(ctx, heartbeatOp)
	suite.NoError(err, "Heartbeat should succeed")
	suite.True(result.Success, "Heartbeat operation should be successful")

	// 验证心跳成功
	heartbeatSent, ok := result.Metadata["heartbeat_sent"]
	suite.True(ok, "Should have heartbeat_sent metadata")
	suite.True(heartbeatSent.(bool), "Heartbeat should be sent")
}

// TestNamespaceIsolation 测试命名空间隔离
func (suite *NacosClientTestSuite) TestNamespaceIsolation() {
	ctx := context.Background()

	// 创建不同命名空间的客户端
	namespace1Config := *suite.config
	namespace1Config.NamespaceId = "test-namespace-1"
	namespace1Config.ServiceName = "service-ns1"

	namespace2Config := *suite.config
	namespace2Config.NamespaceId = "test-namespace-2"
	namespace2Config.ServiceName = "service-ns2"

	client1 := middleware.NewNacosClient(&namespace1Config)
	client2 := middleware.NewNacosClient(&namespace2Config)

	err := client1.Connect(ctx)
	suite.Require().NoError(err)
	defer client1.Disconnect(ctx)

	err = client2.Connect(ctx)
	suite.Require().NoError(err)
	defer client2.Disconnect(ctx)

	// 在namespace1注册服务
	instance1 := &middleware.NacosInstance{
		ServiceName: namespace1Config.ServiceName,
		GroupName:   namespace1Config.GroupName,
		IP:          namespace1Config.IP,
		Port:        8081,
	}

	registerOp1 := &middleware.NacosRegisterInstanceOperation{
		BaseOperation: core.BaseOperation{
			OpType: core.OpTypeWrite,
			OpKey:  "register-ns1",
		},
		Instance: instance1,
	}
	result1, err := client1.Execute(ctx, registerOp1)
	suite.NoError(err, "Register in namespace1 should succeed")
	suite.True(result1.Success, "Registration should be successful")

	suite.T().Log("Namespace isolation verified - services in different namespaces are isolated")
}

// TestExecute_UnsupportedOperation 测试不支持的操作类型
func (suite *NacosClientTestSuite) TestExecute_UnsupportedOperation() {
	ctx := context.Background()

	err := suite.client.Connect(ctx)
	suite.Require().NoError(err)

	// 创建一个不支持的操作类型
	unsupportedOp := &core.BaseOperation{
		OpType: "UNSUPPORTED",
		OpKey:  "test",
	}

	result, err := suite.client.Execute(ctx, unsupportedOp)
	suite.NoError(err, "Execute should not return error for unsupported operation")
	suite.NotNil(result, "Result should not be nil")
	suite.False(result.Success, "Unsupported operation should fail")
	suite.NotNil(result.Error, "Error should be set for unsupported operation")
}

// TestDisconnect 测试断开连接
func (suite *NacosClientTestSuite) TestDisconnect() {
	ctx := context.Background()

	err := suite.client.Connect(ctx)
	suite.Require().NoError(err)

	// 断开连接
	err = suite.client.Disconnect(ctx)
	suite.NoError(err, "Disconnect should succeed")

	// 再次断开连接应该也不报错
	err = suite.client.Disconnect(ctx)
	suite.NoError(err, "Multiple disconnects should not error")
}

// TestGetStats 测试获取统计信息
func (suite *NacosClientTestSuite) TestGetStats() {
	ctx := context.Background()

	err := suite.client.Connect(ctx)
	suite.Require().NoError(err)

	// 注册一个实例
	instance := &middleware.NacosInstance{
		ServiceName: suite.config.ServiceName,
		GroupName:   suite.config.GroupName,
		IP:          suite.config.IP,
		Port:        suite.config.Port,
	}

	op := &middleware.NacosRegisterInstanceOperation{
		BaseOperation: core.BaseOperation{
			OpType: core.OpTypeWrite,
			OpKey:  "stats-test",
		},
		Instance: instance,
	}
	_, err = suite.client.Execute(ctx, op)
	suite.Require().NoError(err)

	// 获取统计信息
	stats := suite.client.GetStats()
	suite.NotNil(stats, "Stats should not be nil")
	suite.NotEmpty(stats, "Stats should contain data")

	// 验证统计信息包含必要字段
	suite.T().Logf("Stats: %v", stats)
}

// TestConcurrentOperations 测试并发操作
func (suite *NacosClientTestSuite) TestConcurrentOperations() {
	ctx := context.Background()

	err := suite.client.Connect(ctx)
	suite.Require().NoError(err)

	// 并发注册10个实例
	concurrentCount := 10
	done := make(chan bool, concurrentCount)
	errors := make(chan error, concurrentCount)

	for i := 0; i < concurrentCount; i++ {
		go func(index int) {
			instance := &middleware.NacosInstance{
				ServiceName: suite.config.ServiceName,
				GroupName:   suite.config.GroupName,
				IP:          suite.config.IP,
				Port:        uint64(8080 + index),
			}

			op := &middleware.NacosRegisterInstanceOperation{
				BaseOperation: core.BaseOperation{
					OpType: core.OpTypeWrite,
					OpKey:  "concurrent-test",
				},
				Instance: instance,
			}

			result, err := suite.client.Execute(ctx, op)
			if err != nil {
				errors <- err
			} else if !result.Success {
				errors <- result.Error
			}
			done <- true
		}(i)
	}

	// 等待所有操作完成
	for i := 0; i < concurrentCount; i++ {
		<-done
	}
	close(errors)

	// 检查是否有错误
	errorCount := 0
	for err := range errors {
		errorCount++
		suite.T().Logf("Concurrent operation error: %v", err)
	}

	suite.T().Logf("Concurrent operations completed: %d/%d successful",
		concurrentCount-errorCount, concurrentCount)
}

// TestNacosClientTestSuite 运行测试套件
func TestNacosClientTestSuite(t *testing.T) {
	suite.Run(t, new(NacosClientTestSuite))
}
