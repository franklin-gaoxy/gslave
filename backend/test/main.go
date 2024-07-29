package main

// func Loginit() {
// 	// 设置日志格式为文本格式
// 	log.SetFormatter(&log.TextFormatter{
// 		// 是否显示完整时间戳
// 		FullTimestamp: true,
// 		// 时间戳格式
// 		TimestampFormat: "2006-01-02 15:04:05",
// 		// 是否显示日志级别
// 		DisableLevelTruncation: true,
// 	})

// 	// 设置日志级别为Debug
// 	log.SetLevel(log.DebugLevel)

// 	// 将日志输出到标准输出
// 	log.SetOutput(os.Stdout)
// }

// func main() {
// 	Loginit()
// 	log.WithFields(log.Fields{
// 		"animal": "walrus",
// 		"size":   10,
// 	}).Info("A group of walrus emerges from the ocean")

// 	log.Info("this is a info log!")
// 	log.Warning("this is a warning!")
// 	log.Error("this is a error!")

// 	fmt.Println("end.")
// }

import (
	"fmt"
	"log"
	"os"

	"golang.org/x/crypto/ssh"
)

// 主机信息
type host struct {
	hostname string
	username string
	password string
	port     int
	key      string
	keyfile  string
	ip       string
}

// 会话信息
type session struct {
	session  *ssh.Session
	hostname string
	ip       string
}

func createSession(hostSlice []host) []session {
	var sessionSlice []session
	for _, h := range hostSlice {
		config := &ssh.ClientConfig{
			User: h.username,
			Auth: []ssh.AuthMethod{
				ssh.Password(h.password),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}
		client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", h.ip, h.port), config)
		if err != nil {
			log.Println("Failed to dial: ", err)
			continue
		}
		client_session, err := client.NewSession()
		if err != nil {
			log.Println("Failed to create session: ", err)
			continue
		}
		tmpSessionSlice := session{
			session:  client_session,
			hostname: h.hostname,
			ip:       h.ip,
		}
		sessionSlice = append(sessionSlice, tmpSessionSlice)

	}
	return sessionSlice
}

func main() {
	var hostSlice []host = []host{
		{
			ip:       "192.168.137.100",
			username: "root",
			password: "1qaz@WSX",
			port:     22,
			hostname: "debian",
		},
		{
			ip:       "192.168.137.200",
			username: "root",
			password: "1qaz@WSX",
			port:     22,
			hostname: "centos",
		},
	}
	for _, s := range createSession(hostSlice) {
		output, err := s.session.CombinedOutput("ls -l /root/")
		if err != nil {
			log.Fatal("Failed to run: ", err)
		}
		fmt.Printf("================= %v =================\n", s.hostname)
		os.Stdout.Write(output)
		fmt.Printf("======================================\n")
	}
	fmt.Print("end.")
}
