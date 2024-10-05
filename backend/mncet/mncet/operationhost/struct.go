package operationhost

type DiskInfo struct {
	BlockDevices []BlockDevice `json:"blockdevices"`
}

type BlockDevice struct {
	Name       string        `json:"name"`
	Size       int           `json:"size"`
	Mountpoint []string      `json:"mountpoints"`
	Children   []BlockDevice `json:"children,omitempty"`
}
