package spider

import (
	"fmt"
	myappv1 "github.com/coolsun/cloud-app/k8s/crd/pkg/apis/app/v1"
	"github.com/coolsun/cloud-app/model"
	"github.com/coolsun/cloud-app/utils/log"
	"github.com/imroc/req"
	"github.com/pkg/errors"
	"math"
	"strings"
)

//获取每一小类app的所有tag号
func loadAppTags() {
	appLittleCategory := plat.App().GetLittleCategoryName()
	for _, v := range appLittleCategory {
		library, _ := getLibraryAndName(v)
		tags, err := getTags(v)
		if err != nil {
			log.Error(err)
			continue
		}
		for _, t := range tags {
			engine.InsertOne(&model.Tag{
				Library:           library,
				AppLittleCategory: v,
				Version:           t,
			})
		}

	}
}

func getLibraryAndName(littleCategory string) (library string, name string) {
	tmpl := plat.App().GetTemplates()
	for _, v := range tmpl {
		if v.AppLittleCategory == littleCategory {
			arr := strings.Split(v.Pod[myappv1.Main].Image, "/")
			if len(arr) > 1 {
				library = arr[len(arr)-2]
				name = arr[len(arr)-1]
			}
		}
	}
	if library == "" {
		library = "library"
	}
	return
}

func getTags(littleCategory string) (tags []string, err error) {
	library, name := getLibraryAndName(littleCategory)
	tags, err = getAppTags(library, name)
	if err != nil {
		return
	}
	return
}

//从dockerhub获取镜像的tag号
func getAppTags(library string, name string) (tags []string, err error) {
	if library == "mssql" && name == "server" {
		return getMicrosoftSQLServer()
	} else {
		return getCommonAppTags(library, name)
	}
}

func getMicrosoftSQLServer() (tags []string, err error) {
	type Res struct {
		Tags []string `json:"tags"`
	}
	resp, err := req.Get("https://mcr.microsoft.com/v2/mssql/server/tags/list")
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	res := &Res{}
	err = resp.ToJSON(res)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	tags = res.Tags
	return
}

func getCommonAppTags(library string, name string) (tags []string, err error) {
	count, newTags, err := getTagsFromDockerHub(library, name, 1)
	if err != nil {
		return
	}
	tags = append(tags, newTags...)
	pages := int(math.Ceil(float64(count) / float64(100)))
	if pages == 1 {
		return
	}
	for i := 1; i <= pages; i++ {
		_, newTags, _ = getTagsFromDockerHub(library, name, i)
		tags = append(tags, newTags...)
	}
	return
}

func getTagsFromDockerHub(library string, name string, page int) (count int, tags []string, err error) {
	resp, err := req.Get(fmt.Sprintf("https://hub.docker.com/v2/repositories/%v/%v/tags/?ordering=last_updated&page=%v&page_size=100", library, name, page))
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	tagsInfo := &model.TagsInfo{}
	err = resp.ToJSON(tagsInfo)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	count = tagsInfo.Count
	for _, v := range tagsInfo.Results {
		tags = append(tags, v.Name)
	}
	return
}
