package mncet

import "fmt"

/*
这里是cobra启动函数 实际运行代码逻辑的地方
这里通过前面的--config参数获取了一个配置文件的路径 并且传递给了启动函数
*/

// 默认命令执行
func NewStart(configFilePath string) {
	fmt.Println(configFilePath)
}

// init 命令执行
func initEnvironment(configFilePath string) {
	fmt.Println(configFilePath)
}
