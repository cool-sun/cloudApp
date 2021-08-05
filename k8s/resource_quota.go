package k8s

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	k8s_error "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k *Kube) deleteResourceQuota(name string) (err error) {
	err1 := k.kubeClient.CoreV1().ResourceQuotas(name).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err1 != nil && !k8s_error.IsNotFound(err1) {
		err = errors.WithStack(err1)
		return
	}
	return
}

func (k *Kube) createResourceQuota(name string, cpu, mem string) (err error) {
	cpuQuantity, err := resource.ParseQuantity(fmt.Sprintf("%vGi", cpu))
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	memoryQuantity, err := resource.ParseQuantity(fmt.Sprintf("%vGi", mem))
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	rq := &corev1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: name,
		},
		Spec: corev1.ResourceQuotaSpec{
			Hard: map[corev1.ResourceName]resource.Quantity{
				corev1.ResourceLimitsCPU:    cpuQuantity,
				corev1.ResourceLimitsMemory: memoryQuantity,
			},
		},
	}
	_, exist := k.CheckResourceExist(ResourceQuota, rq.Namespace, rq.Name)
	if exist {
		_, err = k.kubeClient.CoreV1().ResourceQuotas(name).Update(context.TODO(), rq, metav1.UpdateOptions{})
		if err != nil {
			err = errors.WithStack(err)
			return
		}
	} else {
		_, err = k.kubeClient.CoreV1().ResourceQuotas(name).Create(context.TODO(), rq, metav1.CreateOptions{})
		if err != nil {
			err = errors.WithStack(err)
			return
		}
	}
	return
}
