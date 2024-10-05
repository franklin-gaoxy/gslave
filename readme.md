# mncet

## brief introduction

mncet(Multi node command execution tool :多节点命令执行工具)

## use

### 锚点

```yaml
hosts:
  - &host1 node1
  - &host2 node2
  - &host3 node3

taskName: "install service"
recordLog:
  file: /var/log/mncet
commandList:
  - stage:
      name: init
      hosts: [*host1, *host2, *host3]
      command: "curl 10.0.0.1/init.sh | bash"
      concurrentMode: concurrent
      encounteredAnError: true
  - stage:
      name: install nginx
      hosts: [*host1]
      command: "apt-get install nginx -y"
      concurrentMode: concurrent
      encounteredAnError: false
  - stage:
      name: start service
      hosts: [*host2, *host3]
      command: "java -jar app.jar"
      concurrentMode: concurrent
      encounteredAnError: true
      uploadFile:
        fromNetwork: "https://10.0.0.1/app.jar"
        fileSystem: /data/installer_package/app.jar
```



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

### demo2

```yaml
taskName: "install service"
recordLog:
  file: /var/log/mncet
list:
  - stage:
    name: init
    hosts: ["node1","node2","node3"] # or "all"
    group: ["allnode"]
    mode: command
    type: command
    describe:
      command: "curl 10.0.0.1/init.sh|bash"
      hostconcurrentMode: concurrent
      stepMode: "serial|background"
      encounteredAnError: true
  - stage:
    name: installer nginx
    hosts: ["node1"] # or "all"
    group: ["allnode"]
    mode: command
    type: command
    describe:
      command: "apt-get install nginx -y"
      hostconcurrentMode: concurrent
      stepMode: "serial|background"
      encounteredAnError: false
  - stage:
    name: start service
    hosts: ["node2","node3"] # or "all"
    mode: command
    type: command
    describe:
      command: "java -jar app.jar"
      hostconcurrentMode: concurrent
      stepMode: "serial|background"
      encounteredAnError: true
      uploadFile: 
        fromNetwork: "https://10.0.0.1/app.jar"
        fileSystem: /data/installer_package/app.jar
  - stage:
    name: put file
    hosts: ["node2","node3"] # or "all"
    mode: file
    type: local/file
    describe:
      uploadFile: 
        fromNetwork: "https://10.0.0.1/app.jar"
        fileSystem: /data/installer_package/app.jar
```



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

已经实现：

### `/status/information`

> 用于检查服务运行是否正常，无需传参，返回json

```
传参：
	无需传参
返回
	{"status": "normal"}
```

### `/host/add`

> 添加主机，传递一个数组，自动循环添加
>
> hostname 和address两个有一个为必须参数 login.username login.password和login.sshKey 两个有一个为必须参数

#### 可接受的所有参数

```golang
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
		CPU       string `yaml:"cpu" bson:"cpu"`
		Memory    string `yaml:"memory" bson:"memory"`
		Disk      []MountDisk
		TotalSize float64 `yaml:"totalSize" bson:"totalSize"`
	} `yaml:"hostInfo" bson:"hostInfo"`
	Status   string `yaml:"status" bson:"status"`
	Describe string `yaml:"describe" bson:"describe"`
}
type MountDisk struct {
	Device     string   `yaml:"device" bson:"device"`
	Name       string   `yaml:"name" bson:"name"`
	MountPoint []string `yaml:"mountpoints" bson:"mountpoints"`
	Size       int      `yaml:"size" bson:"size"`
}
```

#### demo

```
传参：
[{
    "hostname": "develop",
    "address": "192.168.0.12",
    "group": "host1",
    "login": {"username":"root","password":"1qaz@WSX","port":22}
}]
返回：
{
    "status": 200
}
```

### `/host/update`

> 更新主机CPU 内存信息,可选传参，不传参则更新所有

#### 可接受的所有参数

```golang
	type kv struct {
		key   string
		value string
	}
```



#### demo

```
传参：
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

### 启动配置

```yaml
port: 8000
database:
  databaseType: mongodb
  # connpath和host等二选一
  connPath: mongodb://192.168.0.12:27017
  host: 192.168.0.12
  port: 27017
  # 其他参数
  authSource: 
  authType: username
  description:
    username: myUserAdmin
    password: 1qaz@WSX
  basename: mncet
login:
  username: admin
  password: admin
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
			"mountpoint": "size"

		}, {
			"mountpoint": "size"
		}]
	}
	"status": "active"
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



# 可能用到

策略模式简化代码避免大量的判断

```go
type Handler interface {
	Execute()
}

type CommandHandler struct {
	CommandDescribe CommandDescribe
}

func (h *CommandHandler) Execute() {
	// 执行命令逻辑
}

type URLHandler struct {
	URLDescribe URLDescribe
}

func (h *URLHandler) Execute() {
	// 执行 URL 逻辑
}

type LocalHandler struct {
	URLDescribe URLDescribe
}

func (h *LocalHandler) Execute() {
	// 执行 Local 逻辑
}

func main() {
	var config Config

	// 假设已解析 YAML 到 config

	handlerMap := map[string]map[string]Handler{
		"command": {
			"command": &CommandHandler{CommandDescribe: config.Describe.(CommandDescribe)},
		},
		"file": {
			"file":   &URLHandler{URLDescribe: config.Describe.(URLDescribe)},
			"local": &LocalHandler{URLDescribe: config.Describe.(URLDescribe)},
		},
	}

	// 获取处理器并执行
	if handler, ok := handlerMap[config.Mode][config.Type]; ok {
		handler.Execute()
	}
}

```

kubernetes使用的策略模式

```go
type Plugin interface {
    Filter(node *Node, pod *Pod) bool
    Score(node *Node, pod *Pod) int
}

type Scheduler struct {
    plugins map[string]Plugin
}

func (s *Scheduler) Schedule(pod *Pod) {
    // 过滤阶段
    for _, node := range nodes {
        for _, plugin := range s.plugins {
            if !plugin.Filter(node, pod) {
                continue // 不满足条件，跳过
            }
        }
        // 打分阶段
        score := 0
        for _, plugin := range s.plugins {
            score += plugin.Score(node, pod)
        }
        // 记录得分
    }
}

```

