package command

type Command struct {
	Command            string `yaml:"command" bson:"command"`
	HostConcurrentMode string `yaml:"hostConcurrentMode" bson:"hostConcurrentMode"`
	stepMode           string `yaml:"stepMode" bson:"stepMode"`
	EncounteredAnError bool   `yaml:"encounteredAnError" bson:"encounteredAnError"`
}
