package file

import (
	"fmt"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io"
	"k8s.io/klog"
	"mncet/mncet/mncet/servertools"
	"mncet/mncet/tools"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type config struct {
	From               string `json:"from" yaml:"from"`               // 从当前主机查找
	FromNetwork        string `json:"fromNetwork" yaml:"fromNetwork"` // 从网络下载
	To                 string `json:"to" yaml:"to"`                   // 发送到对端主机的存储位置
	HostConcurrentMode string `json:"hostConcurrentMode" yaml:"hostConcurrentMode"`
}

type File struct {
	conf      config
	data      *tools.Stage
	ser       *tools.StageExecutionRecord
	stageinfo *tools.StageInfo
}

func (f *File) Details() {
	klog.Infof("mode: file.")
}

func (f *File) CallMethodByType(ser *tools.StageExecutionRecord, typeName string, arg *tools.Stage) error {
	return servertools.CallMethodByName(f, ser, typeName, arg)
}

func (f *File) ParameterBinding(ser *tools.StageExecutionRecord, data *tools.Stage) {
	// 获取自定义参数
	f.conf = config{
		From:               data.Describe["from"].(string),
		FromNetwork:        data.Describe["fromNetwork"].(string),
		To:                 data.Describe["to"].(string),
		HostConcurrentMode: data.Describe["hostConcurrentMode"].(string),
	}
	f.data = data
	f.ser = ser
	f.stageinfo = &tools.StageInfo{HostExecuteResult: make(map[string]tools.StageHostStatus)}
}

/*
任务处理函数
*/

func (f *File) LocalFiles(ser *tools.StageExecutionRecord, args *tools.Stage) error {
	klog.Infof("[file.go:LocalFiles]: start execute %s!", args.Name)
	f.ParameterBinding(ser, args)

	// 开始获取主机
	switch f.conf.HostConcurrentMode {
	case "concurrent":
		klog.V(6).Infof("[file.go:LocalFiles]: mode:concurrent start execute %s!", args.Name)
		// 同时执行
		var wg sync.WaitGroup
		for _, host := range args.HostsConn {
			wg.Add(1)
			go f.CopyFileToRemote(&host, &wg)
		}
		wg.Wait()
	case "serial":
		klog.V(6).Infof("[file.go:LocalFiles]: mode:serial start execute %s!", args.Name)
		// 顺序执行
		for _, host := range args.HostsConn {
			f.CopyFileToRemote(&host, nil)
		}
	}

	return nil
}

func (f *File) RemoteFile(ser *tools.StageExecutionRecord, args *tools.Stage) error {
	return nil
}

/*
模块内部其他函数
*/

func (f *File) CreateSSHClient(host *tools.HostInfo) (*ssh.Client, bool) {
	/*
		创建到远程主机的连接
	*/

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

func (f *File) CopyFileToRemote(host *tools.HostInfo, wg *sync.WaitGroup) {
	sshconn, s := f.CreateSSHClient(host)
	if s != true {
		klog.Errorf("[file.go:CopyFileToRemote]: create remote client failed! exit.")
		return
	}

	// create ftp client
	client, err := sftp.NewClient(sshconn)
	if err != nil {
		klog.Errorf("[file.go:CopyFileToRemote]: Failed to create SFTP client: %v", err)
		return
	}

	// 检查本地路径是否为文件或目录
	err = f.uploadPath(client, f.conf.From, f.conf.To)
	if err != nil {
		klog.Errorf("[file.go:CopyFileToRemote]: Failed to upload: %v", err)
		return
	}

	klog.V(6).Infof("[file.go:CopyFileToRemote]: Upload completed successfully!")

	// stop
	defer func() {
		// close sftp client
		err := client.Close()
		if err != nil {
			klog.Errorf("[file.go:CopyFileToRemote] close remote sftp client failed! exit.")
			return
		}

		// close ssh client
		err = sshconn.Close()
		if err != nil {
			klog.Errorf("[file.go:CopyFileToRemote] close remote ssh client failed! exit.")
			return
		}

		// check goroutinue lock
		if wg != nil {
			wg.Done()
		}
	}()
}

func (f *File) uploadPath(client *sftp.Client, localPath string, remoteDir string) error {
	/*
		检查是目录还是文件 然后调用不同方法上传
	*/

	info, err := os.Stat(localPath)
	if err != nil {
		return err
	}

	// 如果是目录，递归上传
	if info.IsDir() {
		return f.uploadDirectory(client, localPath, remoteDir, "")
	}

	// 否则上传文件
	return f.uploadFile(client, localPath, remoteDir)
}

func (f *File) uploadFile(client *sftp.Client, localFilePath string, remoteDir string) error {
	/*
		上传文件到远程主机
	*/

	// 保留文件名
	fileName := filepath.Base(localFilePath)
	remoteFilePath := filepath.ToSlash(filepath.Join(remoteDir, fileName)) // 确保远程路径使用正斜杠

	// 打开本地文件
	localFile, err := os.Open(localFilePath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %v", err)
	}
	defer func() {
		if err := localFile.Close(); err != nil {
			klog.Errorf("[file.go:uploadFile] close local file failed! exit.")
		}
	}()

	// 创建远程文件
	remoteFile, err := client.Create(remoteFilePath)
	if err != nil {
		return fmt.Errorf("failed to create remote file: %v", err)
	}
	defer func() {
		if err := remoteFile.Close(); err != nil {
			klog.Errorf("[file.go:uploadFile]: close remote file failed! exit.")
		}
	}()

	// 复制文件内容
	_, err = io.Copy(remoteFile, localFile)
	if err != nil {
		return fmt.Errorf("[file.go:uploadFile]: failed to copy file to remote: %v", err)
	}

	klog.V(4).Infof("[file.go:uploadFile]: Uploaded file: %s -> %s\n", localFilePath, remoteFilePath)
	return nil
}

func (f *File) uploadDirectory(client *sftp.Client, localDirPath string, remoteDirPath string, basePath string) error {
	/*
		上传目录到远程主机
	*/

	// 设置基础路径，用于构建相对路径
	if basePath == "" {
		basePath = localDirPath
	}

	// 设置主目录
	normalizedPath := filepath.ToSlash(localDirPath)
	parts := strings.Split(normalizedPath, "/")
	mainDir := parts[len(parts)-1]

	// 遍历本地目录
	err := filepath.Walk(localDirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 获取相对路径，并保留目录结构
		relPath, err := filepath.Rel(basePath, path)
		if err != nil {
			return err
		}

		// 构造远程路径
		remotePath := filepath.ToSlash(filepath.Join(remoteDirPath, mainDir, relPath))

		// 如果是目录，创建远程目录
		if info.IsDir() {
			if _, err := client.Stat(remotePath); os.IsNotExist(err) {
				klog.V(4).Infof("Creating remote sub-directory: %s\n", remotePath)
				err := client.Mkdir(remotePath)
				if err != nil {
					return fmt.Errorf("failed to create remote sub-directory: %v", err)
				}
			}
		} else {
			// 上传文件
			err = f.uploadFile(client, path, filepath.Dir(remotePath))
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}
