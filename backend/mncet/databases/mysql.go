package databases

import "log"

type Mysql struct{}

func NewMysql() *Mysql {
	return &Mysql{}
}
func (m *Mysql) init() {
	log.Println("start init ...")
}
