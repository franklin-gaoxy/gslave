package mncet

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"k8s.io/klog"
)

func InitStart() bool {
	klogInit()

	return true
}

func klogInit() {
	klog.InitFlags(nil)
	// flag.Set("V", "2")
	// flag.Parse()

	parameterProcessing()
	flag.CommandLine.Parse(nil)
	// klog.Infof("klog init: log event %d\n", tools.LogEvent)
	defer klog.Flush()
}

func parameterProcessing() {
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
