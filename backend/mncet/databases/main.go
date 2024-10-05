package databases

import (
	"k8s.io/klog"
	"mncet/mncet/tools"
)

type Databases interface {
	/*
		databases interface
	*/

	// init func conn database return conn,err
	Init(config tools.ServerConfig)

	// Add Hosts
	AddHosts(*[]tools.Hosts) bool
	// query all hosts
	QueryHosts(key string, value string) *[]tools.Hosts
}

func NewDatabases(databaseType string) Databases {
	klog.Infof("database type: %v", databaseType)

	switch databaseType {
	//case "mysql":
	//	return NewMysql()
	case "mongodb":
		return NewMongodb()
	default:
		return nil
	}
}
