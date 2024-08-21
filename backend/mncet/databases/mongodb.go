package databases

import (
	"context"
	"mncet/mncet/tools"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"k8s.io/klog"
)

type Mongodb struct {
	client *mongo.Client
}

func NewMongodb() *Mongodb {
	return &Mongodb{}
}
func (m *Mongodb) init() {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		klog.V(5).Fatal(err)
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		klog.V(5).Fatal(err)
	}

	klog.V(5).Info("Successfully connected to MongoDB")
}

func (m *Mongodb) AddHosts([]tools.Hosts) {
	m.client.Database("mncet").Collection("hosts").InsertOne(context.TODO(), "")
}
func (m *Mongodb) QueryHosts() {

}

func (m *Mongodb) AddTasks() {

}

func (m *Mongodb) QueryTasks() {

}
