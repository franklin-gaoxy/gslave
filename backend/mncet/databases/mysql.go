package databases

import (
	"log"
	"mncet/mncet/tools"
)

type Mysql struct{}

func NewMysql() *Mysql {
	return &Mysql{}
}
func (m *Mysql) Init(config tools.ServerConfig) {
	log.Println("start init ...")
}
