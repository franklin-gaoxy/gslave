package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"golang.org/x/crypto/ssh"
)

func connectToHost(user, password, host string, port int) (*ssh.Client, error) {
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	address := fmt.Sprintf("%s:%d", host, port)
	client, err := ssh.Dial("tcp", address, config)
	if err != nil {
		return nil, fmt.Errorf("Failed to dial: %s", err)
	}

	return client, nil
}

func uploadFile(client *ssh.Client, localFilePath, remoteFilePath string) error {
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("Failed to create session: %s", err)
	}
	defer session.Close()

	// 打开本地文件
	localFile, err := os.Open(localFilePath)
	if err != nil {
		return fmt.Errorf("Unable to open local file: %s", err)
	}
	defer localFile.Close()

	// 获取本地文件信息
	fileInfo, err := localFile.Stat()
	if err != nil {
		return fmt.Errorf("Unable to get file info: %s", err)
	}

	// 创建远程文件
	go func() {
		w, _ := session.StdinPipe()
		defer w.Close()
		fmt.Fprintf(w, "C0644 %d %s\n", fileInfo.Size(), fileInfo.Name())
		io.Copy(w, localFile)
		fmt.Fprint(w, "\x00")
	}()

	if err := session.Run(fmt.Sprintf("scp -t %s", remoteFilePath)); err != nil {
		return fmt.Errorf("Failed to run scp command: %s", err)
	}

	return nil
}

func main() {
	user := "root"
	password := "1qaz@WSX"
	host := "192.168.137.100"
	port := 22
	localFilePath := "/path/to/local/file"
	remoteFilePath := "/path/to/remote/file"

	client, err := connectToHost(user, password, host, port)
	if err != nil {
		log.Fatalf("Failed to connect to host: %s", err)
	}
	defer client.Close()

	err = uploadFile(client, localFilePath, remoteFilePath)
	if err != nil {
		log.Fatalf("Failed to upload file: %s", err)
	}

	fmt.Println("File uploaded successfully")
}
