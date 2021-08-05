package k8s

import (
	"context"
	myappv1 "github.com/coolsun/cloud-app/k8s/crd/pkg/apis/app/v1"
	"github.com/coolsun/cloud-app/model"
	"github.com/coolsun/cloud-app/utils/log"
	"github.com/mohae/deepcopy"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//自定义控制器创建app
//依次创建对应的deployment或者statefulset之类的控制器，svc,pvc等
//问题，这里为什么不同步创建对应的configmap？主要有如下几个原因：
//1.configmap不是必须的，用户可能直接使用镜像自带的配置；
//2.为了良好的用户体验,即使需要创建对应的configmap，也先获取容器里默认带的配置文件作为模板(获取容器内默认的配置文件就需要等到容器已经运行了)，方便用户填写。
func (k *Kube) createApp(app *myappv1.App) (err error) {
	err = k.createService(app)
	if err != nil {
		return
	}
	if app.Spec.Controller == string(StatefulSet) {
		err = k.createStatefulset(app)
		if err != nil {
			return
		}
		//使用statefulset的控制器时,还要创建headlessService
		err = k.createHeadlessService(app)
		if err != nil {
			return
		}
	} else {
		if app.Spec.Controller == string(DaemonSet) {
			err = k.createDaemonSet(app)
			if err != nil {
				return
			}
		} else if app.Spec.Controller == string(Deployment) {
			err = k.createDeployment(app)
			if err != nil {
				return
			}
		}
		if app.Spec.Pvc != nil {
			err = k.createPVC(app)
			if err != nil {
				return
			}
		}
	}
	return
}

func (k *Kube) getTemplates() (tmpl []*myappv1.AppSpec) {
	tmpl = make([]*myappv1.AppSpec, 0, 0)
	for _, v := range k.templates {
		tmpl = append(tmpl, deepcopy.Copy(v).(*myappv1.AppSpec))
	}
	return
}

//自定义控制器删除app
func (k *Kube) deleteApp(namespace, name string) {
	_ = k.deleteDeployment(namespace, name)
	_ = k.deleteService(namespace, name)
	_ = k.deletePVC(namespace, name)
	_ = k.deleteHpa(namespace, name)
	_ = k.deleteIngress(namespace, name)
	_ = k.deleteConfigmap(namespace, name)
}

//获取app的配置
//首先看有没有相应名称的configmap存在，有的话就返回configmap的配置
//如果没有相应的configmap存在，就根据给定的配置文件路径，在pod中执行exec命令找到默认的配置文件来
func (k *Kube) getConfig(namespace, appName, configDir string, configFiles []string) (m map[string]string, err error) {
	cm, err := k.getConfigmap(namespace, appName)
	if err != nil {
		m, err = k.getDefaultConfigFromPod(namespace, appName, configDir, configFiles)
		return
	}
	m = cm.Data
	return
}

//更新app配置，
//大致的思路是创建相应的configmap,然后将configmap以文件的形式挂载到容器内
//首次创建对应configmap的时候需要更新对应的deployment,也可能是更新别的类型的控制器
func (k *Kube) updateConfig(namespace, appName, configMountPath string, data map[string]string) (err error) {
	app, appExist := k.CheckResourceExist(APP, namespace, appName)
	if !appExist {
		err = errors.New(model.E10012)
		return
	}
	appapp := app.(*myappv1.App)
	_, cmExist := k.CheckResourceExist(ConfigMap, namespace, appName)
	if cmExist {
		err = k.updateConfigmap(appapp, data)
		if err != nil {
			return
		}
	} else {
		err = k.createConfigmap(appapp, data)
		if err != nil {
			return
		}
	}
	return k.updateDeploymentMountConfigmap(namespace, appName, configMountPath)
}

//更新app的状态
func (k *Kube) updateAppStatus(namespace, name string, condition myappv1.Condition) {
	app, err := k.lister.appLister.Apps(namespace).Get(name)
	if err != nil {
		log.Error(err)
		return
	}
	if app.Status.Condition == condition {
		return
	}
	app.Status.Condition = condition
	_, err = k.appClient.CloudV1().Apps(namespace).Update(context.TODO(), app, metav1.UpdateOptions{})
	if err != nil {
		log.Error(err)
		return
	}
}

func (k *Kube) getLittleCategoryName() (littleCategoryArr []string) {
	littleCategoryArr = make([]string, 0, 0)
	for _, v := range k.templates {
		littleCategoryArr = append(littleCategoryArr, v.AppLittleCategory)
	}
	return
}

func (k *Kube) getStatus(namespace, name string) (status string) {
	app, err := k.lister.appLister.Apps(namespace).Get(name)
	if err != nil {
		log.Error(err)
		return
	}
	status = string(app.Status.Condition)
	return
}

//todo 没有填挂载路径的时候，不创建pvc
//todo 加载模板的时候必须检查 pod 下面的main容器是否存在,以及main容器的image地址是否合法
