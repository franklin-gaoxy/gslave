package databases

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"mncet/mncet/tools"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"k8s.io/klog"
)

type Mongodb struct {
	client   *mongo.Client
	database string
}

func NewMongodb() *Mongodb {
	//var mongo *Mongodb
	return &Mongodb{}
}
func (m *Mongodb) Init(serverConfig tools.ServerConfig) {
	var err error
	klog.V(5).Info("begin exec MongoDB init")
	klog.Info(serverConfig)

	// format database
	m.database = serverConfig.Database.BaseName

	// 如果连接串不为空则格式化 构建 MongoDB 连接字符串
	var connpath string
	if serverConfig.Database.ConnPath != "" {
		connpath = fmt.Sprintf("mongodb://%s:%s@%s:%d/%s?authSource=%s", serverConfig.Database.Description.Username,
			serverConfig.Database.Description.Password, serverConfig.Database.Host,
			serverConfig.Database.Port, serverConfig.Database.BaseName, serverConfig.Database.AuthSource)
	} else {
		connpath = serverConfig.Database.ConnPath
	}

	klog.V(8).Infoln("Connecting to string:", connpath)
	clientOptions := options.Client().ApplyURI(connpath)
	m.client, err = mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		klog.Fatal(connpath, err)
	}

	err = m.client.Ping(context.TODO(), nil)
	if err != nil {
		klog.Fatal(err)
	}

	klog.V(5).Info("Successfully connected to MongoDB")
}

func (m *Mongodb) AddHosts(hosts *[]tools.HostInfo) bool {
	klog.V(8).Infoln("AddHosts function start, args: hosts: ", hosts)

	collection := m.client.Database(m.database).Collection("hosts")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	for _, host := range *hosts {
		filter := bson.M{
			"$or": []bson.M{
				{"hostname": host.Hostname},
				{"address": host.Address},
			},
		}
		var existingHost tools.HostInfo
		err := collection.FindOne(ctx, filter).Decode(&existingHost)
		if errors.Is(err, mongo.ErrNoDocuments) {
			// 不存在，插入新的主机信息
			_, insertErr := collection.InsertOne(ctx, host)
			if insertErr != nil {
				klog.Errorf("Failed to insert host: %s, Error: %v", host.Hostname, insertErr)
				return false
			}
			klog.V(8).Infof("Inserted new host, hostname is : %s, ip address is %s", host.Hostname, host.Address)
		} else if err != nil {
			// 查找过程中出现错误
			klog.Errorf("Failed to find host: %s, Error: %v", host.Hostname, err)
			return false
		} else {
			// 存在，执行更新
			update := bson.M{
				"$set": host, // 更新整个主机信息
			}
			_, updateErr := collection.UpdateOne(ctx, filter, update)
			if updateErr != nil {
				klog.Errorf("Failed to update host: %s, Error: %v", host.Hostname, updateErr)
				return false
			}
			klog.V(8).Infof("Updated existing host: %s", host.Hostname)
		}
	}
	return true
}
func (m *Mongodb) QueryHosts(key string, value string) *[]tools.HostInfo {
	collection := m.client.Database(m.database).Collection("hosts")
	// 检查key是否为all
	var cursor *mongo.Cursor
	var err error
	if key == "" && value == "" {
		cursor, err = collection.Find(context.TODO(), bson.D{})
	} else {
		filter := bson.M{key: value}
		cursor, err = collection.Find(context.TODO(), filter)
	}

	//cursor, err := collection.Find(context.TODO(), bson.D{})
	if err != nil {
		klog.V(6).Infoln("Find error: ", err)
	}
	defer func() {
		if err := cursor.Close(context.TODO()); err != nil {
			klog.V(6).Infoln("Find Close error: ", err)
		}
	}()

	var hosts []tools.HostInfo
	for cursor.Next(context.TODO()) {
		var host tools.HostInfo
		if err := cursor.Decode(&host); err != nil {
			klog.V(6).Infoln("Decode error: ", err)
		}
		hosts = append(hosts, host)
	}

	// check cursor error
	if err := cursor.Err(); err != nil {
		klog.V(6).Infoln("Cursor error: ", err)
	}
	return &hosts
}

func (m *Mongodb) DeleteHost(key string, value string) bool {
	klog.V(8).Infof("[mongodb.go:DeleteHost]: start remove host {%s:%s} .", key, value)
	collection := m.client.Database(m.database).Collection("hosts")
	filter := bson.M{key: value}
	result, err := collection.DeleteOne(context.TODO(), filter)
	if err != nil {
		klog.Error("Error deleting host:", err)
		return false
	}

	// 如果删除的文档数量大于 0，说明删除成功
	if result.DeletedCount > 0 {
		klog.Infof("remove host success,remove number: %d", result.DeletedCount)
		return true
	}

	// 如果不大于0 则没有数据被删除
	klog.Warningln("0 data were deleted")
	return true
}

func (m *Mongodb) SaveTasksTemplate(data *tools.TemplateAndValues) bool {
	klog.V(6).Info("start save task to mongodb ...")
	collection := m.client.Database(m.database).Collection("tasks")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"taskName": data.TaskName,
	}
	var existingTasks tools.Tasks
	err := collection.FindOne(ctx, filter).Decode(&existingTasks)
	if errors.Is(err, mongo.ErrNoDocuments) {
		// 不存在，插入新的任务
		_, insertErr := collection.InsertOne(ctx, data)
		if insertErr != nil {
			klog.Errorf("Failed to insert task: %s, Error: %v", data.TaskName, insertErr)
			return false
		}
		klog.V(8).Infof("Inserted new host, task name is : %s", data.TaskName)
	} else if err != nil {
		// 查找过程中出现错误
		klog.Errorf("Failed to find host: %s, Error: %v", data.TaskName, err)
		return false
	} else {
		// 存在，执行更新
		update := bson.M{
			"$set": data, // 更新整个主机信息
		}
		_, updateErr := collection.UpdateOne(ctx, filter, update)
		if updateErr != nil {
			klog.Errorf("Failed to update task: %s, Error: %v", data.TaskName, updateErr)
			return false
		}
		klog.V(8).Infof("Updated existing task: %s", data.TaskName)
	}

	return true
}

func (m *Mongodb) QueryTasks(TaskName string) *tools.TemplateAndValues {
	// init
	klog.V(8).Infof("[mongodb.go:QueryTasks]: start query tasks by task name %s.", TaskName)
	collection := m.client.Database(m.database).Collection("tasks")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// query
	filter := bson.M{
		"taskName": TaskName,
	}
	var TaskTemplate tools.TemplateAndValues
	err := collection.FindOne(ctx, filter).Decode(&TaskTemplate)
	if errors.Is(err, mongo.ErrNoDocuments) {
		// 不存在 返回错误
		klog.Errorf("Failed to find tasks: %s", TaskName)
		return nil
	} else if err != nil {
		// 查找过程中出现错误
		klog.Errorf("Failed to find task: %s, Error: %v", TaskName, err)
	}
	// 存在 返回
	klog.V(8).Infof("find task: %s success!", TaskName)
	return &TaskTemplate
}

func (m *Mongodb) GenerateID() (int, error) {
	// 选择数据库和集合
	collection := m.client.Database(m.database).Collection("taskNumber")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 尝试查找TaskNumber文档
	filter := bson.M{"_id": "taskNumber"}
	update := bson.M{
		"$inc": bson.M{"sequence_value": 1}, // 每次调用自增1
	}
	options := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)

	var updatedDoc bson.M
	err := collection.FindOneAndUpdate(ctx, filter, update, options).Decode(&updatedDoc)
	if err != nil {
		klog.Errorf("failed to generate,err: %v", err)
		return 0, err
	}

	// 返回更新后的数字
	if value, ok := updatedDoc["sequence_value"].(int32); ok {
		return int(value), nil
	}

	klog.Errorf("failed to generate id: %d")
	return 0, fmt.Errorf("failed to generate ID")
}

func (m *Mongodb) SaveTaskResult(StageExecutionRecord *tools.StageExecutionRecord) bool {
	// init
	klog.V(8).Infof("[mongodb.go:SaveTaskResult]: start save task result ID: %d.",
		StageExecutionRecord.TaskID)
	collection := m.client.Database(m.database).Collection("stageexecutionresults")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// query
	filter := bson.M{
		"ID": StageExecutionRecord.TaskID,
	}
	var TmpSER tools.StageExecutionRecord
	err := collection.FindOne(ctx, filter).Decode(&TmpSER)
	if errors.Is(err, mongo.ErrNoDocuments) {
		// 不存在，插入新的任务
		_, insertErr := collection.InsertOne(ctx, StageExecutionRecord)
		if insertErr != nil {
			klog.Errorf("Failed to insert task record: %d, Error: %v",
				StageExecutionRecord.TaskID, insertErr)
			return false
		}
		klog.V(8).Infof("Inserted new task record, ID name is : %d", StageExecutionRecord.TaskID)
	} else if err != nil {
		// 查找过程中出现错误
		klog.Errorf("Failed to find task record ID: %d, Error: %v", StageExecutionRecord.TaskID, err)
		return false
	} else {
		// 存在，执行更新
		update := bson.M{
			"$set": StageExecutionRecord, // 更新整个主机信息
		}
		_, updateErr := collection.UpdateOne(ctx, filter, update)
		if updateErr != nil {
			klog.Errorf("Failed to update task record: %d, Error: %v", StageExecutionRecord.TaskID, updateErr)
			return false
		}
		klog.V(8).Infof("Updated existing task record: %s", StageExecutionRecord.TaskID)
	}
	return true
}

// 查询运行的历史记录
func (m *Mongodb) QueryTaskResult(ID *int) *tools.StageExecutionRecord {
	// init
	var stageExecutionRecord tools.StageExecutionRecord
	klog.V(8).Infof("[mongodb.go:QueryTaskResult]: start query tasks record by task ID %s.", ID)
	collection := m.client.Database(m.database).Collection("stageexecutionresults")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// query
	filter := bson.M{
		"ID": ID,
	}
	err := collection.FindOne(ctx, filter).Decode(&stageExecutionRecord)
	if errors.Is(err, mongo.ErrNoDocuments) {
		// 不存在 返回错误
		klog.Errorf("Failed to find tasks record ID: %s", ID)
		return nil
	} else if err != nil {
		// 查找过程中出现错误
		klog.Errorf("Failed to find task record ID: %s, Error: %v", ID, err)
	}
	// 存在 返回
	klog.V(8).Infof("find task record ID: %s success!", ID)

	return &stageExecutionRecord
}
