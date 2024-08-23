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
	// klog.Infof("klog init: log event %d\n", tools.LogEvent)
	defer klog.Flush()
}
