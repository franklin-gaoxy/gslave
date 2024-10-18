package databases

import (
	"k8s.io/klog"
	"mncet/mncet/tools"
)

type Mysql struct{}

func NewMysql() *Mysql {
	return &Mysql{}
}
func (m *Mysql) Init(config tools.ServerConfig) {
	klog.Infof("start init ...")
}
