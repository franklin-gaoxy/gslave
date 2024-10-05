package operationhost

import (
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/ssh"
	"k8s.io/klog"
	"mncet/mncet/tools"
	"strconv"
	"strings"
	"sync"
	"time"
)

func GetHostMeta(hosts *[]tools.Hosts) {
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

func GetHostMetaWorker(host *tools.Hosts, wg *sync.WaitGroup, pool chan struct{}) {
	defer func() {
		wg.Done()
		pool <- struct{}{}
		klog.Infof("get host %s meta worker finished", host.Address)
	}()

	klog.Infof("connect to host : %s", host.Address)
	client, status := sshRemoteHost(host)
	if !status {
		klog.Infof("connect to host fail: %s", host.Address)
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
		fmt.Printf("- Name: %s, Size: %d, Mountpoint: %s\n", device.Name, device.Size, device.Mountpoint)
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

func sshRemoteHost(host *tools.Hosts) (*ssh.Client, bool) {
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
