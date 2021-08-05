package k8s

import (
	"context"
	myappv1 "github.com/coolsun/cloud-app/k8s/crd/pkg/apis/app/v1"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k *Kube) deleteConfigmap(namespace, name string) (err error) {
	err = k.kubeClient.CoreV1().ConfigMaps(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	return
}

func (k *Kube) getConfigmap(namespace, name string) (cm *corev1.ConfigMap, err error) {
	cm, err = k.kubeClient.CoreV1().ConfigMaps(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	return
}

func (k *Kube) buildConfigmap(app *myappv1.App, data map[string]string) (cm *corev1.ConfigMap) {
	cm = &corev1.ConfigMap{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: k.buildObjectMeta(app),
		Data:       data,
	}
	return
}

func (k *Kube) createConfigmap(app *myappv1.App, data map[string]string) (err error) {
	cm := k.buildConfigmap(app, data)
	_, exist := k.CheckResourceExist(ConfigMap, app.Namespace, app.Name)
	if exist {
		_, err = k.kubeClient.CoreV1().ConfigMaps(app.Namespace).Update(context.TODO(), cm, metav1.UpdateOptions{})
		if err != nil {
			err = errors.WithStack(err)
			return
		}
	} else {
		_, err = k.kubeClient.CoreV1().ConfigMaps(app.Namespace).Create(context.TODO(), cm, metav1.CreateOptions{})
		if err != nil {
			err = errors.WithStack(err)
			return
		}
	}
	return
}

func (k *Kube) updateConfigmap(app *myappv1.App, data map[string]string) (err error) {
	_, err = k.kubeClient.CoreV1().ConfigMaps(app.Namespace).Update(context.TODO(), k.buildConfigmap(app, data), metav1.UpdateOptions{})
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	return
}
