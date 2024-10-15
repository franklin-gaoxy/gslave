package plugins

import (
	"mncet/mncet/mncet/plugins/command"
	"mncet/mncet/mncet/plugins/file"
	"mncet/mncet/tools"
)

func CreatePlugin() map[string]tools.Desctibe {
	var stages = map[string]tools.Desctibe{
		"Command": &command.Command{},
		"File":    &file.File{},
	}
	return stages
}
