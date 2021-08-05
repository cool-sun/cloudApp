package k8s

import (
	"context"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k *Kube) createNamespace(name string) (namespace *corev1.Namespace, err error) {
	ns := &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: name,
		},
	}
	nsInter, exist := k.CheckResourceExist(Namespace, "", name)
	if exist {
		namespace = nsInter.(*corev1.Namespace)
	} else {
		namespace, err = k.kubeClient.CoreV1().Namespaces().Create(context.TODO(), ns, metav1.CreateOptions{})
		if err != nil {
			err = errors.WithStack(err)
			return
		}
	}
	return
}
