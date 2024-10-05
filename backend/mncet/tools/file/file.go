package file

// 定义一个File 结构体
type File struct {
	UploadFile struct {
		FromNetwork string `json:"fromNetwork"`
		FileSystem  string `json:"fileSystem"`
	} `json:"uploadFile"`
}
