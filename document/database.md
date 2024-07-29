
# database

## mysql

#### hosts

> 记录所有的主机

```
id 自增不可重复
name
group
ip 机器IP
username 连接机器的用户名
password 密码,可为空
key 秘钥
```

#### host_alias

> 记录所有的别名
> 根据主机的ID进行查询 关系为1对多

```
id 可重复
alias_name 别名
```

#### task

> 记录所有的历史记录和运行任务信息

```
id: task id,每个完整的任务(一个yaml)生成一个单独的ID 一对多关系
name: task名称
status: 运行状态
time: 最后一次的运行时间
hosts: 运行在那些主机上
group: 运行在那些主机组
```







---

#### running_task

> 记录所有运行中的总任务信息

```
id
name 总任务名称
all_hosts 需要执行的所有主机
```
#### running_task_steps

> 记录运行中的任务每个步骤的信息

```
id 和running_task的ID对应 用于查询属于哪个任务
status : failed running unknown waiting success
command: 运行的命令
put_file: 上传文件 url或者路径
```


#### history_task
> 历史任务运行状态记录

```
id
name 总任务名称
all_hosts 需要执行的所有主机
```

#### history_task_steps
```
id 和history_task的ID对应 用于查询属于哪个任务
status : failed success
command: 运行的命令
```


----

```
hosts(id name ip alias)
# 记录正在运行的任务和历史记录 alias运行主机别名
task(id stageName host alias status time)
user(username password token) 
```



```
task(id)
```

## MongoDB!

或许MongoDB更适合这个场景.

这样yaml就无需转换为二维表格,直接存储在MongoDB就可以了.

否则需要记录上传的yaml信息

1. 保存到主机,每次获取都查询解析一遍yaml
2. 创建一个宽表,记录任务的每个字段信息,创建一个task的元数据表,记录元数据信息.