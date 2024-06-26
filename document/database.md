
# database

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