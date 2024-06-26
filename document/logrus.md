# logrus

## 安装

```shell
go get github.com/sirupsen/logrus
```

## 使用

```go
package main

import (
	log "github.com/sirupsen/logrus"
	"os"
)

func init() {
	// 设置日志格式为文本格式
	log.SetFormatter(&log.TextFormatter{
		// 是否显示完整时间戳
		FullTimestamp: true,
		// 时间戳格式
		TimestampFormat: "2006-01-02 15:04:05",
		// 是否显示日志级别
		DisableLevelTruncation: true,
	})
	// 设置日志级别为Debug 如果更改为InfoLevel,那么将只输出info及更高级别的日志
	log.SetLevel(log.DebugLevel)
	// 将日志输出到标准输出
	log.SetOutput(os.Stdout)
}

func main() {
	log.WithFields(log.Fields{
		"animal": "walrus",
		"size":   10,
	}).Info("A group of walrus emerges from the ocean")
}
```

### Withfields

> 这将允许你附加额外的变量到日志中

```go
package main

import (
    "github.com/sirupsen/logrus"
)

func main() {
    // 创建日志记录器
    log := logrus.New()

    // 附加额外的字段到日志条目中
    log.WithFields(logrus.Fields{
        "animal": "walrus",
        "size":   10,
    }).Info("A group of walrus emerges from the ocean")
}
```

### 不同级别的日志

```go
package main

import (
    "github.com/sirupsen/logrus"
)

func main() {
    // 创建日志记录器
    log := logrus.New()

    // 记录不同级别的日志消息
    log.Debug("This is a debug message")
    log.Info("This is an info message") // 等价 Print Printf Println
    log.Warn("This is a warning message") // Warning 等价
    log.Error("This is an error message")
    log.Fatal("This is a fatal message") // 程序会退出
    log.Panic("This is a panic message") // 程序会 panic
}
```

> - `Print`：记录普通信息级别的日志消息，不带任何格式化。
> - `Printf`：记录格式化后的普通信息级别的日志消息，类似于 `fmt.Printf`。
> - `Println`：记录普通信息级别的日志消息，并在结尾添加换行符。

## 日志保存和压缩

```shell
go get gopkg.in/natefinch/lumberjack.v2
```

```go
package main

import (
    "github.com/sirupsen/logrus"
    "gopkg.in/natefinch/lumberjack.v2"
)

func main() {
    // 创建一个新的 lumberjack.Logger 对象，用于写入日志文件
    logFile := &lumberjack.Logger{
        Filename:   "app.log", // 日志文件的路径和文件名
        MaxSize:    100,       // 每个日志文件的最大尺寸，单位为 MB
        MaxBackups: 3,         // 保留的旧日志文件的最大数量
        MaxAge:     28,        // 保留的旧日志文件的最大保存天数
        Compress:   true,      // 是否压缩旧的日志文件
    }

    // 创建日志记录器
    log := logrus.New()

    // 将日志写入到文件
    log.SetOutput(logFile)

    // 记录日志消息
    log.Info("This is an info message")
}
```

