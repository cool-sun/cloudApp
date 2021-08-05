package model

type Regexp = string

const (
	//版本号(version)格式必须为X.Y.Z
	Version Regexp = "^\\d+(?:\\.\\d+){2}$"
	//linux文件夹路径
	FolderPath Regexp = "^\\/(\\w+\\/?)+$"
	//非零的正整数
	PositiveInteger Regexp = "^\\+?[1-9][0-9]*$"
	//空或者非零的正整数
	NilOrPositiveInteger Regexp = "^([0-9]*[1-9][0-9]*)?$"
	//k8s中各资源名称合法性校验
	KubeResourceName Regexp = "^[a-z0-9-]+$"
)
