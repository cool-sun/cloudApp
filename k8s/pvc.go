package k8s

import (
	"context"
	"fmt"
	myappv1 "github.com/coolsun/cloud-app/k8s/crd/pkg/apis/app/v1"
	"github.com/coolsun/cloud-app/utils"
	"github.com/coolsun/cloud-app/utils/log"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

func (k *Kube) deletePVC(namespace, name string) (err error) {
	requirement, err := labels.NewRequirement(CloudAppName, selection.Equals, []string{name})
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	pvcList, err := k.lister.pvcLister.PersistentVolumeClaims(namespace).List(labels.NewSelector().Add(*requirement))
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	for _, v := range pvcList {
		err = k.kubeClient.CoreV1().PersistentVolumeClaims(namespace).Delete(context.TODO(), v.Name, metav1.DeleteOptions{})
		if err != nil {
			log.Errorf("%+v", err)
			continue
		}
	}
	return
}

func (k *Kube) createPVC(app *myappv1.App) (err error) {
	pvc, err := k.buildPVC(app)
	if err != nil {
		return
	}
	_, exist := k.CheckResourceExist(PVC, app.Namespace, app.Name)
	if exist {
	} else {
		_, err = k.kubeClient.CoreV1().PersistentVolumeClaims(app.Namespace).Create(context.TODO(), pvc, metav1.CreateOptions{})
		if err != nil {
			err = errors.WithStack(err)
			return
		}
	}
	return
}

func (k *Kube) createPvcFromSnapshot(oldPvc *corev1.PersistentVolumeClaim, namespace, snapshotName string) (newPvcName string, err error) {
	newPvcName = oldPvc.Name + "-" + utils.GetValidateCode(4)
	APIGroup := "snapshot.storage.k8s.io"
	typedLocalObjectReference := &corev1.TypedLocalObjectReference{
		APIGroup: &APIGroup,
		Kind:     "VolumeSnapshot",
		Name:     snapshotName,
	}
	newPvc := &corev1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:   newPvcName,
			Labels: oldPvc.ObjectMeta.Labels,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes:      oldPvc.Spec.AccessModes,
			Resources:        oldPvc.Spec.Resources,
			StorageClassName: oldPvc.Spec.StorageClassName,
			VolumeMode:       oldPvc.Spec.VolumeMode,
			DataSource:       typedLocalObjectReference,
		},
	}
	_, err = k.kubeClient.CoreV1().PersistentVolumeClaims(namespace).Create(context.TODO(), newPvc, metav1.CreateOptions{})
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	return
}
func (k *Kube) buildPVC(app *myappv1.App) (pvc *corev1.PersistentVolumeClaim, err error) {
	var storage resource.Quantity
	size := app.Spec.Pvc.Size
	sc := app.Spec.Pvc.SC
	if size == "" {
		size = defaultPvcSize
	}
	storage, err = resource.ParseQuantity(fmt.Sprintf("%vGi", size))
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	pvc = &corev1.PersistentVolumeClaim{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: k.buildObjectMeta(app),
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					corev1.ResourceStorage: storage,
				},
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: storage,
				},
			},
		},
	}
	if sc != "" {
		pvc.Spec.StorageClassName = &sc
	}
	return
}
