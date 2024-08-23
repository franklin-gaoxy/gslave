package databases

import "mncet/mncet/tools"

type Databases interface {
	/*
		databases interface
	*/

	// init func conn database return conn,err
	Init(config tools.ServerConfig)
}

func NewDatabases(databaseType string) Databases {
	switch databaseType {
	case "mysql":
		return NewMysql()
	case "mongodb":
		return NewMongodb()
	default:
		return nil
	}
}
