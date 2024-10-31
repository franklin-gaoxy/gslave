package mncet

import (
	"github.com/gin-gonic/gin"
	"k8s.io/klog"
	"mncet/mncet/databases"
	"mncet/mncet/mncet/operationhost"
	"mncet/mncet/mncet/servertools"
	"mncet/mncet/tools"
	"net/http"
	"strconv"
)

type kv struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func AddHost(c *gin.Context, database databases.Databases) {
	/*
		先记录 然后即可返回成功 接下来获取主机信息 重新记录
	*/
	var HostInfo []tools.HostInfo
	if err := c.ShouldBindJSON(&HostInfo); err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}
	klog.V(1).Infof("AddHost: %v", HostInfo)
	// check args
	for _, host := range HostInfo {
		if host.Address == "" || host.Hostname == "" {
			c.JSON(400, gin.H{"status": "failed", "error info": "address or hostname is empty!"})
		}
		if host.Login.Username == "" || host.Login.Password == "" {
			// 如果没有用户名密码检查sshkey是否为空
			if host.Login.SSHKey == "" {
				c.JSON(400, gin.H{"status": "failed", "error info": "username and password or sshkey is empty!"})
				return
			}
		}
	}

	if !database.AddHosts(&HostInfo) {
		c.JSON(500, gin.H{"error": "add host failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": 200})
	// start conn hosts
	operationhost.GetHostMeta(&HostInfo)
	// 记录更新后的信息
	klog.V(8).Info("Start updating information.")
	klog.V(8).Info(database.AddHosts(&HostInfo))
}

func UpdateHost(c *gin.Context, database databases.Databases) {
	// 检查查询条件是否为空
	var kvalue kv
	if err := c.ShouldBindJSON(&kvalue); err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}
	hosts := database.QueryHosts(kvalue.Key, kvalue.Value)

	//然后调用update
	if !database.AddHosts(hosts) {
		c.JSON(500, gin.H{"error": "add host failed,exec func AddHosts error!"})
		klog.V(6).Infoln("add hosts failed,exec func AddHosts error,The return value is not true.")
		return
	}
	// 如果key value 一个为空一个有值 那么则错误
	if kvalue.Key == "" || kvalue.Value == "" {
		c.JSON(400, gin.H{"status": "failed", "error info": "The key and value are either all empty or all have values"})
	}

	// 如果key value都是空 那么则更新所有
	if kvalue.Key == "" && kvalue.Value == "" {
		c.JSON(200, gin.H{"status": "success", "updatehost": "all"})
	}
	c.JSON(http.StatusOK, gin.H{"status": 200})
	klog.V(4).Infof("update hosts success,key is %s,value is %s", kvalue.Key, kvalue.Value)
	return
}

func DeleteHost(c *gin.Context, database databases.Databases) {
	// delete host
	var kvalue kv
	if err := c.ShouldBindJSON(&kvalue); err != nil {
		c.JSON(400, gin.H{"status": "failed", "error info:": err.Error()})
		return
	}
	klog.V(8).Infof("[gin_routes.go:DeleteHost]: get args is {%s:%s}", kvalue.Key, kvalue.Value)
	if !database.DeleteHost(kvalue.Key, kvalue.Value) {
		c.JSON(500, gin.H{"status": "failed", "error info:": kvalue.Key + " delete failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": 200})
	return
}

func TaskAdd(c *gin.Context, database databases.Databases) {
	// 获取提交的yaml
	var tav tools.TemplateAndValues
	if err := c.ShouldBindJSON(&tav); err != nil {
		klog.V(8).Infof("Abnormal parameter retrieval!error:%v", err)
		c.JSON(400, gin.H{"status": "failed", "describe": "Abnormal parameter retrieval.", "error info:": err.Error()})
		return
	}

	// 检查参数不为空
	if tav.TemplateData == "" && tav.ValuesData == "" {
		c.JSON(400, gin.H{"status": "failed", "describe": "TemplateData is empty!"})
		return
	}

	// 开始格式化yaml变量 同时转换参数类型
	status, content, err := servertools.FormatYamlContent([]byte(tav.TemplateData), []byte(tav.ValuesData))
	if status == false {
		klog.V(8).Infof("Format template content error:%v", err)
		c.JSON(400, gin.H{"status": "failed", "describe": "Format template content error!", "error info:": err.Error()})
		return
	}

	// 检查所有主机是否都存在
	for _, v := range content.ExecutionList {
		if _, err := servertools.CheckHostExist(&v.Stages, database); err != nil {
			c.JSON(500, gin.H{"status": "failed", "describe": "Check host exits failed!"})
			return
		}
	}

	// 将现有提交的yaml保存到mongodb
	if !database.SaveTasksTemplate(&tav) {
		c.JSON(500, gin.H{"status": "failed", "describe": "Save tasks failed!"})
		return
	}

	// 返回提交成功
	c.JSON(http.StatusOK, gin.H{"status": 200})
	return
}

func TaskRun(c *gin.Context, database databases.Databases) {
	// 获取参数
	var runtaskargs tools.RunTask
	if err := c.ShouldBindJSON(&runtaskargs); err != nil {
		c.JSON(400, gin.H{"status": "failed", "describe": "Abnormal parameter retrieval.", "error info:": err.Error()})
		return
	}

	// 检查不为空
	if runtaskargs.TaskName == "" {
		c.JSON(400, gin.H{"status": "failed", "describe": "Task Name is empty!"})
		return
	}

	// 根据taskName字段 从数据库找出保存的模板和变量
	tav := database.QueryTasks(runtaskargs.TaskName)
	klog.V(8).Infof("[gin_routes.go:TaskRun]:query tasks info:%v", tav)
	if tav == nil {
		c.JSON(400, gin.H{"status": "failed", "describe": "query database is nil! please check task name!"})
		return
	}

	// 创建一个序号ID
	ID, err := database.GenerateID()
	if err != nil {
		c.JSON(500, gin.H{"status": "failed", "describe": "Generate tasks failed!", "error info:": err.Error()})
		return
	}

	// 开始运行任务 传递开始位置等参数 不等待直接 返回状态
	c.JSON(http.StatusOK, gin.H{"status": 200, "task id": ID})
	go operationhost.ExecuteTasks(&ID, &runtaskargs, tav, database)
	return
}

func TaskGet(c *gin.Context, database databases.Databases) {
	idStr := c.Query("id")

	// 将 id 转换为整数类型
	id, err := strconv.Atoi(idStr)
	if err != nil {
		// 如果转换失败，返回一个错误信息
		c.JSON(http.StatusBadRequest, gin.H{
			"status":      "failed",
			"describe":    "ID conversion error",
			"error info:": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, database.QueryTaskResult(&id))
	return
}
