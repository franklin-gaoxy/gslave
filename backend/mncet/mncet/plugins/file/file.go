package file

import (
	"k8s.io/klog"
	"mncet/mncet/mncet/servertools"
	"mncet/mncet/tools"
)

type File struct {
}

func (f *File) Details() {
	klog.Infof("mode: file.")
}

func (f *File) CallMethodByType(ser *tools.StageExecutionRecord, typeName string, arg *tools.Stage) error {
	return servertools.CallMethodByName(f, ser, typeName, arg)
}

func (f File) Local(ser *tools.StageExecutionRecord, args *tools.Stage) error {
	return nil
}
