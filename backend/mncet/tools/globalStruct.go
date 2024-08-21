package tools

/*
服务配置
*/
var Version string = "1.0.0"

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
	TaskName  string `yaml:"taskName" bson:"taskName"`
	RecordLog struct {
		File string `yaml:"file" bson:"file"`
	} `yaml:"recordLog" bson:"recordLog"`
	CommandList struct {
		Stages []Stage
	} `yaml:"commandList" bson:"commandList"`
}
type Stage struct {
	Name               string   `yaml:"name" bson:"name"`
	Hosts              []string `yaml:"hosts" bson:"hosts"`
	Command            string   `yaml:"command" bson:"command"`
	ConcurrentMode     string   `yaml:"concurrentMode" bson:"concurrentMode"`
	EncounteredAnError bool     `yaml:"encounteredAnError" bson:"encounteredAnError"`
	UploadFile         struct {
		FromNetwork string `yaml:"fromNetwork" bson:"fromNetwork"`
		FileSystem  string `yaml:"fileSystem" bson:"fileSystem"`
	} `yaml:"uploadFile" bson:"uploadFile"`
}

/*
任务运行信息相关
*/
type TaskInfo struct {
	TaskName string `yaml:"taskName" bson:"taskName"`
	TaskId   string `yaml:"taskId" bson:"taskId"`
	Stage    []struct {
		StageName   string `yaml:"stageName" bson:"stageName"`
		StageResult string `yaml:"stageResult" 	bson:"stageResult"`
		StageStatus string `yaml:"stageStatus" bson:"stageStatus"`
	} `yaml:"stage" bson:"stage"`
}

/*
主机相关
*/
type Hosts struct {
	Hostname string `yaml:"hostname" bson:"hostname"`
	Address  string `yaml:"address" bson:"address"`
	Group    string `yaml:"group" bson:"group"`
	Login    struct {
		Username string `yaml:"username" bson:"username"`
		Password string `yaml:"password" bson:"password"`
		Port     int8   `yaml:"port" bson:"port"`
		SSHKey   string `yaml:"sshKey" bson:"sshKey"`
	} `yaml:"login" bson:"login"`
	HostInfo struct {
		CPU    string `yaml:"cpu" bson:"cpu"`
		Memory string `yaml:"memory" bson:"memory"`
		Disk   []MountDisk
	} `yaml:"hostInfo" bson:"hostInfo"`
	Status string `yaml:"status" bson:"status"`
}
type MountDisk struct {
	MountPoint string `yaml:"mountpoint" bson:"mountpoint"`
	Size       string `yaml:"size" bson:"size"`
}

/*
系统信息相关
*/
type SystemInfo struct {
	SystemStatus string `yaml:"systemStatus" bson:"systemStatus"`
	Task         struct {
		AllTask     int `yaml:"allTask" bson:"allTask"`
		RunningTask int `yaml:"runningTask" bson:"runningTask"`
		FailedTask  int `yaml:"failedTask" bson:"failedTask"`
	} `yaml:"task" bson:"task"`
	Version string `yaml:"version" bson:"version"`
}
