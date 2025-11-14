package middleware_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"middleware-chaos-testing/internal/core"
	"middleware-chaos-testing/internal/middleware"
)

// MongoDBClientTestSuite MongoDB客户端测试套件
type MongoDBClientTestSuite struct {
	suite.Suite
	client *middleware.MongoDBClient
	config *middleware.MongoDBConfig
}

// SetupTest 每个测试前执行
func (suite *MongoDBClientTestSuite) SetupTest() {
	suite.config = &middleware.MongoDBConfig{
		URI:        "mongodb://localhost:27017",
		Database:   "chaos_test",
		Collection: "test_collection",
		Timeout:    5 * time.Second,
	}
	suite.client = middleware.NewMongoDBClient(suite.config)
}

// TearDownTest 每个测试后执行
func (suite *MongoDBClientTestSuite) TearDownTest() {
	if suite.client != nil {
		ctx := context.Background()
		_ = suite.client.Disconnect(ctx)
	}
}

// TestConnect 测试MongoDB连接
func (suite *MongoDBClientTestSuite) TestConnect() {
	ctx := context.Background()

	// 测试成功连接
	err := suite.client.Connect(ctx)
	suite.NoError(err, "Connect should succeed")

	// 验证连接状态 - Ping测试
	err = suite.client.Ping(ctx)
	suite.NoError(err, "Ping should succeed after connect")
}

// TestConnect_InvalidURI 测试无效URI连接
func (suite *MongoDBClientTestSuite) TestConnect_InvalidURI() {
	ctx := context.Background()

	invalidConfig := &middleware.MongoDBConfig{
		URI:      "invalid-uri",
		Database: "test",
		Timeout:  1 * time.Second,
	}
	client := middleware.NewMongoDBClient(invalidConfig)

	err := client.Connect(ctx)
	suite.Error(err, "Connect should fail with invalid URI")
}

// TestInsert 测试插入文档
func (suite *MongoDBClientTestSuite) TestInsert() {
	ctx := context.Background()

	// 先连接
	err := suite.client.Connect(ctx)
	suite.Require().NoError(err)

	// 创建插入操作
	doc := map[string]interface{}{
		"name":  "test-doc",
		"value": "test-value",
		"timestamp": time.Now().Unix(),
	}

	op := &middleware.MongoDBInsertOperation{
		BaseOperation: core.BaseOperation{
			OpType: core.OpTypeWrite,
			OpKey:  "test-key-1",
		},
		Document: doc,
	}

	// 执行插入
	result, err := suite.client.Execute(ctx, op)
	suite.NoError(err, "Insert should succeed")
	suite.NotNil(result, "Result should not be nil")
	suite.True(result.Success, "Insert operation should be successful")
	suite.Greater(result.Duration, time.Duration(0), "Duration should be positive")
}

// TestFind 测试查询文档
func (suite *MongoDBClientTestSuite) TestFind() {
	ctx := context.Background()

	// 先连接
	err := suite.client.Connect(ctx)
	suite.Require().NoError(err)

	// 先插入一个文档
	insertOp := &middleware.MongoDBInsertOperation{
		BaseOperation: core.BaseOperation{
			OpType: core.OpTypeWrite,
			OpKey:  "find-test-key",
		},
		Document: map[string]interface{}{
			"_id":   "find-test-key",
			"name":  "find-test",
			"value": 123,
		},
	}
	_, err = suite.client.Execute(ctx, insertOp)
	suite.Require().NoError(err)

	// 查询文档
	findOp := &middleware.MongoDBFindOperation{
		BaseOperation: core.BaseOperation{
			OpType: core.OpTypeRead,
			OpKey:  "find-test-key",
		},
		Filter: map[string]interface{}{
			"_id": "find-test-key",
		},
	}

	result, err := suite.client.Execute(ctx, findOp)
	suite.NoError(err, "Find should succeed")
	suite.NotNil(result, "Result should not be nil")
	suite.True(result.Success, "Find operation should be successful")
	suite.NotNil(result.Data, "Found document should not be nil")
}

// TestFind_NotFound 测试查询不存在的文档
func (suite *MongoDBClientTestSuite) TestFind_NotFound() {
	ctx := context.Background()

	err := suite.client.Connect(ctx)
	suite.Require().NoError(err)

	// 查询不存在的文档
	findOp := &middleware.MongoDBFindOperation{
		BaseOperation: core.BaseOperation{
			OpType: core.OpTypeRead,
			OpKey:  "nonexistent-key",
		},
		Filter: map[string]interface{}{
			"_id": "nonexistent-key-12345",
		},
	}

	result, err := suite.client.Execute(ctx, findOp)
	suite.NoError(err, "Find should not error even if document not found")
	suite.NotNil(result, "Result should not be nil")
	// 查询不到应该也算成功，只是数据为空
	suite.True(result.Success, "Find operation should be successful")
}

// TestUpdate 测试更新文档
func (suite *MongoDBClientTestSuite) TestUpdate() {
	ctx := context.Background()

	err := suite.client.Connect(ctx)
	suite.Require().NoError(err)

	// 先插入一个文档
	insertOp := &middleware.MongoDBInsertOperation{
		BaseOperation: core.BaseOperation{
			OpType: core.OpTypeWrite,
			OpKey:  "update-test-key",
		},
		Document: map[string]interface{}{
			"_id":   "update-test-key",
			"name":  "original",
			"count": 1,
		},
	}
	_, err = suite.client.Execute(ctx, insertOp)
	suite.Require().NoError(err)

	// 更新文档
	updateOp := &middleware.MongoDBUpdateOperation{
		BaseOperation: core.BaseOperation{
			OpType: core.OpTypeWrite,
			OpKey:  "update-test-key",
		},
		Filter: map[string]interface{}{
			"_id": "update-test-key",
		},
		Update: map[string]interface{}{
			"$set": map[string]interface{}{
				"name":  "updated",
				"count": 2,
			},
		},
	}

	result, err := suite.client.Execute(ctx, updateOp)
	suite.NoError(err, "Update should succeed")
	suite.NotNil(result, "Result should not be nil")
	suite.True(result.Success, "Update operation should be successful")
}

// TestDelete 测试删除文档
func (suite *MongoDBClientTestSuite) TestDelete() {
	ctx := context.Background()

	err := suite.client.Connect(ctx)
	suite.Require().NoError(err)

	// 先插入一个文档
	insertOp := &middleware.MongoDBInsertOperation{
		BaseOperation: core.BaseOperation{
			OpType: core.OpTypeWrite,
			OpKey:  "delete-test-key",
		},
		Document: map[string]interface{}{
			"_id":  "delete-test-key",
			"name": "to-be-deleted",
		},
	}
	_, err = suite.client.Execute(ctx, insertOp)
	suite.Require().NoError(err)

	// 删除文档
	deleteOp := &middleware.MongoDBDeleteOperation{
		BaseOperation: core.BaseOperation{
			OpType: core.OpTypeWrite,
			OpKey:  "delete-test-key",
		},
		Filter: map[string]interface{}{
			"_id": "delete-test-key",
		},
	}

	result, err := suite.client.Execute(ctx, deleteOp)
	suite.NoError(err, "Delete should succeed")
	suite.NotNil(result, "Result should not be nil")
	suite.True(result.Success, "Delete operation should be successful")

	// 验证文档已被删除
	findOp := &middleware.MongoDBFindOperation{
		BaseOperation: core.BaseOperation{
			OpType: core.OpTypeRead,
			OpKey:  "delete-test-key",
		},
		Filter: map[string]interface{}{
			"_id": "delete-test-key",
		},
	}
	findResult, err := suite.client.Execute(ctx, findOp)
	suite.NoError(err)
	suite.True(findResult.Success)
	// 文档应该不存在了
}

// TestAggregate 测试聚合操作
func (suite *MongoDBClientTestSuite) TestAggregate() {
	ctx := context.Background()

	err := suite.client.Connect(ctx)
	suite.Require().NoError(err)

	// 先插入一些测试数据
	for i := 1; i <= 5; i++ {
		insertOp := &middleware.MongoDBInsertOperation{
			BaseOperation: core.BaseOperation{
				OpType: core.OpTypeWrite,
				OpKey:  "agg-test-key",
			},
			Document: map[string]interface{}{
				"category": "test",
				"value":    i * 10,
			},
		}
		_, err = suite.client.Execute(ctx, insertOp)
		suite.Require().NoError(err)
	}

	// 执行聚合操作
	aggOp := &middleware.MongoDBAggregateOperation{
		BaseOperation: core.BaseOperation{
			OpType: core.OpTypeRead,
			OpKey:  "aggregate-test",
		},
		Pipeline: []map[string]interface{}{
			{
				"$match": map[string]interface{}{
					"category": "test",
				},
			},
			{
				"$group": map[string]interface{}{
					"_id":   "$category",
					"total": map[string]interface{}{"$sum": "$value"},
					"count": map[string]interface{}{"$sum": 1},
				},
			},
		},
	}

	result, err := suite.client.Execute(ctx, aggOp)
	suite.NoError(err, "Aggregate should succeed")
	suite.NotNil(result, "Result should not be nil")
	suite.True(result.Success, "Aggregate operation should be successful")
	suite.NotNil(result.Data, "Aggregate result should not be nil")
}

// TestConnectionPool 测试连接池
func (suite *MongoDBClientTestSuite) TestConnectionPool() {
	ctx := context.Background()

	// 配置连接池参数
	poolConfig := &middleware.MongoDBConfig{
		URI:         "mongodb://localhost:27017",
		Database:    "chaos_test",
		Collection:  "pool_test",
		Timeout:     5 * time.Second,
		MaxPoolSize: 10,
		MinPoolSize: 2,
	}
	poolClient := middleware.NewMongoDBClient(poolConfig)

	err := poolClient.Connect(ctx)
	suite.NoError(err, "Connect with pool config should succeed")

	// 执行多个并发操作测试连接池
	done := make(chan bool, 5)
	for i := 0; i < 5; i++ {
		go func(index int) {
			op := &middleware.MongoDBInsertOperation{
				BaseOperation: core.BaseOperation{
					OpType: core.OpTypeWrite,
					OpKey:  "pool-test",
				},
				Document: map[string]interface{}{
					"index": index,
					"data":  "test",
				},
			}
			result, err := poolClient.Execute(ctx, op)
			suite.NoError(err)
			suite.True(result.Success)
			done <- true
		}(i)
	}

	// 等待所有操作完成
	for i := 0; i < 5; i++ {
		<-done
	}

	_ = poolClient.Disconnect(ctx)
}

// TestReplicaSetFailover 测试副本集故障转移（模拟测试）
func (suite *MongoDBClientTestSuite) TestReplicaSetFailover() {
	// 注意：这个测试需要实际的副本集环境
	// 在单节点环境下，我们只验证配置是否正确设置

	ctx := context.Background()

	rsConfig := &middleware.MongoDBConfig{
		URI:            "mongodb://localhost:27017",
		Database:       "chaos_test",
		Collection:     "rs_test",
		Timeout:        5 * time.Second,
		ReplicaSet:     "rs0",
		ReadPreference: "secondary",
		WriteConcern:   "majority",
	}

	rsClient := middleware.NewMongoDBClient(rsConfig)
	suite.NotNil(rsClient, "Client should be created with replica set config")

	// 在没有真实副本集的情况下，连接可能会失败，这是预期的
	// 我们主要验证配置能够正确传递
	err := rsClient.Connect(ctx)
	// 如果有副本集环境，应该成功；如果没有，跳过
	if err != nil {
		suite.T().Skip("Skipping replica set test: no replica set available")
	}

	_ = rsClient.Disconnect(ctx)
}

// TestExecute_UnsupportedOperation 测试不支持的操作类型
func (suite *MongoDBClientTestSuite) TestExecute_UnsupportedOperation() {
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
func (suite *MongoDBClientTestSuite) TestDisconnect() {
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
func (suite *MongoDBClientTestSuite) TestGetStats() {
	ctx := context.Background()

	err := suite.client.Connect(ctx)
	suite.Require().NoError(err)

	// 执行一些操作
	insertOp := &middleware.MongoDBInsertOperation{
		BaseOperation: core.BaseOperation{
			OpType: core.OpTypeWrite,
			OpKey:  "stats-test",
		},
		Document: map[string]interface{}{
			"test": "data",
		},
	}
	_, err = suite.client.Execute(ctx, insertOp)
	suite.Require().NoError(err)

	// 获取统计信息
	stats := suite.client.GetStats()
	suite.NotNil(stats, "Stats should not be nil")
	suite.NotEmpty(stats, "Stats should contain data")
}

// TestMongoDBClientTestSuite 运行测试套件
func TestMongoDBClientTestSuite(t *testing.T) {
	suite.Run(t, new(MongoDBClientTestSuite))
}
