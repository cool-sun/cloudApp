package k8s

import (
	"fmt"
	myappv1 "github.com/coolsun/cloud-app/k8s/crd/pkg/apis/app/v1"
	"github.com/coolsun/cloud-app/model"
	"github.com/coolsun/cloud-app/utils"
	"github.com/coolsun/cloud-app/utils/log"
	"github.com/mohae/deepcopy"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"path"
	"regexp"
	"strings"
)

const templateDirName = "./template"

func (k *Kube) getTemplateByLittleCategory(appLittleCategory string) (tmpl *myappv1.AppSpec, err error) {
	for _, v := range k.templates {
		if v.AppLittleCategory == appLittleCategory {
			tmpl = deepcopy.Copy(v).(*myappv1.AppSpec)
			return
		}
	}
	err = errors.New(model.E10010)
	return
}

func loadTemplate() (template []*myappv1.AppSpec, err error) {
	if !utils.Exists(templateDirName) {
		return
	}
	infos, err := ioutil.ReadDir(templateDirName)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	template = make([]*myappv1.AppSpec, 0)
	for _, info := range infos {
		if !info.IsDir() {
			continue
		}
		if err := validName(info.Name()); err != nil {
			log.Errorf("%+v", err)
			continue
		}
		template = append(template, readAppLittleType(path.Join(info.Name()))...)
	}

	return
}

func validName(name string) (err error) {
	matched, err := regexp.MatchString(model.KubeResourceName, name)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	if !matched {
		err = errors.New(fmt.Sprintf("%v不符合k8s资源名称要求", name))
		return
	}
	return
}

func readAppLittleType(bigCategoryDir string) (t []*myappv1.AppSpec) {
	t = make([]*myappv1.AppSpec, 0)
	bigCategoryFullDir := path.Join(templateDirName, bigCategoryDir)
	infos, err := ioutil.ReadDir(bigCategoryFullDir)
	if err != nil {
		log.Errorf("%+v", err)
		return
	}
	for _, info := range infos {
		suffix := ".yaml"
		if info.IsDir() || !strings.Contains(info.Name(), suffix) {
			continue
		}
		if err := validName(strings.ReplaceAll(info.Name(), ".yaml", "")); err != nil {
			log.Errorf("%+v", err)
			continue
		}
		fileByte, err := ioutil.ReadFile(path.Join(bigCategoryFullDir, info.Name()))
		if err != nil {
			log.Errorf("%+v", err)
			continue
		}
		var tmpl = &myappv1.AppSpec{}
		err = yaml.Unmarshal(fileByte, &tmpl)
		if err != nil {
			log.Errorf("%+v", err)
			continue
		}
		tmpl.AppBigCategory = bigCategoryDir
		tmpl.AppLittleCategory = strings.Replace(info.Name(), suffix, "", 1)
		if err := utils.Translate(utils.Validate().Struct(tmpl)); err != nil {
			log.Errorf("%+v", err)
			continue
		}
		t = append(t, tmpl)

	}
	return
}
