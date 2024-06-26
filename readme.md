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
    command: "curl 10.0.0.1/init.sh|bash"
    concurrentMode: concurrent
    encounteredAnError: true
  - stage:
    name: installer nginx
    hosts: ["node1"] # or "all"
    command: "apt-get install nginx -y"
    concurrentMode: concurrent
    encounteredAnError: false
  - stage:
    name: start service
    hosts: ["node2","node3"] # or "all"
    command: "java -jar app.jar"
    concurrentMode: concurrent
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
> command: 要执行的命令
>
> concurrentMode: 执行模式，concurrent为并行，所有主机同时开始，serial(串行)一个一个执行，batch(批次)每次执行几个机器
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
/status/{number}/status :对应序号任务的执行状态 {"task":[{"name":"init","status":"successed","log":"xxx"},{"name":"install nginx","status":"running","log":"xxx"}]}
/add/task :提交任务 yaml or json 返回 {"status":"successed","number": 2,"log":""}
/add/host :添加主机 {"hostname": "node1","ip":"10.0.0.10", "username":"root","password":"","sshkey":""}
/add/restart/{number}/{number} :重新从某个任务的阶段开始向下
/add/rerun/{number}/{number} :重新运行某个任务失败的阶段
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



