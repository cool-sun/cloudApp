package k8s

import (
	"context"
	"fmt"
	myappv1 "github.com/coolsun/cloud-app/k8s/crd/pkg/apis/app/v1"
	"github.com/coolsun/cloud-app/model"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metrics_v1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

var k *Kube

type AppClient struct {
}

type HelmClient struct {
}

func (a *AppClient) UpdatePvcDataSource(namespace, name, snapshotName string) (err error) {
	oldPvc, err := k.lister.pvcLister.PersistentVolumeClaims(namespace).Get(name)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	newPvcName, err := k.createPvcFromSnapshot(oldPvc, namespace, snapshotName)
	if err != nil {
		return
	}

	return k.updateDeploymentVolumes(namespace, name, newPvcName)
}

func (a *AppClient) GetEndPoint(namespace, name string) (endPoint *model.EndPoint) {
	endPoint = &model.EndPoint{
		In:  make([]string, 0),
		Out: make([]string, 0),
	}
	svc, _ := k.lister.svcLister.Services(namespace).Get(name)
	for _, v := range svc.Spec.Ports {
		endPoint.In = append(endPoint.In, fmt.Sprintf("%v.%v.svc:%v", name, namespace, v.Port))
	}
	if svc.Spec.Type == corev1.ServiceTypeNodePort {
		for _, v := range svc.Spec.Ports {
			endPoint.Out = append(endPoint.Out, fmt.Sprintf("%v:%v", k.cfg.MasterIp, v.NodePort))
		}
	}
	return
}
func (h *HelmClient) UninstallRelease(namespace, releaseName string) error {
	return k.uninstallRelease(namespace, releaseName)
}

func (h *HelmClient) InstallOrUpgradeChart(namespace, repoName, repoURL, releaseName, chartName, version, values string) error {
	return k.installOrUpgradeChart(namespace, repoName, repoURL, releaseName, chartName, version, values)
}

func (h *HelmClient) GetReleaseValues(namespace, releaseName string) (map[string]interface{}, error) {
	return k.getReleaseValues(namespace, releaseName)
}
func (h *HelmClient) GetReleaseEndPoint(namespace, releaseName string) (*model.EndPoint, error) {
	return k.getReleaseEndPoint(namespace, releaseName)
}
func (a *AppClient) SnapshotDelete(namespace, name string) (err error) {
	return k.snapshotDelete(namespace, name)
}
func (a *AppClient) SnapshotCreate(namespace, appName, name string) (err error) {
	return k.snapshotCreate(namespace, appName, name)
}
func (a *AppClient) Create(app *myappv1.App) (err error) {
	_, exist := k.CheckResourceExist(APP, app.Namespace, app.Name)
	if exist {
		_, err = k.appClient.CloudV1().Apps(app.Namespace).Update(context.TODO(), app, metav1.UpdateOptions{})
	} else {
		_, err = k.appClient.CloudV1().Apps(app.Namespace).Create(context.TODO(), app, metav1.CreateOptions{})
	}
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	return
}

func (a *AppClient) Delete(namespace, name string) (err error) {
	err1 := k.appClient.CloudV1().Apps(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err1 != nil && !k8s_errors.IsNotFound(err1) {
		err = errors.WithStack(err)
		return
	}
	return

}

func (a *AppClient) Get(namespace, appName string) (app *myappv1.App, err error) {
	app, err = k.appClient.CloudV1().Apps(namespace).Get(context.TODO(), appName, metav1.GetOptions{})
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	return
}

func (a *AppClient) GetTemplates() (tmpl []*myappv1.AppSpec) {
	return k.getTemplates()
}

func (a *AppClient) GetLittleCategoryName() (littleCategoryArr []string) {
	return k.getLittleCategoryName()
}

func (a *AppClient) GetLittleCategoryFillInfo(appLittleCategory string) (littleCategory LittleCategory) {
	return k.getLittleCategoryFillInfo(appLittleCategory)
}

func (a *AppClient) GetBigCategoryFillInfo(appBigCategory string) (bigCategory BigCategory) {
	return k.getBigCategoryFillInfo(appBigCategory)
}

func (a *AppClient) GetMetric(namespace, appName string) (podMetrics []metrics_v1beta1.PodMetrics, err error) {
	return k.getMetric(namespace, appName)
}

func (a *AppClient) GetStatus(namespace, name string) (status string) {
	return k.getStatus(namespace, name)
}

//开启互联网访问入口
func (a *AppClient) OpenOrClosePort(namespace, appName string, isOpen bool) (err error) {
	return k.revertServiceType(namespace, appName, isOpen)
}

//重启app
func (a *AppClient) DeleteAppPods(namespace, appName string) (err error) {
	return k.deleteAppPods(namespace, appName)
}

func (a *AppClient) GetConfig(namespace, appName, configDir string, configFiles []string) (config map[string]string, err error) {
	return k.getConfig(namespace, appName, configDir, configFiles)
}

func (a *AppClient) UpdateConfig(namespace, appName, configMountPath string, data map[string]string) (err error) {
	return k.updateConfig(namespace, appName, configMountPath, data)
}

func (a *AppClient) GetTemplateByLittleCategory(appLittleCategory string) (appSpec *myappv1.AppSpec, err error) {
	return k.getTemplateByLittleCategory(appLittleCategory)
}

func (a *AppClient) DeleteUser(name string) (err error) {
	err = k.kubeClient.CoreV1().Namespaces().Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	return
}

func (a *AppClient) CreateUser(name string, resourceQuota *model.ResourceQuota) (err error) {
	ns, err := k.createNamespace(name)
	if err != nil {
		return
	}
	//状态检测，只有命名空间创建好了，再创建命名空间下别的资源才不会报错
	err = k.blockUntilResourceExist(Namespace, "", ns.Name)
	if err != nil {
		return
	}

	if resourceQuota != nil {
		return k.createResourceQuota(name, resourceQuota.CPU, resourceQuota.Mem)
	} else {
		return k.deleteResourceQuota(name)
	}
}

func (a *AppClient) GetScList() (list []*model.SC, err error) {
	list = make([]*model.SC, 0)
	scList, err := k.GetScList()
	if err != nil {
		return
	}
	for _, v := range scList {
		obj := &model.SC{
			Name: v.Name,
		}
		if v.AllowVolumeExpansion != nil {
			obj.AllowVolumeExpansion = *v.AllowVolumeExpansion
		}
		list = append(list, obj)
	}
	return
}

func New(cfg *model.Config, k8sNewDone chan int) (plat Platform, err error) {
	k, err = new(cfg, k8sNewDone)
	if err != nil {
		return
	}
	plat = &PlatformClient{}
	return
}

type PlatformClient struct {
}

func (p PlatformClient) App() App {
	return &AppClient{}
}

func (p PlatformClient) Helm() Helm {
	return &HelmClient{}
}
