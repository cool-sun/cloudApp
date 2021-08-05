package k8s

import (
	"context"
	myappv1 "github.com/coolsun/cloud-app/k8s/crd/pkg/apis/app/v1"
	"github.com/coolsun/cloud-app/k8s/crd/pkg/generated/clientset/versioned"
	appinformers "github.com/coolsun/cloud-app/k8s/crd/pkg/generated/informers/externalversions"
	lmyappv1 "github.com/coolsun/cloud-app/k8s/crd/pkg/generated/listers/app/v1"
	"github.com/coolsun/cloud-app/model"
	vssvv "github.com/coolsun/cloud-app/utils/kubernetes-csi/external-snapshotter/client/clientset/versioned"
	"github.com/coolsun/cloud-app/utils/log"
	vssvvInformer "github.com/kubernetes-csi/external-snapshotter/client/v4/informers/externalversions"
	vssvvListerv1 "github.com/kubernetes-csi/external-snapshotter/client/v4/listers/volumesnapshot/v1"
	"github.com/pkg/errors"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	crdinformers "k8s.io/apiextensions-apiserver/pkg/client/informers/externalversions"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	lappv1 "k8s.io/client-go/listers/apps/v1"
	lhpav1 "k8s.io/client-go/listers/autoscaling/v1"
	lcv1 "k8s.io/client-go/listers/core/v1"
	lnetv1 "k8s.io/client-go/listers/networking/v1"
	storageV1 "k8s.io/client-go/listers/storage/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
	"os"
	"path"
	"time"
)

type labelsKeyName = string

const (
	CloudAppBigCategory    labelsKeyName = "cloudAppBigCategory"
	CloudAppLittleCategory labelsKeyName = "cloudAppLittleCategory"
	CloudAppName           labelsKeyName = "cloudAppName"
)

const (
	defaultPvcSize = "10"
)

var restConfig *rest.Config

type Kube struct {
	crdClient     apiextensionsclientset.Interface
	appClient     versioned.Interface
	vssClient     vssvv.Interface
	kubeClient    kubernetes.Interface
	metricsClient metricsv.Interface
	lister        *lister
	templates     []*myappv1.AppSpec
	cfg           *model.Config
}

type lister struct {
	namespaceLister     lcv1.NamespaceLister
	podLister           lcv1.PodLister
	pvcLister           lcv1.PersistentVolumeClaimLister
	configmapLister     lcv1.ConfigMapLister
	deployLister        lappv1.DeploymentLister
	daemonSetLister     lappv1.DaemonSetLister
	svcLister           lcv1.ServiceLister
	hpaLister           lhpav1.HorizontalPodAutoscalerLister
	ingressLister       lnetv1.IngressLister
	resourceQuotaLister lcv1.ResourceQuotaLister
	stateLister         lappv1.StatefulSetLister
	appLister           lmyappv1.AppLister
	scLister            storageV1.StorageClassLister
	vsscLister          vssvvListerv1.VolumeSnapshotClassLister
	vssLister           vssvvListerv1.VolumeSnapshotLister
}

func kubeClientInit() (kubeClient kubernetes.Interface, err error) {
	kubeClient, err = kubernetes.NewForConfig(restConfig)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	return
}

func appClientInit() (client versioned.Interface, err error) {
	client, err = versioned.NewForConfig(restConfig)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	return
}

func metricsClientInit() (metricsClient metricsv.Interface, err error) {
	metricsClient, err = metricsv.NewForConfig(restConfig)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	return
}

func crdClientInit() (crdClient apiextensionsclientset.Interface, err error) {
	crdClient, err = apiextensionsclientset.NewForConfig(restConfig)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	return
}

func vssClientInit() (vssClient vssvv.Interface, err error) {
	vssClient, err = vssvv.NewForConfig(restConfig)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	return
}

func new(cfg *model.Config, k8sNewDone chan int) (kube *Kube, err error) {
	//先默认从集群内加载k8s配置
	//集群内加载失败的话从环境变量的KubeConfig路径获取配置
	//KubeConfig未指定或者也加载配置失败的话，再从默认位置加载k8s配置
	restConfig, err = inClusterConnect()
	if err != nil {
		if cfg.KubeConfig == "" {
			restConfig, err = outClusterConnect()
		} else {
			restConfig, err = outClusterConnect(cfg.KubeConfig)
		}
		if err != nil {
			err = errors.WithStack(err)
			return
		}
	}
	//多种client依次初始化
	kubeClient, err := kubeClientInit()
	if err != nil {
		return
	}
	appClient, err := appClientInit()
	if err != nil {
		return
	}
	metricsClient, err := metricsClientInit()
	if err != nil {
		return
	}
	crdClient, err := crdClientInit()
	if err != nil {
		return
	}
	vssClient, err := vssClientInit()
	if err != nil {
		return
	}

	tmpl, err := loadTemplate()
	if err != nil {
		return
	}
	err = loadFillInfo()
	if err != nil {
		return
	}
	kube = &Kube{
		crdClient:     crdClient,
		appClient:     appClient,
		vssClient:     vssClient,
		kubeClient:    kubeClient,
		metricsClient: metricsClient,
		templates:     tmpl,
		cfg:           cfg,
	}

	err = kube.createCRD()
	if err != nil {
		return
	}
	go kube.loadLister(k8sNewDone)
	//自定义资源控制器初始化
	go controllerStart(kube)
	return
}

func (k *Kube) loadLister(k8sNewDone chan int) {
	// 初始化 informer factory（为了测试方便这里设置每30s重新 List 一次）
	defaultResync := time.Second * 30
	informerFactory := informers.NewSharedInformerFactory(k.kubeClient, defaultResync)
	appInformerFactory := appinformers.NewSharedInformerFactory(k.appClient, defaultResync)
	crdInformerFactory := crdinformers.NewSharedInformerFactory(k.crdClient, defaultResync)
	vssvvInformerFactory := vssvvInformer.NewSharedInformerFactory(k.vssClient, defaultResync)
	k.lister = &lister{
		namespaceLister:     informerFactory.Core().V1().Namespaces().Lister(),
		podLister:           informerFactory.Core().V1().Pods().Lister(),
		pvcLister:           informerFactory.Core().V1().PersistentVolumeClaims().Lister(),
		configmapLister:     informerFactory.Core().V1().ConfigMaps().Lister(),
		deployLister:        informerFactory.Apps().V1().Deployments().Lister(),
		daemonSetLister:     informerFactory.Apps().V1().DaemonSets().Lister(),
		svcLister:           informerFactory.Core().V1().Services().Lister(),
		hpaLister:           informerFactory.Autoscaling().V1().HorizontalPodAutoscalers().Lister(),
		ingressLister:       informerFactory.Networking().V1().Ingresses().Lister(),
		resourceQuotaLister: informerFactory.Core().V1().ResourceQuotas().Lister(),
		stateLister:         informerFactory.Apps().V1().StatefulSets().Lister(),
		appLister:           appInformerFactory.Cloud().V1().Apps().Lister(),
		scLister:            informerFactory.Storage().V1().StorageClasses().Lister(),
		vsscLister:          vssvvInformerFactory.Snapshot().V1().VolumeSnapshotClasses().Lister(),
		vssLister:           vssvvInformerFactory.Snapshot().V1().VolumeSnapshots().Lister(),
	}
	stopper := make(chan struct{})
	defer close(stopper)
	k.watchDeleteEvent(informerFactory)
	appInformerFactory.Start(stopper)
	crdInformerFactory.Start(stopper)
	informerFactory.Start(stopper)

	appInformerFactory.WaitForCacheSync(stopper)
	crdInformerFactory.WaitForCacheSync(stopper)
	informerFactory.WaitForCacheSync(stopper)
	k8sNewDone <- 1
	<-stopper

}

func (k *Kube) reCreateApp(namespace, name string) {
	app, err := k.lister.appLister.Apps(namespace).Get(name)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			log.Errorf("%+v", err)
		}
		return
	}
	err = k.createApp(app)
	if err != nil {
		if !apierrors.IsAlreadyExists(err) {
			log.Errorf("%+v", err)
		}
		return
	}
}

func inClusterConnect() (config *rest.Config, err error) {
	config, err = rest.InClusterConfig()
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	return
}

func outClusterConnect(paths ...string) (config *rest.Config, err error) {
	var kubeConfigPath = path.Join(os.Getenv("HOME"), "/.kube/config")
	if len(paths) != 0 {
		kubeConfigPath = paths[0]
	}
	config, err = clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	return
}

type ResourceType string

const (
	APP           ResourceType = "app"
	Namespace     ResourceType = "Namespace"
	Pod           ResourceType = "Pod"
	PVC           ResourceType = "PVC"
	ConfigMap     ResourceType = "ConfigMap"
	Deployment    ResourceType = "Deployment"
	Service       ResourceType = "Service"
	HPA           ResourceType = "HPA"
	Ingress       ResourceType = "Ingress"
	ResourceQuota ResourceType = "ResourceQuota"
	StatefulSet   ResourceType = "StatefulSet"
	CRD           ResourceType = "CRD"
	DaemonSet     ResourceType = "DaemonSet"
)

//检查一个k8s资源是否存在,对于不区分命名空间的资源，namespace字段传空字符串就行；
func (k *Kube) CheckResourceExist(t ResourceType, namespace, name string) (o interface{}, e bool) {
	var err error
	switch t {
	case APP:
		o, err = k.lister.appLister.Apps(namespace).Get(name)
	case Namespace:
		o, err = k.lister.namespaceLister.Get(name)
	case Pod:
		o, err = k.lister.podLister.Pods(namespace).Get(name)
	case PVC:
		o, err = k.lister.pvcLister.PersistentVolumeClaims(namespace).Get(name)
	case ConfigMap:
		o, err = k.lister.configmapLister.ConfigMaps(namespace).Get(name)
	case Deployment:
		o, err = k.lister.deployLister.Deployments(namespace).Get(name)
	case Service:
		o, err = k.lister.svcLister.Services(namespace).Get(name)
	case HPA:
		o, err = k.lister.hpaLister.HorizontalPodAutoscalers(namespace).Get(name)
	case Ingress:
		o, err = k.lister.ingressLister.Ingresses(namespace).Get(name)
	case ResourceQuota:
		o, err = k.lister.resourceQuotaLister.ResourceQuotas(namespace).Get(name)
	case StatefulSet:
		o, err = k.lister.stateLister.StatefulSets(namespace).Get(name)
	case CRD:
		o, err = k.crdClient.ApiextensionsV1().CustomResourceDefinitions().Get(context.TODO(), name, metav1.GetOptions{})
	case DaemonSet:
		o, err = k.lister.daemonSetLister.DaemonSets(namespace).Get(name)
	default:
		err = errors.New(model.E10014)
	}
	if err == nil {
		e = true
	}
	return
}

//阻塞直到资源创建成功
func (k *Kube) blockUntilResourceExist(t ResourceType, namespace, name string) (err error) {
	timeout := time.After(time.Second * 10)
	for {
		select {
		case <-timeout:
			err = errors.New(model.E10009)
			return
		default:
			_, exist := k.CheckResourceExist(t, namespace, name)
			if exist {
				return
			}
		}
	}
}

//阻塞直到资源删除成功
func (k *Kube) blockUntilResourceNotExist(t ResourceType, namespace, name string) (err error) {
	timeout := time.After(time.Second * 10)
	for {
		select {
		case <-timeout:
			err = errors.New(model.E10020)
			return
		default:
			_, exist := k.CheckResourceExist(t, namespace, name)
			if !exist {
				return
			}
		}
	}
}
