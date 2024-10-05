package tools

/*
服务配置
*/
var Version string = "1.0.0"

type ServerConfig struct {
	Port     int16 `yaml:"port"`
	Database struct {
		DataBaseType string `yaml:"databaseType"`
		ConnPath     string `yaml:"connPath"`
		//Type         string     `yaml:"type"`
		Path        string     `yaml:"path"`
		Host        string     `yaml:"host"`
		Port        int16      `yaml:"port"`
		AuthSource  string     `yaml:"authSource"`
		AuthType    string     `yaml:"authType"`
		Description UserConfig `yaml:"description"`
		BaseName    string     `yaml:"basename"`
	}
	Login struct {
		User UserConfig
	} `yaml:"login"`
}

type UserConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
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

// 定义了接口来匹配不同模式下的describe的内容
type Desctibe interface{}

type Stage struct {
	Name  string   `yaml:"name" bson:"name"`
	Hosts []string `yaml:"hosts" bson:"hosts"`
	Group string   `yaml:"group" bson:"group"`
	Mode  string   `yaml:"mode" bson:"mode"`
	Type  string   `yaml:"type" bson:"type"`
	// 该字段根据不同的mode和type来匹配不同的值
	Describe Desctibe `yaml:"describe" bson:"describe"`
}

/*
任务运行信息相关
*/
type TaskInfo struct {
	TaskName string `yaml:"taskName" bson:"taskName"`
	TaskId   string `yaml:"taskId" bson:"taskId"`
	Stage    []struct {
		StageName   string `yaml:"stageName" bson:"stageName"`
		StageResult string `yaml:"stageResult" bson:"stageResult"`
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
		Port     int16  `yaml:"port" bson:"port"`
		SSHKey   string `yaml:"sshKey" bson:"sshKey"`
	} `yaml:"login" bson:"login"`
	HostInfo struct {
		CPU       string `yaml:"cpu" bson:"cpu"`
		Memory    string `yaml:"memory" bson:"memory"`
		Disk      []MountDisk
		TotalSize float64 `yaml:"totalSize" bson:"totalSize"`
	} `yaml:"hostInfo" bson:"hostInfo"`
	Status   string `yaml:"status" bson:"status"`
	Describe string `yaml:"describe" bson:"describe"`
}
type MountDisk struct {
	Device     string   `yaml:"device" bson:"device"`
	Name       string   `yaml:"name" bson:"name"`
	MountPoint []string `yaml:"mountpoints" bson:"mountpoints"`
	Size       int      `yaml:"size" bson:"size"`
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
