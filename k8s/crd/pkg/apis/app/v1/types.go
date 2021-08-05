package v1

import (
	"fmt"
	"github.com/coolsun/cloud-app/utils/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"strconv"
	"strings"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type App struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AppSpec   `json:"spec"`
	Status AppStatus `json:"status,omitempty"`
}

type AppStatus struct {
	Condition Condition `json:"condition"`
}

//Condition为 空字符串""表示不可用(比如还没创建好)，1表示停止(关机)，2表示部分可用，3表示完全可用
type Condition string

const (
	ConditionWaitReady Condition = ""
	ConditionSleep     Condition = "1"
	ConditionPartReady Condition = "2"
	ConditionAllReady  Condition = "3"
)

type ContainerType string

const (
	Init    ContainerType = "init"
	Sidecar ContainerType = "sidecar"
	Main    ContainerType = "main"
)

type AppSpec struct {
	AppBigCategory    string                       `yaml:"app_big_category" json:"app_big_category"`
	AppLittleCategory string                       `yaml:"app_little_category" json:"app_little_category"`
	Strategy          int32                        `yaml:"strategy" json:"strategy"`     //升级方式，默认为0表示滚动升级，1表示recreate
	Replicas          int32                        `yaml:"replicas" json:"replicas"`     //副本数
	Controller        string                       `yaml:"controller" json:"controller"` //控制器类型
	Pod               map[ContainerType]*Container `yaml:"pod" json:"pod"`
	Pvc               *Pvc                         `yaml:"pvc" json:"pvc"`
	Config            ConfigInfo                   `yaml:"config" json:"config,omitempty"`
	InternetAccess    bool                         `yaml:"internet_access" json:"internet_access"`
}

type Pvc struct {
	Size string `json:"size" yaml:"size"` //容量大小
	SC   string `json:"sc" yaml:"sc"`     //存储类
}
type ConfigInfo struct {
	Path string   `yaml:"path" json:"path"`
	File []string `yaml:"file" json:"file"`
}

type EnvVar string

func (e *EnvVar) GetName() string {
	return strings.Split(string(*e), ":")[0]
}

func (e *EnvVar) SetValue(value string) {
	*e = EnvVar(fmt.Sprintf("%v:%v", e.GetName(), value))
}

func (e *EnvVar) GetValue() string {
	arr := strings.Split(string(*e), ":")
	if len(arr) == 1 {
		return ""
	}
	return arr[1]
}

type Port string

func (p *Port) GetPort() (port int) {
	s := strings.Split(string(*p), ":")[0]
	port, err := strconv.Atoi(s)
	if err != nil {
		log.Errorf("%+v", err)
		return
	}
	return
}
func (p *Port) GetProtocol() (protocol corev1.Protocol) {
	arr := strings.Split(string(*p), ":")
	if len(arr) == 1 {
		protocol = corev1.ProtocolTCP
		return
	}
	pro := strings.ToLower(strings.TrimSpace(arr[1]))
	if pro == "udp" {
		protocol = corev1.ProtocolUDP
	} else if pro == "sctp" {
		protocol = corev1.ProtocolSCTP
	} else {
		protocol = corev1.ProtocolTCP
	}
	return
}

type Container struct {
	Image       string    `yaml:"image" json:"image"`
	Tag         string    `yaml:"tag" json:"tag"`
	CPU         string    `yaml:"cpu" json:"cpu"`
	Mem         string    `yaml:"mem" json:"mem"`
	Env         []*EnvVar `yaml:"env" json:"env"`
	Ports       []*Port   `yaml:"ports" json:"ports"`
	VolumeMount []string  `yaml:"volume_mount" json:"volume_mount"`
	Command     []string  `yaml:"command" json:"command"`
	Args        []string  `yaml:"args" json:"args"`
}

func (a *AppSpec) GetObjectKind() schema.ObjectKind {
	panic("implement me")
}

func (a *AppSpec) DeepCopyObject() runtime.Object {
	panic("implement me")
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type AppList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []App `json:"items"`
}

func init() {
	SchemeBuilder.Register(
		&App{},
		&AppList{},
	)
}
