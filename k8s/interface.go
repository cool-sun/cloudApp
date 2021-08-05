package k8s

import (
	myappv1 "github.com/coolsun/cloud-app/k8s/crd/pkg/apis/app/v1"
	"github.com/coolsun/cloud-app/model"
	metrics_v1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

type App interface {
	Get(namespace, appName string) (*myappv1.App, error)
	Create(app *myappv1.App) error
	Delete(namespace, name string) error
	UpdatePvcDataSource(namespace, name, snapshotName string) error
	GetTemplates() []*myappv1.AppSpec
	GetLittleCategoryName() []string
	GetLittleCategoryFillInfo(appLittleCategory string) LittleCategory
	GetBigCategoryFillInfo(appBigCategory string) BigCategory
	GetMetric(namespace, appName string) ([]metrics_v1beta1.PodMetrics, error)
	GetStatus(namespace, appName string) string
	GetConfig(namespace, appName, configDir string, configFiles []string) (map[string]string, error)
	GetTemplateByLittleCategory(appLittleCategory string) (*myappv1.AppSpec, error)
	GetScList() ([]*model.SC, error)
	GetEndPoint(namespace, appName string) *model.EndPoint

	CreateUser(name string, resourceQuota *model.ResourceQuota) error
	DeleteUser(name string) error

	OpenOrClosePort(namespace, appName string, isOpen bool) error
	DeleteAppPods(namespace, appName string) error

	UpdateConfig(namespace, appName, configMountPath string, data map[string]string) error

	SnapshotCreate(namespace, appName, name string) error
	SnapshotDelete(namespace, name string) error
}

type Helm interface {
	UninstallRelease(namespace, releaseName string) error
	InstallOrUpgradeChart(namespace, repoName, repoURL, releaseName, chartName, version, values string) error
	GetReleaseValues(namespace, releaseName string) (map[string]interface{}, error)
	GetReleaseEndPoint(namespace, releaseName string) (*model.EndPoint, error)
}

type Platform interface {
	App() App
	Helm() Helm
}
