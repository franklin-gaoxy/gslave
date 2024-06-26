package mncet

import (
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

/*
cobra相关内容
*/

// configFilePath 此变量用于接受--config参数的内容 然后传递到启动函数里
var configFilePath string

// rootCmd 主命令 也就是不加任何子命令情况 执行此函数
var rootCmd = cobra.Command{
	Use:   "config",
	Short: "input config file address.",
	Run: func(cmd *cobra.Command, args []string) {
		if checkConfigFile(configFilePath) {
			// 默认启动程序 也就是不加任何子命令 只指定--config参数
			NewStart(configFilePath)
		}

	},
}

// 增加一个新的子命令 version
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "print version.",
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Println("v1.0")
		klog.Infoln("v0.0")
	},
}

// 增加一个新的子命令 init 需要指定参数 --config 这里是他的启动方法
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "used for initializing the environment for the first time.",
	Run: func(cmd *cobra.Command, args []string) {
		if checkConfigFile(configFilePath) {
			initEnvironment(configFilePath)
		}
	},
}

// init cobra框架 将所有的都添加到rootCmd这个主命令下
func init() {
	rootCmd.PersistentFlags().StringVar(&configFilePath, "config", "", "config file path.")
	rootCmd.AddCommand(versionCmd)
	// 添加一个命令 init 需要指定参数 --config
	initCmd.PersistentFlags().StringVar(&configFilePath, "config", "", "config file path.")
	rootCmd.AddCommand(initCmd)
}

func checkConfigFile(configFilePath string) bool {
	if configFilePath == "" {
		// fmt.Println("please input --config!")
		klog.Fatalln("please input --config!")
		return false
	}
	// fmt.Println("start!Use config file is :", configFilePath)
	klog.V(2).Info("start!Use config file is :", configFilePath)
	return true
}

// cobra的启动函数
func Start() {

	if err := rootCmd.Execute(); err != nil {
		klog.Fatalln("start error! please check database config!")
	}
}

// go get -u github.com/spf13/cobra
