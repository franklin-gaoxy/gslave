# mncet

## brief introduction

mncet(Multi node command execution tool :多节点命令执行工具)

## use

### demo

```yaml
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