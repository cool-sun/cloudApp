package k8s

import (
	"context"
	myappv1 "github.com/coolsun/cloud-app/k8s/crd/pkg/apis/app/v1"
	"github.com/pkg/errors"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k *Kube) createDaemonSet(app *myappv1.App) (err error) {
	daemonSet, err := k.buildDaemonSet(app)
	if err != nil {
		return
	}
	_, exist := k.CheckResourceExist(DaemonSet, app.Namespace, app.Name)
	if exist {
		_, err = k.kubeClient.AppsV1().DaemonSets(app.Namespace).Update(context.TODO(), daemonSet, metav1.UpdateOptions{})
		if err != nil {
			err = errors.WithStack(err)
			return
		}
	} else {
		_, err = k.kubeClient.AppsV1().DaemonSets(app.Namespace).Create(context.TODO(), daemonSet, metav1.CreateOptions{})
		if err != nil {
			err = errors.WithStack(err)
			return
		}
	}
	return
}

func (k *Kube) buildDaemonSet(app *myappv1.App) (d *appv1.DaemonSet, err error) {
	containers, err := k.buildContainers(app)
	if err != nil {
		return
	}
	initContainers, err := k.buildInitContainer(app)
	if err != nil {
		return
	}
	d = &appv1.DaemonSet{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: k.buildObjectMeta(app),
		Spec: appv1.DaemonSetSpec{
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
			MinReadySeconds:      0,
			RevisionHistoryLimit: nil,
		},
	}
	return
}
