package main

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

type Hosts struct {
	Hostname string `yaml:"hostname" bson:"hostname"`
	Address  string `yaml:"address" bson:"address"`
	Group    string `yaml:"group" bson:"group"`
	Login    struct {
		Username string `yaml:"username" bson:"username"`
		Password string `yaml:"password" bson:"password"`
		Port     int16  `yaml:"port" bson:"port"`
		SSHKey   string `yaml:"sshKey" bson:"sshKey"`
	} `yaml:"login" bson:"login"`
	HostInfo struct {
		CPU    string `yaml:"cpu" bson:"cpu"`
		Memory string `yaml:"memory" bson:"memory"`
		Disk   []MountDisk
	} `yaml:"hostInfo" bson:"hostInfo"`
	Status string `yaml:"status" bson:"status"`
}
type MountDisk struct {
	MountPoint string `yaml:"mountpoint" bson:"mountpoint"`
	Size       string `yaml:"size" bson:"size"`
}

func main() {
	var hosts Hosts
	hosts.Hostname = "example-host"
	hosts.Address = "192.168.1.1"
	hosts.Group = "example-group"
	hosts.Login.Username = "example-username"
	hosts.Login.Password = "example-password"
	hosts.Login.Port = 22
	// 设置 MongoDB 客户端连接选项
	clientOptions := options.Client().ApplyURI("mongodb://myUserAdmin:new_password@192.168.0.12:27017/mncet?authSource=admin")

	// 连接到 MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// 检查连接
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB!")

	// 选择数据库和集合
	collection := client.Database("mncet").Collection("testcollection")

	// 准备要插入的数据
	document := bson.D{
		{Key: "name", Value: "John Doe"},
		{Key: "age", Value: 30},
		{Key: "created_at", Value: time.Now()},
	}

	// 插入数据
	insertResult, err := collection.InsertOne(context.TODO(), document)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Inserted document with ID: %v\n", insertResult.InsertedID)

	// 关闭连接
	err = client.Disconnect(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connection to MongoDB closed.")
}
