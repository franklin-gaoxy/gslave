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

	// Add HostInfo
	AddHosts(*[]tools.HostInfo) bool
	// query all hosts
	QueryHosts(key string, value string) *[]tools.HostInfo
	// delete host
	DeleteHost(key string, value string) bool
	// save submit task content
	SaveTasksTemplate(*tools.TemplateAndValues) bool
	// query task
	QueryTasks(TaskName string) *tools.TemplateAndValues
	// save task execute result
	SaveTaskResult(StageExecutionRecord *tools.StageExecutionRecord) bool
	// generate task number
	GenerateID() (int, error)
	// by id query task describe
	QueryTaskResult(ID *int) *tools.StageExecutionRecord
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
