package k8s

import (
	"context"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k *Kube) deleteHpa(namespace, name string) (err error) {
	err = k.kubeClient.AutoscalingV1().HorizontalPodAutoscalers(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	return
}
