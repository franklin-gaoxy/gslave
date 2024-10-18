package operationhost

import (
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/ssh"
	"k8s.io/klog"
	"mncet/mncet/databases"
	"mncet/mncet/mncet/plugins"
	"mncet/mncet/mncet/servertools"
	"mncet/mncet/tools"
	"strconv"
	"strings"
	"sync"
	"time"
)

func GetHostMeta(hosts *[]tools.HostInfo) {
	var wg sync.WaitGroup
	var pool chan struct{} = make(chan struct{}, 10)

	for i, host := range *hosts {
		wg.Add(1)
		pool <- struct{}{}
		klog.V(8).Infof("connect to host : %s", host.Address)
		//connect to host
		// for _,host := range *hosts {}时，host依然是从hosts这个slice中拷贝的值，而不是本身的。所以传递需要&(*hosts)[i]
		go GetHostMetaWorker(&(*hosts)[i], &wg, pool)
	}
	wg.Wait()
	klog.V(8).Infoln("get host meta successfully.")
}

func GetHostMetaWorker(host *tools.HostInfo, wg *sync.WaitGroup, pool chan struct{}) {
	defer func() {
		wg.Done()
		pool <- struct{}{}
		klog.Infof("get host %s meta worker finished", host.Address)
	}()

	klog.Infof("connect to host : %s", host.Address)
	client, status := sshRemoteHost(host)
	if !status {
		klog.Infof("connect to host fail: %s", host.Address)
		host.Status = "failed"
		host.Describe = "failed!SSH connection to remote host failed."
		return
	}

	// 获取信息
	var err error
	host.HostInfo.CPU, err = runCommand(client, "nproc")
	host.HostInfo.CPU = strings.TrimSpace(host.HostInfo.CPU)
	var memory string
	var memoryint int
	memory, err = runCommand(client, "grep MemTotal /proc/meminfo | awk '{print $2}'")
	memory = strings.TrimSpace(memory)
	memoryint, err = strconv.Atoi(memory)
	memoryGB := float64(memoryint) / 1024 / 1024
	host.HostInfo.Memory = fmt.Sprintf("%.2f", memoryGB)
	// disk
	diskinfo, err := runCommand(client, "lsblk --json -b")
	var disks DiskInfo
	err = json.Unmarshal([]byte(strings.TrimSpace(diskinfo)), &disks)
	var totalsize float64 = 0
	for _, device := range disks.BlockDevices {
		klog.Infof("- Name: %s, Size: %d, Mountpoint: %s\n", device.Name, device.Size, device.Mountpoint)
		totalsize = totalsize + float64(device.Size)
	}
	host.HostInfo.TotalSize = totalsize/1024/1024/1024 - 1

	//klog.V(8).Infof("Host %s information: CPU:%s memory:%s disk total size: %.2f", host.Hostname,
	//host.HostInfo.CPU, host.HostInfo.Memory, host.HostInfo.TotalSize)

	var devices []tools.MountDisk
	for _, d := range disks.BlockDevices {
		var device tools.MountDisk
		device.Device = d.Name
		if len(d.Children) == 0 {
			device.Name = d.Name
			device.Size = d.Size
			device.MountPoint = d.Mountpoint
			devices = append(devices, device)
		} else {
			for _, child := range d.Children {
				if len(child.Children) == 0 {
					klog.V(8).Infoln(child.Name, child.Size, child.Mountpoint)
					device.Name = child.Name
					device.Size = child.Size
					device.MountPoint = child.Mountpoint
					devices = append(devices, device)
				} else {
					for _, volume := range child.Children {
						device.Name = volume.Name
						device.Size = volume.Size
						device.MountPoint = volume.Mountpoint
						devices = append(devices, device)
					}
				}
			}
			//devices = append(devices, device)
		}

	}
	host.HostInfo.Disk = devices

	if err != nil {
		host.Status = "failed"
		host.Describe = err.Error()
	} else {
		host.Status = "success"
	}
}

func runCommand(client *ssh.Client, command string) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(command)
	if err != nil {
		return "", fmt.Errorf("failed to run command: %w", err)
	}

	return string(output), nil
}

func sshRemoteHost(host *tools.HostInfo) (*ssh.Client, bool) {
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
	} else {
		addr = fmt.Sprintf("%s:%d", host.Hostname, host.Login.Port)
	}

	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		klog.Errorf("ssh remote host %s failed! error is : %w", addr, err)
		return nil, false
	}
	return client, true
}

func ExecuteTasks(ID *int, RunTaskArgs *tools.RunTask, data *tools.TemplateAndValues, dbs databases.Databases) bool {
	klog.V(6).Infof("[operation_host.go:ExecuteTasks]: ExecuteTasks start execute!")
	klog.V(8).Infof("[operation_host.go:ExecuteTasks]: All parameters passed in: {{ RunTaskArgs }}:>>%s<<\n{{ data }}:>>%s<<\n", RunTaskArgs, data)

	// 格式化yaml中的变量
	var ser tools.StageExecutionRecord // 用来记录执行过程中的信息
	ser.TaskID = *ID
	status, descriptionfile, err := servertools.FormatYamlContent([]byte(data.TemplateData), []byte(data.ValuesData))
	if status == false {
		klog.V(8).Infof("[operation_host.go:ExecuteTasks]: Format template content error:%v", err)
		return false
	}

	// 更新到数据库 对应ID任务标记为执行状态
	ser.Status = "running"
	ser.StageInfos = make(map[string]tools.StageInfo)
	klog.V(6).Infof("[operation_host.go:ExecuteTasks]: YAML formatting completed, starting loop stage.")
	klog.V(8).Infof("[operation_host.go:ExecuteTasks]: {{ descriptionfile }}:>>%s<<", descriptionfile)

	for i, v := range descriptionfile.ExecutionList {
		// 需要检查RunTaskArgs 从那个位置开始启动
		klog.V(8).Infof("[operation_host.go:ExecuteTasks]: Start checking the start and stop positions.")
		if RunTaskArgs.StartPosition != "" {
			if v.Stages.Name != RunTaskArgs.StartPosition {
				// 如果当前阶段的Stage Name 和传入的StartPosition 不同 那么直接循环下一个
				// 如果是 则开始执行
				klog.V(6).Infof("[operation_host.go:ExecuteTasks]: The starting position has been specified, but it has not yet arrived. loop again.")
				continue
			}

			// 如果StartPosition 不为空 那么则检查StopPosition
			if RunTaskArgs.StopPosition != "" && v.Stages.Name == RunTaskArgs.StopPosition {
				// 如果当前阶段的Name和StopPosition相同 那么则退出不再继续执行
				klog.V(8).Infof("[operation_host.go:ExecuteTasks]: Designated end position, arrived, exit")
				break
			}
		}
		klog.V(6).Infof("[operation_host.go:ExecuteTasks]: Task running parameters (start position and stop position) check completed.")

		// 提取此阶段需要的主机信息 格式化到v.HostInfo
		//v.HostsConn = make(map[string]tools.HostInfo)
		hosts, err := servertools.CheckHostExist(&v.Stages, dbs)
		if err != nil {
			klog.V(8).Infof("[operation_host.go:ExecuteTasks]: CheckHostExist error! error is %v", err)
		}
		v.Stages.HostsConn = *hosts

		// 创建stage 更新到数据库 某个stage开始执行 v.Name
		ser.StageInfos[v.Stages.Name] = tools.StageInfo{Status: "running"}
		// 同步数据库
		dbs.SaveTaskResult(&ser)
		klog.Infof("[operation_host.go:ExecuteTasks]: executed {{ i }}:<<%d>>, {{ v.Name }}:<<%s>>, {{ v.Mode }}:<<%s>>, {{ v.Type }}:<<%s>>", i, v.Stages.Name, v.Stages.Mode, v.Stages.Type)

		// 调用方法
		stage := plugins.CreatePlugin()[v.Stages.Mode]
		err = stage.CallMethodByType(&ser, v.Stages.Type, &v.Stages)

		// 每执行完成一个stage 就保存一次执行信息
		dbs.SaveTaskResult(&ser)

		if err != nil {
			klog.Errorf("[operation_host.go:ExecuteTasks]: execute stage:%s, mode:%s, type:%s, error:%v", v.Stages.Name, v.Stages.Mode, v.Stages.Type, err)
			// 遇到错误判断是否可以继续往下执行
			if !v.Stages.EncounteredAnError {
				klog.Infof("[operation_host.go:ExecuteTasks]: Task ID %d:Step %s is incorrect, this stage has failed and cannot continue", *ID, v.Stages.Name)
				return false
			}
		}
		klog.V(8).Infof("[operation_host.go:ExecuteTasks]: {{ ser }}:<<%s>>", ser)
	}
	ser.Status = "succeed"
	dbs.SaveTaskResult(&ser)
	return true
}
