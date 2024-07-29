package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type host struct {
	ip       string
	port     int
	username string
	password string
}

func connectToHost(h host) (*sftp.Client, error) {
	// SSH配置
	config := &ssh.ClientConfig{
		User: h.username,
		Auth: []ssh.AuthMethod{
			ssh.Password(h.password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// 连接到SSH服务器
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", h.ip, h.port), config)
	if err != nil {
		return nil, fmt.Errorf("Failed to dial: %s", err)
	}

	// 创建SFTP客户端
	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("Unable to start SFTP subsystem: %s", err)
	}

	return sftpClient, nil
}

func uploadFile(sftpClient *sftp.Client, localFilePath, remoteFilePath string) error {
	// 打开本地文件
	localFile, err := os.Open(localFilePath)
	if err != nil {
		return fmt.Errorf("Unable to open local file: %s", err)
	}
	defer localFile.Close()

	// 创建远程文件
	remoteFile, err := sftpClient.Create(remoteFilePath)
	if err != nil {
		return fmt.Errorf("Unable to create remote file: %s", err)
	}
	defer remoteFile.Close()

	// 拷贝文件内容
	bytes, err := ioutil.ReadAll(localFile)
	if err != nil {
		return fmt.Errorf("Failed to read local file: %s", err)
	}

	_, err = remoteFile.Write(bytes)
	if err != nil {
		return fmt.Errorf("Failed to write to remote file: %s", err)
	}

	return nil
}

func uploadDirectory(sftpClient *sftp.Client, localDirPath, remoteDirPath string) error {
	// 创建远程目录
	err := sftpClient.MkdirAll(remoteDirPath)
	if err != nil {
		return fmt.Errorf("Unable to create remote directory: %s", err)
	}

	// 读取本地目录中的文件和子目录
	files, err := ioutil.ReadDir(localDirPath)
	if err != nil {
		return fmt.Errorf("Unable to read local directory: %s", err)
	}

	// 遍历本地目录中的内容
	for _, file := range files {
		localFilePath := path.Join(localDirPath, file.Name())
		remoteFilePath := path.Join(remoteDirPath, file.Name())

		if file.IsDir() {
			// 递归上传子目录
			err = uploadDirectory(sftpClient, localFilePath, remoteFilePath)
			if err != nil {
				return err
			}
		} else {
			// 上传文件
			err = uploadFile(sftpClient, localFilePath, remoteFilePath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func main() {
	host := host{
		ip:       "192.168.137.100",
		port:     22,
		username: "root",
		password: "1qaz@WSX",
	}

	sftpClient, err := connectToHost(host)
	if err != nil {
		log.Fatalf("Failed to connect to host: %s", err)
	}
	defer sftpClient.Close()

	// 上传文件
	err = uploadFile(sftpClient, "/path/to/local/file", "/path/to/remote/file")
	if err != nil {
		log.Fatalf("Failed to upload file: %s", err)
	}

	// 上传目录
	err = uploadDirectory(sftpClient, "/path/to/local/directory", "/path/to/remote/directory")
	if err != nil {
		log.Fatalf("Failed to upload directory: %s", err)
	}

	log.Println("File and directory upload completed successfully")
}
