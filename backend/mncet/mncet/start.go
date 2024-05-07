package mncet

import (
	"fmt"
	"io/ioutil"
	"log"
	"mncet/mncet/tools"

	"gopkg.in/yaml.v2"
)

/*
这里是cobra启动函数 实际运行代码逻辑的地方
这里通过前面的--config参数获取了一个配置文件的路径 并且传递给了启动函数
*/

// 默认命令执行
func NewStart(configFilePath string) {
	var config tools.Config
	yamlFile, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		log.Fatalf("Error reading YAML file: %s\n", err)
	}

	// 解析YAML文件内容到结构体
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		log.Fatalf("Error parsing YAML file: %s\n", err)
	}

	fmt.Println(config)

}

// init 命令执行
func initEnvironment(configFilePath string) {
	fmt.Println(configFilePath)
}
