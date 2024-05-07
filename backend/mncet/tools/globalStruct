package tools

/*
服务配置
*/

type Config struct {
	Port     int8 `yaml:"port"`
	Database struct {
		Path     string
		Port     int8
		User     UserConfig
		Basename string
	}
	login struct {
		User UserConfig
	}
}

type UserConfig struct {
	Username string
	Password string
}

/*
任务相关
*/
type Tasks struct {
	TaskName  string `yaml:"taskName"`
	RecordLog struct {
		File string `yaml:"file"`
	} `yaml:"recordLog"`
	CommandList struct {
		Stages []Stage
	} `yaml:"commandList"`
}
type Stage struct {
	Name               string   `yaml:"name"`
	Hosts              []string `yaml:"hosts"`
	Command            string   `yaml:"command"`
	ConcurrentMode     string   `yaml:"concurrentMode"`
	EncounteredAnError bool     `yaml:"encounteredAnError"`
	UploadFile         struct {
		FromNetwork string `yaml:"fromNetwork"`
		FileSystem  string `yaml:"fileSystem"`
	} `yaml:"uploadFile"`
}
