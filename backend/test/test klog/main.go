package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

var (
	configFilePath string
	rootCmd        = &cobra.Command{
		Use:   "app",
		Short: "A brief description of your application",
		Long:  `A longer description that spans multiple lines and likely contains examples and usage of using your application.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Using config file:", configFilePath)
			fmt.Println("Remaining args:", args)
		},
	}
)

func init() {
	// 初始化 klog 标志
	klog.InitFlags(nil)

	// 在根命令中添加一个持久标志
	rootCmd.PersistentFlags().StringVar(&configFilePath, "config", "", "config file path.")
}

func handleVParam() {
	// 临时存储 os.Args
	args := os.Args[1:]
	remainingArgs := []string{os.Args[0]}

	for _, arg := range args {
		if strings.HasPrefix(arg, "--v=") {
			// 处理 --v 参数
			vValue := strings.TrimPrefix(arg, "--v=")
			fmt.Printf("Handling --v=%s parameter\n", vValue)

			// 强制设置 klog 的 -v 参数
			if err := flag.Set("v", vValue); err != nil {
				fmt.Printf("Failed to set klog -v flag: %v\n", err)
			}
		} else {
			remainingArgs = append(remainingArgs, arg)
		}
	}

	// 重新设置 os.Args 为剩余参数，不包含 --v 参数
	os.Args = remainingArgs
}

func main() {
	// 在 Cobra 解析之前处理 --v 参数
	handleVParam()

	// 解析 klog 标志 (只解析 klog 自己的标志)
	flag.CommandLine.Parse(nil)

	// 执行 Cobra 命令
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
