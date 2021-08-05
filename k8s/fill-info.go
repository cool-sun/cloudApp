package k8s

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

type BigCategory struct {
	ShowName string                     `json:"show_name" yaml:"show_name"`
	Items    map[string]*LittleCategory `json:"items" yaml:"items"`
}
type LittleCategory struct {
	Summary  string `json:"summary" yaml:"summary"`
	ShowName string `json:"show_name" yaml:"show_name"`
	Icon     string `json:"icon" yaml:"icon"`
}

var infos = make(map[string]*BigCategory)

func loadFillInfo() (err error) {
	infoByte, err := ioutil.ReadFile(templateDirName + "/fill-info.yaml")
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	err = yaml.Unmarshal(infoByte, &infos)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	return
}

func (k *Kube) getLittleCategoryFillInfo(appLittleCategory string) (littleCategory LittleCategory) {
	for _, v1 := range infos {
		for k2, v2 := range v1.Items {
			if k2 == appLittleCategory {
				littleCategory = *v2
			}
		}
	}
	return
}

func (k *Kube) getBigCategoryFillInfo(appBigCategory string) (bigCategory BigCategory) {
	for k, v := range infos {
		if k == appBigCategory {
			bigCategory = *v
		}
	}
	return
}
