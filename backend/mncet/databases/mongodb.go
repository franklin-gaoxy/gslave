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

func (m *Mongodb) AddHosts(hosts *[]tools.Hosts) bool {
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
		var existingHost tools.Hosts
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
func (m *Mongodb) QueryHosts(key string, value string) *[]tools.Hosts {
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

	var hosts []tools.Hosts
	for cursor.Next(context.TODO()) {
		var host tools.Hosts
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

func (m *Mongodb) AddTasks() {

}

func (m *Mongodb) QueryTasks() {

}
