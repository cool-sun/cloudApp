package k8s

import (
	myappv1 "github.com/coolsun/cloud-app/k8s/crd/pkg/apis/app/v1"
	"github.com/coolsun/cloud-app/utils/log"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

//监听资源事件做出相应的操作
//监听删除事件 实现类似deployment管理的pod被删除后还能自动创建回来的功能
//监听更新事件 用来更新自定义资源app的status
func (k *Kube) watchDeleteEvent(informerFactory informers.SharedInformerFactory) {
	informerFactory.Apps().V1().Deployments().Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		DeleteFunc: k.dealDeleteEvent,
		UpdateFunc: k.dealUpdateEvent,
	})
	informerFactory.Apps().V1().StatefulSets().Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		DeleteFunc: k.dealDeleteEvent,
	})
	informerFactory.Core().V1().Services().Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		DeleteFunc: k.dealDeleteEvent,
	})
	informerFactory.Core().V1().PersistentVolumeClaims().Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		DeleteFunc: k.dealDeleteEvent,
	})
}

//处理删除事件
func (k *Kube) dealDeleteEvent(i interface{}) {
	switch t := i.(type) {
	case *appv1.Deployment:
		k.reCreateApp(t.Namespace, t.Name)
	case *appv1.StatefulSet:
		k.reCreateApp(t.Namespace, t.Name)
	case *corev1.Service:
		k.reCreateApp(t.Namespace, t.Name)
	case *corev1.PersistentVolumeClaim:
		k.reCreateApp(t.Namespace, t.Name)
	default:
		log.Error("收到未知类型的删除事件")
	}
}

func (k *Kube) isOwnCloudApp(namespace, name string) (b bool) {
	app, err := k.lister.appLister.Apps(namespace).Get(name)
	if err == nil && app != nil {
		b = true
	}
	return
}

//处理更新事件,主要用来更新app的status
func (k *Kube) dealUpdateEvent(oldObj, newObj interface{}) {
	switch t := newObj.(type) {
	case *appv1.Deployment:
		if k.isOwnCloudApp(t.Namespace, t.Name) {
			var condition myappv1.Condition = ""
			if *t.Spec.Replicas == 0 {
				condition = myappv1.ConditionSleep
			} else if t.Status.ReadyReplicas == 0 {
				condition = myappv1.ConditionWaitReady
			} else if t.Status.ReadyReplicas > 0 && t.Status.ReadyReplicas < t.Status.Replicas {
				condition = myappv1.ConditionPartReady
			} else if t.Status.ReadyReplicas == t.Status.Replicas {
				condition = myappv1.ConditionAllReady
			}
			k.updateAppStatus(t.Namespace, t.Name, condition)
		}
	}
}
