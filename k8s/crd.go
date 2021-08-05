package k8s

import (
	"context"
	"fmt"
	myappv1 "github.com/coolsun/cloud-app/k8s/crd/pkg/apis/app/v1"
	"github.com/coolsun/cloud-app/model"
	"github.com/coolsun/cloud-app/utils"
	"github.com/pkg/errors"
	apiextensionv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
)

const (
	Kind         string = "App"
	Plural       string = "apps"
	Singular     string = "app"
	GroupVersion string = myappv1.Version
	CRDName      string = Plural + "." + myappv1.GroupName
	ShortName    string = "app"
)

const (
	StrategyRollingUpdate int32 = 0
	StrategyRecreate      int32 = 1
)

var (
	ControllerDeployment  string = `"` + string(Deployment) + `"`
	ControllerStatefulSet string = `"` + string(StatefulSet) + `"`
	ControllerDaemonSet   string = `"` + string(DaemonSet) + `"`
)

var strategyArr = []int32{StrategyRollingUpdate, StrategyRecreate}
var controllerArr = []string{ControllerDeployment, ControllerStatefulSet, ControllerDaemonSet}

func GetEnum(op int32) (j []apiextensionv1.JSON) {
	j = make([]apiextensionv1.JSON, 0, 0)
	var interArr []interface{}
	if op == 1 {
		interArr = utils.ToSlice(strategyArr)
	} else if op == 2 {
		interArr = utils.ToSlice(controllerArr)
	}
	for _, v := range interArr {
		j = append(j, apiextensionv1.JSON{
			Raw: []byte(fmt.Sprintf("%v", v)),
		})
	}
	return
}

//创建crd
func (k *Kube) createCRD() (err error) {
	containerProp := apiextensionv1.JSONSchemaProps{
		Type: "object",
		Properties: map[string]apiextensionv1.JSONSchemaProps{
			"image": {
				Type: "string",
			},
			"tag": {
				Type: "string",
			},
			"args": {
				Type: "array",
				Items: &apiextensionv1.JSONSchemaPropsOrArray{
					Schema: &apiextensionv1.JSONSchemaProps{Type: "string"},
				},
			},
			"command": {
				Type: "array",
				Items: &apiextensionv1.JSONSchemaPropsOrArray{
					Schema: &apiextensionv1.JSONSchemaProps{Type: "string"},
				},
			},
			"cpu": {
				Type:    "string",
				Pattern: model.PositiveInteger,
			},
			"mem": {
				Type:    "string",
				Pattern: model.PositiveInteger,
			},
			"env": {
				Type: "array",
				Items: &apiextensionv1.JSONSchemaPropsOrArray{
					Schema: &apiextensionv1.JSONSchemaProps{Type: "string"},
				},
			},
			"ports": {
				Type: "array",
				Items: &apiextensionv1.JSONSchemaPropsOrArray{
					Schema: &apiextensionv1.JSONSchemaProps{Type: "string"},
				},
			},
			"volume_mount": {
				Type: "array",
				Items: &apiextensionv1.JSONSchemaPropsOrArray{
					Schema: &apiextensionv1.JSONSchemaProps{Type: "string", Pattern: model.FolderPath},
				},
			},
		},
	}
	schema := &apiextensionv1.CustomResourceValidation{
		OpenAPIV3Schema: &apiextensionv1.JSONSchemaProps{
			Type: "object",
			Properties: map[string]apiextensionv1.JSONSchemaProps{
				"spec": {
					Required: []string{"replicas", "strategy", "controller", "pod"},
					Type:     "object",
					Properties: map[string]apiextensionv1.JSONSchemaProps{
						"app_big_category": {
							Type: "string",
						},
						"app_little_category": {
							Type: "string",
						},
						"replicas": {
							Type: "integer",
						},
						"strategy": {
							Type: "integer",
							Enum: GetEnum(1),
						},
						"controller": {
							Type: "string",
							Enum: GetEnum(2),
						},
						"pod": {
							Type: "object",
							Properties: map[string]apiextensionv1.JSONSchemaProps{
								"main":    containerProp,
								"init":    containerProp,
								"sidecar": containerProp,
							},
						},
						"pvc": {
							Type: "object",
							Properties: map[string]apiextensionv1.JSONSchemaProps{
								"size": {
									Type:    "string",
									Pattern: model.NilOrPositiveInteger,
								},
								"sc": {
									Type: "string",
								},
							},
							Required: []string{"size", "sc"},
						},
						"internet_access": {
							Type: "boolean",
						},
						"config": {
							Type: "object",
							Properties: map[string]apiextensionv1.JSONSchemaProps{
								"path": {
									Type: "string",
								},
								"file": {
									Type: "array",
									Items: &apiextensionv1.JSONSchemaPropsOrArray{
										Schema: &apiextensionv1.JSONSchemaProps{Type: "string"},
									},
								},
							},
						},
					},
				},
				"status": {
					Type: "object",
					Properties: map[string]apiextensionv1.JSONSchemaProps{
						"condition": {
							Type: "string",
						},
					},
				},
			},
		},
	}
	crd := &apiextensionv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{"api-approved.kubernetes.io": "https://github.com/kubernetes/enhancements/pull/1111"},
			Name:        CRDName,
		},
		Spec: apiextensionv1.CustomResourceDefinitionSpec{
			Group: myappv1.GroupName,
			Names: apiextensionv1.CustomResourceDefinitionNames{
				Plural:     Plural,
				Singular:   Singular,
				ShortNames: []string{ShortName},
				Kind:       reflect.TypeOf(myappv1.App{}).Name(),
			},
			Scope: apiextensionv1.NamespaceScoped,
			Versions: []apiextensionv1.CustomResourceDefinitionVersion{
				{
					Name:    myappv1.Version,
					Served:  true,
					Storage: true,
					Schema:  schema,
				},
			},
		},
	}
	resource, exist := k.CheckResourceExist(CRD, "", CRDName)
	if exist {
		oldCrd := resource.(*apiextensionv1.CustomResourceDefinition)
		oldCrd.Spec = crd.Spec
		_, err = k.crdClient.ApiextensionsV1().CustomResourceDefinitions().Update(context.TODO(), oldCrd, metav1.UpdateOptions{})
		if err != nil {
			err = errors.WithStack(err)
			return
		}
	} else {
		_, err = k.crdClient.ApiextensionsV1().CustomResourceDefinitions().Create(context.TODO(), crd, metav1.CreateOptions{})
		if err != nil {
			err = errors.WithStack(err)
			return
		}
	}
	return k.blockUntilResourceExist(CRD, "", CRDName)
}
