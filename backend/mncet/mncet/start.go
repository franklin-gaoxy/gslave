package mncet

import (
	"fmt"
	"io/ioutil"
	"mncet/mncet/databases"
	"mncet/mncet/mncet/servertools"
	"mncet/mncet/tools"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
	"k8s.io/klog"
)

/*
这里是cobra启动函数 实际运行代码逻辑的地方
这里通过前面的--config参数获取了一个配置文件的路径 并且传递给了启动函数
*/

// init 命令执行
func initEnvironment(configFilePath string) {
	fmt.Println(configFilePath)
}

// 默认命令执行
func NewStart(configFilePath string) {
	config := readConfig(configFilePath)

	// fmt.Println(config)
	klog.V(3).Infof("config: %+v\n", config)
	// init database
	// 获取数据库连接
	database := NewDatabase(config.Database.DataBaseType)
	if database == nil {
		klog.Fatal("database Not initialized correctly, is nil!")
	}
	database.Init(config)

	// register plugin
	servertools.RegisterPlugin()
	// start gin server
	startGinServer(int(config.Port), database)

}

func readConfig(configFilePath string) tools.ServerConfig {
	var config tools.ServerConfig
	yamlFile, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		klog.Fatalf("Error reading YAML file: %s\n", err)
	}

	// 解析YAML文件内容到结构体
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		klog.Fatalf("Error parsing YAML file: %s\n", err)
	}

	return config
}

// return database interface
func NewDatabase(databaseType string) databases.Databases {
	return databases.NewDatabases(databaseType)
}

func startGinServer(port int, database databases.Databases) {
	var route *gin.Engine
	route = gin.Default()

	// binding interface
	route.GET("/status/information", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "normal",
		})
	})

	// add host
	route.POST("/host/add", func(c *gin.Context) {
		AddHost(c, database)
	})
	// update host infor
	route.POST("/host/update", func(c *gin.Context) {
		UpdateHost(c, database)
	})

	route.POST("/host/delete", func(c *gin.Context) {
		DeleteHost(c, database)
	})

	route.POST("/task/add", func(c *gin.Context) {
		TaskAdd(c, database)
	})

	route.POST("/task/run", func(c *gin.Context) {
		TaskRun(c, database)
	})

	// 根据ID查询任务运行的状态
	route.GET("/task/get", func(c *gin.Context) {
		TaskGet(c, database)
	})

	klog.V(1).Infof("start gin server on port %d", port)
	_ = route.Run(fmt.Sprintf(":%d", port))
}
