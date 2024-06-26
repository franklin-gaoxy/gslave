package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
)

func Loginit() {
	// 设置日志格式为文本格式
	log.SetFormatter(&log.TextFormatter{
		// 是否显示完整时间戳
		FullTimestamp: true,
		// 时间戳格式
		TimestampFormat: "2006-01-02 15:04:05",
		// 是否显示日志级别
		DisableLevelTruncation: true,
	})

	// 设置日志级别为Debug
	log.SetLevel(log.DebugLevel)

	// 将日志输出到标准输出
	log.SetOutput(os.Stdout)
}

func main() {
	Loginit()
	log.WithFields(log.Fields{
		"animal": "walrus",
		"size":   10,
	}).Info("A group of walrus emerges from the ocean")

	log.Info("this is a info log!")
	log.Warning("this is a warning!")
	log.Error("this is a error!")
	

	fmt.Println("end.")
}
