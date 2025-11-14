package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"middleware-chaos-testing/internal/core"
)

// MongoDBClient MongoDB客户端实现
type MongoDBClient struct {
	config     *MongoDBConfig
	client     *mongo.Client
	database   *mongo.Database
	collection *mongo.Collection
	logger     *Logger
}

// NewMongoDBClient 创建新的MongoDB客户端
func NewMongoDBClient(config *MongoDBConfig) *MongoDBClient {
	config.ApplyDefaults()
	logger := NewLogger("MongoDBClient", false)
	logger.Info("Creating new MongoDB client: uri=%s database=%s collection=%s",
		config.URI, config.Database, config.Collection)

	return &MongoDBClient{
		config: config,
		logger: logger,
	}
}

// Connect 连接到MongoDB
func (m *MongoDBClient) Connect(ctx context.Context) error {
	m.logger.Info("Connecting to MongoDB...")

	// 创建客户端选项
	clientOpts := options.Client().
		ApplyURI(m.config.URI).
		SetConnectTimeout(m.config.Timeout).
		SetServerSelectionTimeout(m.config.Timeout)

	// 配置连接池
	if m.config.MaxPoolSize > 0 {
		clientOpts.SetMaxPoolSize(uint64(m.config.MaxPoolSize))
	}
	if m.config.MinPoolSize > 0 {
		clientOpts.SetMinPoolSize(uint64(m.config.MinPoolSize))
	}
	if m.config.MaxIdleTime > 0 {
		clientOpts.SetMaxConnIdleTime(m.config.MaxIdleTime)
	}

	// 配置副本集
	if m.config.ReplicaSet != "" {
		clientOpts.SetReplicaSet(m.config.ReplicaSet)
	}

	// 配置读偏好
	if m.config.ReadPreference != "" {
		var readPref *readpref.ReadPref
		var err error
		switch m.config.ReadPreference {
		case "primary":
			readPref = readpref.Primary()
		case "secondary":
			readPref = readpref.Secondary()
		case "nearest":
			readPref = readpref.Nearest()
		default:
			readPref = readpref.Primary()
		}
		if err == nil {
			clientOpts.SetReadPreference(readPref)
		}
	}

	// 配置压缩
	if len(m.config.Compressors) > 0 {
		clientOpts.SetCompressors(m.config.Compressors)
	}

	// 连接MongoDB
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		m.logger.Error("Failed to connect to MongoDB: %v", err)
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	m.client = client
	m.database = client.Database(m.config.Database)
	if m.config.Collection != "" {
		m.collection = m.database.Collection(m.config.Collection)
	}

	m.logger.Info("Successfully connected to MongoDB")
	m.logger.Info("Connection pool configured: maxPoolSize=%d minPoolSize=%d",
		m.config.MaxPoolSize, m.config.MinPoolSize)

	return nil
}

// Disconnect 断开连接
func (m *MongoDBClient) Disconnect(ctx context.Context) error {
	m.logger.Info("Disconnecting from MongoDB...")

	if m.client == nil {
		m.logger.Debug("Client already disconnected")
		return nil
	}

	err := m.client.Disconnect(ctx)
	if err != nil {
		m.logger.Error("Failed to disconnect: %v", err)
		return fmt.Errorf("failed to disconnect: %w", err)
	}

	m.client = nil
	m.database = nil
	m.collection = nil

	m.logger.Info("Successfully disconnected from MongoDB")
	return nil
}

// Execute 执行操作
func (m *MongoDBClient) Execute(ctx context.Context, op core.Operation) (*core.Result, error) {
	startTime := time.Now()
	m.logger.Debug("Executing operation: type=%s key=%s", op.Type(), op.Key())

	var result *core.Result

	switch v := op.(type) {
	case *MongoDBInsertOperation:
		result = m.executeInsert(ctx, v, startTime)
	case *MongoDBFindOperation:
		result = m.executeFind(ctx, v, startTime)
	case *MongoDBUpdateOperation:
		result = m.executeUpdate(ctx, v, startTime)
	case *MongoDBDeleteOperation:
		result = m.executeDelete(ctx, v, startTime)
	case *MongoDBAggregateOperation:
		result = m.executeAggregate(ctx, v, startTime)
	default:
		duration := time.Since(startTime)
		opErr := fmt.Errorf("unsupported operation type: %T", op)
		m.logger.Error("Operation failed: %v", opErr)
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
	m.logger.LogOperation(opLog)

	return result, nil
}

// executeInsert 执行插入操作
func (m *MongoDBClient) executeInsert(ctx context.Context, op *MongoDBInsertOperation, startTime time.Time) *core.Result {
	if m.collection == nil {
		duration := time.Since(startTime)
		return core.NewResult(false, duration, fmt.Errorf("collection not initialized"))
	}

	m.logger.Debug("Inserting document: key=%s", op.Key())

	// 执行插入
	insertResult, err := m.collection.InsertOne(ctx, op.Document)
	duration := time.Since(startTime)

	if err != nil {
		m.logger.Error("Failed to insert document: key=%s error=%v duration=%v",
			op.Key(), err, duration)
		return core.NewResult(false, duration, fmt.Errorf("failed to insert document: %w", err))
	}

	m.logger.Debug("Document inserted successfully: key=%s id=%v duration=%v",
		op.Key(), insertResult.InsertedID, duration)

	result := core.NewResult(true, duration, nil)
	result.Metadata["inserted_id"] = insertResult.InsertedID
	return result
}

// executeFind 执行查询操作
func (m *MongoDBClient) executeFind(ctx context.Context, op *MongoDBFindOperation, startTime time.Time) *core.Result {
	if m.collection == nil {
		duration := time.Since(startTime)
		return core.NewResult(false, duration, fmt.Errorf("collection not initialized"))
	}

	m.logger.Debug("Finding document: key=%s filter=%v", op.Key(), op.Filter)

	// 执行查询
	var document bson.M
	err := m.collection.FindOne(ctx, op.Filter).Decode(&document)
	duration := time.Since(startTime)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			// 查询不到文档不算错误
			m.logger.Debug("No document found: key=%s duration=%v", op.Key(), duration)
			result := core.NewResult(true, duration, nil)
			result.Metadata["found"] = false
			return result
		}
		m.logger.Error("Failed to find document: key=%s error=%v duration=%v",
			op.Key(), err, duration)
		return core.NewResult(false, duration, fmt.Errorf("failed to find document: %w", err))
	}

	m.logger.Debug("Document found successfully: key=%s duration=%v", op.Key(), duration)

	result := core.NewResult(true, duration, nil)
	// 将document序列化为JSON
	if data, err := json.Marshal(document); err == nil {
		result.Data = data
	}
	result.Metadata["found"] = true
	return result
}

// executeUpdate 执行更新操作
func (m *MongoDBClient) executeUpdate(ctx context.Context, op *MongoDBUpdateOperation, startTime time.Time) *core.Result {
	if m.collection == nil {
		duration := time.Since(startTime)
		return core.NewResult(false, duration, fmt.Errorf("collection not initialized"))
	}

	m.logger.Debug("Updating document: key=%s filter=%v", op.Key(), op.Filter)

	// 执行更新
	updateResult, err := m.collection.UpdateOne(ctx, op.Filter, op.Update)
	duration := time.Since(startTime)

	if err != nil {
		m.logger.Error("Failed to update document: key=%s error=%v duration=%v",
			op.Key(), err, duration)
		return core.NewResult(false, duration, fmt.Errorf("failed to update document: %w", err))
	}

	m.logger.Debug("Document updated successfully: key=%s matched=%d modified=%d duration=%v",
		op.Key(), updateResult.MatchedCount, updateResult.ModifiedCount, duration)

	result := core.NewResult(true, duration, nil)
	result.Metadata["matched_count"] = updateResult.MatchedCount
	result.Metadata["modified_count"] = updateResult.ModifiedCount
	return result
}

// executeDelete 执行删除操作
func (m *MongoDBClient) executeDelete(ctx context.Context, op *MongoDBDeleteOperation, startTime time.Time) *core.Result {
	if m.collection == nil {
		duration := time.Since(startTime)
		return core.NewResult(false, duration, fmt.Errorf("collection not initialized"))
	}

	m.logger.Debug("Deleting document: key=%s filter=%v", op.Key(), op.Filter)

	// 执行删除
	deleteResult, err := m.collection.DeleteOne(ctx, op.Filter)
	duration := time.Since(startTime)

	if err != nil {
		m.logger.Error("Failed to delete document: key=%s error=%v duration=%v",
			op.Key(), err, duration)
		return core.NewResult(false, duration, fmt.Errorf("failed to delete document: %w", err))
	}

	m.logger.Debug("Document deleted successfully: key=%s deleted=%d duration=%v",
		op.Key(), deleteResult.DeletedCount, duration)

	result := core.NewResult(true, duration, nil)
	result.Metadata["deleted_count"] = deleteResult.DeletedCount
	return result
}

// executeAggregate 执行聚合操作
func (m *MongoDBClient) executeAggregate(ctx context.Context, op *MongoDBAggregateOperation, startTime time.Time) *core.Result {
	if m.collection == nil {
		duration := time.Since(startTime)
		return core.NewResult(false, duration, fmt.Errorf("collection not initialized"))
	}

	m.logger.Debug("Executing aggregate: key=%s pipeline_stages=%d", op.Key(), len(op.Pipeline))

	// 执行聚合
	cursor, err := m.collection.Aggregate(ctx, op.Pipeline)
	duration := time.Since(startTime)

	if err != nil {
		m.logger.Error("Failed to execute aggregate: key=%s error=%v duration=%v",
			op.Key(), err, duration)
		return core.NewResult(false, duration, fmt.Errorf("failed to execute aggregate: %w", err))
	}
	defer cursor.Close(ctx)

	// 读取所有结果
	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		m.logger.Error("Failed to decode aggregate results: key=%s error=%v", op.Key(), err)
		return core.NewResult(false, duration, fmt.Errorf("failed to decode results: %w", err))
	}

	m.logger.Debug("Aggregate executed successfully: key=%s results=%d duration=%v",
		op.Key(), len(results), duration)

	result := core.NewResult(true, duration, nil)
	// 将results序列化为JSON
	if data, err := json.Marshal(results); err == nil {
		result.Data = data
	}
	result.Metadata["result_count"] = len(results)
	return result
}

// Ping 检查连接是否正常
func (m *MongoDBClient) Ping(ctx context.Context) error {
	if m.client == nil {
		m.logger.Error("Ping failed: client not initialized")
		return fmt.Errorf("client not initialized")
	}

	m.logger.Debug("Pinging MongoDB...")

	err := m.client.Ping(ctx, readpref.Primary())
	if err != nil {
		m.logger.Error("Ping failed: %v", err)
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	m.logger.Debug("Ping successful")
	return nil
}

// GetStats 获取统计信息
func (m *MongoDBClient) GetStats() map[string]interface{} {
	stats := make(map[string]interface{})

	if m.client == nil {
		return stats
	}

	// 获取数据库统计信息
	if m.database != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var dbStats bson.M
		err := m.database.RunCommand(ctx, bson.D{{Key: "dbStats", Value: 1}}).Decode(&dbStats)
		if err == nil {
			stats["database_stats"] = dbStats
			m.logger.Debug("Database stats retrieved: %v", dbStats)
		} else {
			m.logger.Warn("Failed to get database stats: %v", err)
		}
	}

	// 添加配置信息
	stats["max_pool_size"] = m.config.MaxPoolSize
	stats["min_pool_size"] = m.config.MinPoolSize
	stats["read_preference"] = m.config.ReadPreference
	stats["write_concern"] = m.config.WriteConcern

	return stats
}
