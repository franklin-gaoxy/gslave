package databases

type Databases interface {
	/*
		databases interface
	*/

	// init func conn database return conn,err
	init()
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
