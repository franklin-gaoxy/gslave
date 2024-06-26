package mncet

import (
	"flag"

	"k8s.io/klog"
)

func InitStart() bool {
	klogInit()

	return true
}

func klogInit() {
	klog.InitFlags(nil)
	flag.Set("V", "2")
	flag.Parse()
	defer klog.Flush()
}
