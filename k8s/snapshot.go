package k8s

import (
	"context"
	"github.com/coolsun/cloud-app/model"
	vssv1 "github.com/kubernetes-csi/external-snapshotter/client/v4/apis/volumesnapshot/v1"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k *Kube) snapshotDelete(namespace, name string) (err error) {
	err1 := k.vssClient.SnapshotV1().VolumeSnapshots(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err1 != nil {
		if !apierrors.IsNotFound(err1) {
			err = errors.WithStack(err1)
		}
		return
	}
	return
}
func (k *Kube) snapshotCreate(namespace, appName, name string) (err error) {
	pvc, err := k.lister.pvcLister.PersistentVolumeClaims(namespace).Get(appName)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	sc, err := k.lister.scLister.Get(*pvc.Spec.StorageClassName)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	volumeSnapshotClass, err := k.getVolumeSnapshotClassByProvisioner(sc.Provisioner)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	if volumeSnapshotClass == nil {
		err = errors.New(model.E10021)
		return
	}
	return k.createSnapshot(volumeSnapshotClass.Name, namespace, appName, name)
}

func (k *Kube) createSnapshot(volumeSnapshotClassName, namespace, appName, name string) (err error) {
	volumeSnapshot := &vssv1.VolumeSnapshot{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    k.buildLabels(appName),
		},
		Spec: vssv1.VolumeSnapshotSpec{
			Source: vssv1.VolumeSnapshotSource{
				PersistentVolumeClaimName: &appName,
			},
			VolumeSnapshotClassName: &volumeSnapshotClassName,
		},
	}
	_, err = k.vssClient.SnapshotV1().VolumeSnapshots(namespace).Create(context.TODO(), volumeSnapshot, metav1.CreateOptions{})
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	return
}

func (k *Kube) getVolumeSnapshotClassByProvisioner(provisioner string) (vsc *vssv1.VolumeSnapshotClass, err error) {
	list, err := k.vssClient.SnapshotV1().VolumeSnapshotClasses().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	for _, v := range list.Items {
		if v.Driver == provisioner {
			vsc = &v
			return
		}
	}
	return
}
