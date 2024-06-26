package databases

import "log"

type Mysql struct{}

func (m *Mysql) init() {
	log.Println("start init ...")
}
