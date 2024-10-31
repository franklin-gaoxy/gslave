package command

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"k8s.io/klog"
	"mncet/mncet/mncet/servertools"
	"mncet/mncet/tools"
	"sync"
	"time"
)

type config struct {
	RemotePath         string
	Command            string `json:"command" yaml:"command"`
	HostConcurrentMode string `json:"hostConcurrentMode" yaml:"hostConcurrentMode"` // 此字段控制执行方式
	StepMode           string `json:"stepMode" yaml:"stepMode"`
	BetchNum           int    `json:"betchNum" yaml:"betchNum"`
}
type Command struct {
	config    config
	data      *tools.Stage
	ser       *tools.StageExecutionRecord
	stageinfo *tools.StageInfo // 存储当前阶段的所有内容 最后添加到ser
}

func (c *Command) Details() {
	klog.Infof("model: command.")
}

func (c *Command) CallMethodByType(ser *tools.StageExecutionRecord, typeName string, arg *tools.Stage) error {
	return servertools.CallMethodByName(c, ser, typeName, arg)
}

// 绑定参数
func (c *Command) ParameterBinding(ser *tools.StageExecutionRecord, data *tools.Stage) {
	// 处理可选参数
	// betachNum 只有批次处理模式才会需要此参数
	var betachNum int
	if data.Describe["betchNum"] == nil && data.Describe["hostConcurrentMode"].(string) != "batch" {
		betachNum = 0
	} else {
		betachNum = data.Describe["betchNum"].(int)
	}

	// 获取自定义参数
	c.config = config{
		Command:            data.Describe["command"].(string),
		HostConcurrentMode: data.Describe["hostConcurrentMode"].(string),
		StepMode:           data.Describe["stepMode"].(string),
		BetchNum:           betachNum,
	}
	c.data = data
	c.ser = ser
	c.stageinfo = &tools.StageInfo{HostExecuteResult: make(map[string]tools.StageHostStatus)}
}

/*
任务处理函数
*/

func (c *Command) ExecuteCommand(ser *tools.StageExecutionRecord, data *tools.Stage) error {
	/*
		ser: 整个流程的执行信息结果记录 此段记录的内容会在函数执行结束后写入数据库
		data: 用户提交的完整的yaml执行内容 hostConn为检查后替换完成的执行地址
	*/

	klog.Infof("[command.go:ExecuteCommand]: start execute stage %s.", data.Name)
	klog.V(8).Infof("[command.go:ExecuteCommand]: {{ ser }}:<<%s>>, {{ data }}:<<%s>>", ser, data)
	klog.V(8).Infof("[command.go:ExecuteCommand]: {{ data.HostsConn }}:<<%s>>", data.HostsConn)

	// 绑定参数
	c.ParameterBinding(ser, data)

	// 判断模式 创建多个连接同时连接到多个主机开始执行命令
	if c.config.HostConcurrentMode == "concurrent" {
		// 并行 所有主机同时开始
		klog.V(6).Infof("[command.go:ExecuteCommand]: %s operating mode: concurrent.", data.Name)
		c.ConcurrentCommand()
	} else if c.config.HostConcurrentMode == "serial" {
		// 串行 每次一个主机 顺序执行
		klog.V(6).Infof("[command.go:ExecuteCommand]: %s operating mode: serial.", data.Name)
	} else if c.config.HostConcurrentMode == "batch" {
		// 批次 每次一批主机执行
		klog.V(6).Infof("[command.go:ExecuteCommand]: %s operating mode: batch.", data.Name)
	} else {
		return fmt.Errorf("[command.go:ExecuteCommand]: %s operating mode: unsupported. stepMode Field is incorrect", data.Name)
	}

	// 结束收尾 更新任务运行状态
	if stageInfo, exists := c.ser.StageInfos[c.data.Name]; exists {
		// 只更新 Status 字段，保留其他字段
		stageInfo.Status = "succeed"
		c.ser.StageInfos[c.data.Name] = stageInfo
	}

	// 遇到错误需要检查是否返回错误 返回错误则退出不再运行
	return nil
}

/*
模块内部其他函数
*/

// 并发执行任务
func (c *Command) ConcurrentCommand() {
	// 循环创建goroutinue连接到对应的主机
	var wg sync.WaitGroup
	for _, host := range c.data.HostsConn {
		wg.Add(1)
		go func() {
			c.stageinfo.Time = time.Now()
			c.stageinfo.StageName = c.data.Name
			err := c.SSHHostExecuteCommand(&wg, &host, c.config.Command)
			if err != nil {
				klog.Errorf("ssh remote host (%s:%s) execute command error! error is %s.", host.Address, host.Hostname, err)
				c.stageinfo.Status = "failed"
				c.stageinfo.Event = err.Error()
			} else {
				c.stageinfo.StageRunStatus = "succeed"

			}
		}()
	}
	wg.Wait()

	// 格式化变量
	c.ser.StageInfos[c.data.Name] = *c.stageinfo
}

func (c *Command) SerialCommand() {
	// 顺序执行
	for _, host := range c.data.HostsConn {
		c.stageinfo.Time = time.Now()
		c.stageinfo.StageName = c.data.Name
		err := c.SSHHostExecuteCommand(nil, &host, c.config.Command)
		if err != nil {
			klog.Errorf("ssh remote host (%s:%s) execute command error! error is %s.", host.Address, host.Hostname, err)
			c.stageinfo.Status = "failed"
			c.stageinfo.Event = err.Error()
		} else {
			c.stageinfo.StageRunStatus = "succeed"

		}
	}

	// 格式化变量
	c.ser.StageInfos[c.data.Name] = *c.stageinfo
}

func (c *Command) BatchCommand() {
	//klog.Infof("[command.go:BatchCommand]: %s: Batch mode was used, but the method has not yet been developed and will not perform any operations.", c.data.Name)

	// 定义最大并发量 // 创建信号量（有缓冲的 channel）控制并发 goroutine 数量
	var sem chan struct{}
	if c.config.BetchNum == 0 {
		sem = make(chan struct{}, 5)
	} else {
		sem = make(chan struct{}, c.config.BetchNum)
	}

	var wg sync.WaitGroup

	for _, host := range c.data.HostsConn {
		wg.Add(1)

		// 发送信号，占用信号量中的一个位置
		sem <- struct{}{}

		// 启动 goroutine
		go func(host tools.HostInfo) {
			defer func() {
				// 完成后释放信号量
				<-sem
				wg.Done()
			}()

			c.stageinfo.Time = time.Now()
			c.stageinfo.StageName = c.data.Name
			err := c.SSHHostExecuteCommand(&wg, &host, c.config.Command)
			if err != nil {
				klog.Errorf("ssh remote host (%s:%s) execute command error! error is %s.", host.Address, host.Hostname, err)
				c.stageinfo.Status = "failed"
				c.stageinfo.Event = err.Error()
			} else {
				c.stageinfo.StageRunStatus = "succeed"
			}
		}(host) // 将 host 作为参数传递给 goroutine，避免闭包问题
	}

	// 等待所有 goroutine 执行完毕
	wg.Wait()

	// 格式化变量
	c.ser.StageInfos[c.data.Name] = *c.stageinfo

}

func (c *Command) SSHHostExecuteCommand(wg *sync.WaitGroup, host *tools.HostInfo, command string) error {
	// 连接到远程主机执行命令
	defer func() {
		// 如果是concurrent模式 才会传入wg参数
		if wg != nil {
			wg.Done()
		}
		klog.Infof("ssh host %s meta worker finished.", host.Address)
	}()

	klog.Infof("connect to host : %s", host.Address)
	client, status := c.sshRemoteHost(host)
	if !status {
		klog.Infof("connect to host fail: %s", host.Address)
		return fmt.Errorf("failed!SSH connection to remote host failed,connect host to %s", host.Address)
	}

	// 执行命令 获取返回结果
	result, err := c.runCommand(client, command)

	// 执行状态信息格式化 传入命令执行返回的内容
	c.stageinfo.HostExecuteResult[c.config.RemotePath] = tools.StageHostStatus{
		Result: result,
	}

	return err
}

// 执行命令 返回执行后结果
func (c *Command) runCommand(client *ssh.Client, command string) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("[command.go:runCommand]: failed to create session: %w", err)
	}
	defer func() {
		if err := session.Close(); err != nil {
			klog.V(6).Infof("[command.go:runCommand]: close session error: %v", err)
		}
	}()

	output, err := session.CombinedOutput(command)
	if err != nil {
		return "", fmt.Errorf("[command.go:runCommand]: failed to run command: %w", err)
	}

	return string(output), nil
}

// 连接远程主机
func (c *Command) sshRemoteHost(host *tools.HostInfo) (*ssh.Client, bool) {
	//connect to host
	var authMethod ssh.AuthMethod
	if host.Login.Username != "" && host.Login.Password != "" {
		authMethod = ssh.Password(host.Login.Password)
	} else if host.Login.SSHKey != "" {
		keys, _ := ssh.ParsePrivateKey([]byte(host.Login.SSHKey))
		authMethod = ssh.PublicKeys(keys)
	}
	config := &ssh.ClientConfig{
		User: host.Login.Username,
		Auth: []ssh.AuthMethod{
			//ssh.Password(host.Login.Password),
			authMethod,
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	// 如果IP地址为空那么则使用主机名尝试
	var addr string
	if host.Address != "" {
		addr = fmt.Sprintf("%s:%d", host.Address, host.Login.Port)
		c.config.RemotePath = host.Address
	} else {
		addr = fmt.Sprintf("%s:%d", host.Hostname, host.Login.Port)
		c.config.RemotePath = host.Hostname
	}

	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		klog.Errorf("ssh remote host %s failed! error is : %w", addr, err)
		return nil, false
	}
	return client, true
}
