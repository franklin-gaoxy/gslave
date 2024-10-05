package mncet

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"k8s.io/klog"
	"mncet/mncet/databases"
	"mncet/mncet/mncet/operationhost"
	"mncet/mncet/tools"
	"net/http"
)

func AddHost(c *gin.Context, database databases.Databases) {
	/*
		先记录 然后即可返回成功 接下来获取主机信息 重新记录
	*/
	var HostInfo []tools.Hosts
	if err := c.ShouldBindJSON(&HostInfo); err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}
	klog.V(1).Info(fmt.Sprintf("AddHost: %v", HostInfo))
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
	type kv struct {
		key   string
		value string
	}
	var kvalue kv
	if err := c.ShouldBindJSON(&kvalue); err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}
	hosts := database.QueryHosts(kvalue.key, kvalue.value)

	//然后调用update
	if !database.AddHosts(hosts) {
		c.JSON(500, gin.H{"error": "add host failed,exec func AddHosts error!"})
		klog.V(6).Infoln("add hosts failed,exec func AddHosts error,The return value is not true.")
		return
	}
	// 如果key value 一个为空一个有值 那么则错误
	if kvalue.key == "" || kvalue.value == "" {
		c.JSON(400, gin.H{"status": "failed", "error info": "The key and value are either all empty or all have values"})
	}

	// 如果key value都是空 那么则更新所有
	if kvalue.key == "" && kvalue.value == "" {
		c.JSON(200, gin.H{"status": "success", "updatehost": "all"})
	}
	c.JSON(http.StatusOK, gin.H{"status": 200})
	klog.V(4).Infof("update hosts success,key is %s,value is %s", kvalue.key, kvalue.value)
	return
}
