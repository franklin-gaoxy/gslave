package file

import (
	"crypto/tls"
	"fmt"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io"
	"k8s.io/klog"
	"mncet/mncet/mncet/servertools"
	"mncet/mncet/tools"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type config struct {
	From               string `json:"from" yaml:"from"`               // 从当前主机查找
	FromNetwork        string `json:"fromNetwork" yaml:"fromNetwork"` // 从网络下载
	To                 string `json:"to" yaml:"to"`                   // 发送到对端主机的存储位置
	SSLVerify          bool   `json:"sslVerify" yaml:"sslVerify"`     // 从网络下载时 是否禁用ssl认证 默认开启
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
	// 处理可选参数
	if data.Describe["from"] != nil && data.Describe["fromNetwork"] != nil {
		klog.Errorf("[file.go:ParameterBinding]: from and fromNetwork is nil!")
		return
	}
	var from string
	if data.Describe["from"] != nil {
		from = data.Describe["from"].(string)
	}

	// 检查从网络下载选项是否为空
	var fromNetwork string
	if data.Describe["fromNetwork"] != nil {
		fromNetwork = data.Describe["fromNetwork"].(string)
	}

	// 处理从网络下载 检查是否设置了ssl 默认为true
	var ssl bool = true
	if data.Describe["fromNetwork"] != nil {
		if data.Describe["sslVerify"] != nil {
			ssl = data.Describe["sslVerify"].(bool)
		}
	}

	// 获取自定义参数
	f.conf = config{
		From:               from,
		FromNetwork:        fromNetwork,
		To:                 data.Describe["to"].(string),
		HostConcurrentMode: data.Describe["hostConcurrentMode"].(string),
		SSLVerify:          ssl,
	}
	f.data = data
	f.ser = ser
	f.stageinfo = &tools.StageInfo{HostExecuteResult: make(map[string]tools.StageHostStatus)}
}

/*
任务处理函数
支持函数: RemoteFile LocalFiles
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

	// 结束收尾 更新任务运行状态
	if stageInfo, exists := f.ser.StageInfos[f.data.Name]; exists {
		// 只更新 Status 字段，保留其他字段
		stageInfo.Status = "succeed"
		f.ser.StageInfos[f.data.Name] = stageInfo
	}

	return nil
}

func (f *File) RemoteFile(ser *tools.StageExecutionRecord, args *tools.Stage) error {
	klog.Infof("[file.go:RemoteFile]: start execute %s!", args.Name)
	f.ParameterBinding(ser, args)

	// download file
	fileName := f.downloadFile()
	f.conf.From = fileName

	// 开始获取主机
	switch f.conf.HostConcurrentMode {
	case "concurrent":
		klog.V(6).Infof("[file.go:RemoteFile]: mode:concurrent start execute %s!", args.Name)
		// 同时执行
		var wg sync.WaitGroup
		for _, host := range args.HostsConn {
			wg.Add(1)
			go f.CopyFileToRemote(&host, &wg)
		}
		wg.Wait()
	case "serial":
		klog.V(6).Infof("[file.go:RemoteFile]: mode:serial start execute %s!", args.Name)
		// 顺序执行
		for _, host := range args.HostsConn {
			f.CopyFileToRemote(&host, nil)
		}
	}

	// 结束收尾 更新任务运行状态
	if stageInfo, exists := f.ser.StageInfos[f.data.Name]; exists {
		// 只更新 Status 字段，保留其他字段
		stageInfo.Status = "succeed"
		f.ser.StageInfos[f.data.Name] = stageInfo
	}

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
	// 设置任务
	f.stageinfo.Time = time.Now()
	f.stageinfo.StageName = f.data.Name

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

		// check error
		if err != nil {
			f.stageinfo.Status = "failed"
			f.stageinfo.Event = err.Error()
		}
		f.ser.StageInfos[f.data.Name] = *f.stageinfo
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

func (f *File) downloadFile() string {
	// 发起 GET 请求
	var client *http.Client
	if f.conf.SSLVerify == false {
		client = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // 禁用证书验证
			},
		}
	} else {
		client = &http.Client{}
	}

	klog.V(6).Infof("[file.go:downloadFile]: start download file,url is %v", f.conf.FromNetwork)
	if f.conf.FromNetwork == "" {
		klog.Errorf("from network is none!")
		return ""
	}

	response, err := client.Get(f.conf.FromNetwork)
	if err != nil {
		klog.V(6).Infof("[file.go:downloadFile]: Failed to download file: %v\n", err)
		return ""
	}

	// 检查 HTTP 响应状态码
	if response.StatusCode != http.StatusOK {
		klog.V(6).Infof("[file.go:downloadFile]: Failed to download file, return status is: %v\n", response.Status)
		return ""
	}

	// 创建本地文件
	// 解析 URL
	parsedURL, err := url.Parse(f.conf.FromNetwork)
	if err != nil {
		klog.V(6).Infof("[file.go:downloadFile]: Failed to parse URL: %v\n", err)
		return ""
	}

	// 使用 path.Base 提取路径中的文件名
	fileName := path.Base(parsedURL.Path)

	localFile, err := os.Create(fileName)
	if err != nil {
		klog.V(6).Infof("[file.go:downloadFile]: Failed to create file: %v\n", err)
		return ""
	}

	// 将 HTTP 响应体中的内容复制到本地文件
	_, err = io.Copy(localFile, response.Body)
	if err != nil {
		klog.V(6).Infof("[file.go:downloadFile]: Failed to save file: %v\n", err)
		return ""
	}

	//fmt.Println("File downloaded successfully!")
	klog.V(6).Infof("[file.go:downloadFile]: File downloaded successfully!\n")
	defer func() {
		if err := response.Body.Close(); err != nil {
			klog.Errorf("[file.go:downloadFile]: close response body failed! exit.")
		}
		if err := localFile.Close(); err != nil {
			klog.Errorf("[file.go:downloadFile]: close local file failed! exit.")
		}
	}()

	return fileName
}
