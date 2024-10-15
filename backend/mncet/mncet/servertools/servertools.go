package servertools

import (
	"bytes"
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"k8s.io/klog"
	"mncet/mncet/databases"
	"mncet/mncet/tools"
	"net"
	"reflect"
	"strings"
	"text/template"
)

func FormatYamlContent(TemplateData []byte, ValuesData []byte) (bool, tools.Tasks, error) {
	//
	klog.V(8).Infof("[servertools.go:FormatYamlContent]: function FormatYamlContent start!")

	var err error
	var values map[string]interface{}
	var ExecutionContent tools.Tasks

	err = yaml.Unmarshal(ValuesData, &values)
	if err != nil {
		klog.Errorf("[servertools.go:FormatYamlContent]: Failed to unmarshal values.yaml: %v", err)
		return false, ExecutionContent, err
	}

	// 定义模板函数映射，添加 upper 函数 实现yaml文件中的函数执行
	funcMap := template.FuncMap{
		"upper": strings.ToUpper,
	}

	// 使用 text/template 解析 template.yaml 并应用解析后的数据和函数映射
	tmpl, err := template.New("yamlTemplate").Funcs(funcMap).Parse(string(TemplateData))
	if err != nil {
		klog.Errorf("[servertools.go:FormatYamlContent]: Failed to parse template.yaml: %v", err)
		return false, ExecutionContent, err
	}

	var output bytes.Buffer
	err = tmpl.Execute(&output, values)
	if err != nil {
		klog.Errorf("[servertools.go:FormatYamlContent]: Failed to execute template: %v", err)
		return false, ExecutionContent, err
	}
	klog.V(6).Infof("[servertools.go:FormatYamlContent]: Rendered YAML output: \n%s", output.String())
	klog.V(6).Infof("[servertools.go:FormatYamlContent]: All processing has been completed and we are now converting to YAML format.")
	klog.V(8).Infof("[servertools.go:FormatYamlContent]: {{ ExecutionContent }}:>>%s<<", ExecutionContent)

	err = yaml.Unmarshal(output.Bytes(), &ExecutionContent)
	if err != nil {
		klog.Errorf("[servertools.go:FormatYamlContent]: Failed to unmarshal yaml: %v", err)
		return false, ExecutionContent, err
	}
	klog.V(8).Infof("[servertools.go:FormatYamlContent]: function FormatYamlContent end!\nreturn parameter: \n{{ ExecutionContent }}:>>%s<<", ExecutionContent)
	return true, ExecutionContent, nil
}

func RegisterPlugin() map[string]map[string]tools.Desctibe {
	handlerMap := map[string]map[string]tools.Desctibe{}
	return handlerMap
}

// 检查主机是否存在 不存在返回nil 存在返回对应的主机信息
func CheckHostExist(content *tools.Stage, database databases.Databases) (*[]tools.HostInfo, error) {
	var key string
	var value string
	var result []tools.HostInfo
	for _, host := range content.Hosts {
		klog.V(8).Infof("start check hosts %s does it esist.", host)
		// 检查使用的是IP地址还是主机名
		if IsValidIP(host) {
			klog.V(8).Infof("host %s is IP.", host)
			key = "address"
			value = host
		} else {
			klog.V(8).Infof("host %s is HostName.", host)
			key = "hostname"
			value = host
		}

		// check duplicate
		hosts := database.QueryHosts(key, value)
		if hosts == nil {
			klog.Errorf("host %s is not exist!.", host)
			return &result, errors.New("host not exist")
		} else {
			// 循环返回的list 全部装入result
			for _, h := range *hosts {
				result = append(result, h)
			}
		}
	}
	return &result, nil
}
func IsValidIP(ipStr string) bool {
	// 使用 net.ParseIP 解析 IP 地址
	ip := net.ParseIP(ipStr)
	// 如果解析结果不为 nil，表示这是一个合法的 IP 地址
	return ip != nil
}

// 调用方法
func CallMethodByName(stage interface{}, ser *tools.StageExecutionRecord, methodName string, arg *tools.Stage) error {
	/*
		methodName 用来检查使用那个struct的那个函数 此参数不会传递给执行函数
	*/
	method := reflect.ValueOf(stage).MethodByName(methodName)
	if !method.IsValid() {
		return fmt.Errorf("method %s not found", methodName)
	}
	// 参数转换
	in := []reflect.Value{reflect.ValueOf(ser), reflect.ValueOf(arg)}

	// 调用方法
	out := method.Call(in)

	// 检查返回值中是否包含错误
	if len(out) > 0 {
		// 假设返回的最后一个值是 error
		if err, ok := out[len(out)-1].Interface().(error); ok && err != nil {
			return err
		}
	}

	return nil
}
