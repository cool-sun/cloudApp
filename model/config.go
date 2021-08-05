package model

type Config struct {
	//基本配置
	ListenPort string `json:"listen_port" env:"listen_port" required:"true" envDefault:"8090"`

	//mysql相关的配置
	MysqlUser     string `json:"mysql_user" env:"mysql_user" required:"true" envDefault:"root"`
	MysqlPassword string `json:"mysql_password" env:"mysql_password" required:"true" envDefault:"cloud-app123,."`
	MysqlHost     string `json:"mysql_host" env:"mysql_host" required:"true" envDefault:"localhost"`
	MysqlPort     int    `json:"mysql_port" env:"mysql_port" required:"true" envDefault:"3306"`
	//系统管理员登录密码
	AdminPassword string `json:"admin_password" env:"admin_password" required:"true" envDefault:"123456"`

	KubeConfig string `json:"kube_config" env:"kube_config"` //连接到k8s集群的配置文件地址
	MasterIp   string `json:"master_ip" env:"master_ip"`     //主节点的ip

	IsDebug bool `json:"is_debug"` //是不是调试模式

}
