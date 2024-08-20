# mncet

## brief introduction

mncet(Multi node command execution tool :多节点命令执行工具)

## use

### demo

```yaml
taskName: "install service"
recordLog:
  file: /var/log/mncet
commandList:
  - stage:
    name: init
    hosts: ["node1","node2","node3"] # or "all"
    group: ["allnode"]
    command: "curl 10.0.0.1/init.sh|bash"
    hostconcurrentMode: concurrent
    stepMode: "serial|background"
    encounteredAnError: true
  - stage:
    name: installer nginx
    hosts: ["node1"] # or "all"
    group: ["allnode"]
    command: "apt-get install nginx -y"
    hostconcurrentMode: concurrent
    stepMode: "serial|background"
    encounteredAnError: false
  - stage:
    name: start service
    hosts: ["node2","node3"] # or "all"
    command: "java -jar app.jar"
    hostconcurrentMode: concurrent
    stepMode: "serial|background"
    encounteredAnError: true
    uploadFile: 
      fromNetwork: "https://10.0.0.1/app.jar"
      fileSystem: /data/installer_package/app.jar
```



> recordLog: 是否记录日志
>
> commandList: 命令列表
>
> stage: 多个步骤
>
> name: 步骤名称
>
> hosts: 要在那些主机上执行
>
> group: 要在那些主机组上执行,和主机可以同时存在
>
> command: 要执行的命令
>
> hostconcurrentMode: 当前步骤在主机的执行模式，concurrent为并行，所有主机同时开始，serial(串行)一个一个执行，batch(批次)每次执行几个机器
>
> stepMode: 当前阶段的运行模式 serial当前阶段执行完成执行下一个 background无需等待执行完成即可执行下一个
>
> encounteredAnError: 遇到错误是否继续执行，false不继续直接退出。
>
> uploadFile: 上传文件到执行命令的主机，可选从网络(fromNetwork)或者从安装主机(fileSystem)


# backend

## interface

```
/status/task :在运行中的任务数量和信息 {"task number": 1,[{"id":112,"taskName":"install task","status":"success","runTime":"xxxx"}]}
/status/hosts :所有主机信息 {"hosts": ["hostname": "node1","ip":"10.0.0.10","cpu":10,"memory": 24,"disk":100]}
/status/information :mncet的状态信息 {"status":"normal"}
/status/{number}/status :对应序号任务的执行状态 {"task":[{"id":1,"name":"init","status":"successed","log":"xxx"},{"id":2,"name":"install nginx","status":"running","log":"xxx"}]}
/add/task :提交任务 yaml or json 返回 {"status":"successed","number": 2,"log":""}
/add/host :添加主机 {"hostname": "node1","ip":"10.0.0.10","username":"root","password":"","sshkey":""}
/add/alias :添加别名
/add/group : 添加组
/add/restart/{number}/{number} :重新从某个任务的阶段开始向下
/add/rerun/{number}/{number} :重新单独运行某个任务失败的阶段
```

启动配置
```yaml
port: 81
database:
  path: 10.0.0.10
  port: 3306
  username: root
  password: 1qaz@WSX
  basename: mncet
login:
  username: frnak
  password: 1qaz@WSX
```



---

# 问题

issues

用户首先添加机器,接下来连接测试确认无误记录到数据库,添加成功.

接下来用户上传installers yaml,首先检查yaml格式,然后记录yaml,同时格式化里面的主机

> 问题: 如何格式化主机?使用变量方式?使用标签方式?

用户开始执行任务,生成一个ID,返回给用户,接下来根据这个ID记录所有的执行步骤,包含状态和执行结果.

执行完成后更新对应ID的任务状态.

host info

```
{
	"hostname": "xxx",
	"address": ["xxx", "xxx"],
	"group": ["xxx", "xxx"],
	"login": {
		"username": "xx",
		"password": "xxx",
		"port": "xxx",
		"sshkey": "xxx"
	},
	"hostinfo": {
		"cpu": 10,
		"memory": 100,
		"disk": [{
			"mountpoint": "size",
			"mountpoint": "size"
		}]
	}
	"status": "active",
}
```

task info

```
{
	"taskname": "xxx",
	"taskid": "xxx",
	"stage": [{
		"stagename": "xx",
		"stageresult": "xxxx",
		"stagestatus": "xxx"
	}, {
		"stagename": "xx",
		"stageresult": "xxxx",
		"stagestatus": "xxx"
	}, {
		"stagename": "xx",
		"stageresult": "xxxx",
		"stagestatus": "xxx"
	}]
}
```

system info

```
{
	"system status": "running",
	"task": {
		"all task": "xx",
		"running task": "xxx",
		"failed task": "xxx"
	},
	"version": "1.0.0"
}
```

