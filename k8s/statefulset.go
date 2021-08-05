package k8s

import (
	"context"
	myappv1 "github.com/coolsun/cloud-app/k8s/crd/pkg/apis/app/v1"
	"github.com/coolsun/cloud-app/utils"
	"github.com/pkg/errors"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k *Kube) createStatefulset(app *myappv1.App) (err error) {
	statefulset, err := k.buildStatefulset(app)
	if err != nil {
		return
	}
	_, exist := k.CheckResourceExist(StatefulSet, app.Namespace, app.Name)
	if exist {
		_, err = k.kubeClient.AppsV1().StatefulSets(app.Namespace).Update(context.TODO(), statefulset, metav1.UpdateOptions{})
		if err != nil {
			err = errors.WithStack(err)
			return
		}
	} else {
		_, err = k.kubeClient.AppsV1().StatefulSets(app.Namespace).Create(context.TODO(), statefulset, metav1.CreateOptions{})
		if err != nil {
			err = errors.WithStack(err)
			return
		}
	}
	return
}
func (k *Kube) buildStatefulset(app *myappv1.App) (statefulSet *appv1.StatefulSet, err error) {
	containers, err := k.buildContainers(app)
	if err != nil {
		return
	}
	initContainers, err := k.buildInitContainer(app)
	if err != nil {
		return
	}
	pvc, err := k.buildPVC(app)
	if err != nil {
		return
	}
	statefulSet = &appv1.StatefulSet{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: k.buildObjectMeta(app),
		Spec: appv1.StatefulSetSpec{
			Replicas: utils.GetInt32Pointer(app.Spec.Replicas),
			Selector: &metav1.LabelSelector{
				MatchLabels: k.buildLabels(app.Name),
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: k.buildLabels(app.Name),
				},
				Spec: corev1.PodSpec{
					Affinity:       k.buildAffinity(app),
					InitContainers: initContainers,
					Containers:     containers,
					Volumes:        k.buildVolumes(app),
				},
			},
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{*pvc},
			ServiceName:          k.buildHeadlessServiceName(app),
			PodManagementPolicy:  "",
			UpdateStrategy:       appv1.StatefulSetUpdateStrategy{},
			RevisionHistoryLimit: nil,
		},
	}
	return
}

func (k *Kube) buildHeadlessServiceName(app *myappv1.App) string {
	return app.Name + "-headless"
}
