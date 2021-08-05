package k8s

import (
	"github.com/pkg/errors"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func (k *Kube) GetScList() (scList []*storagev1.StorageClass, err error) {
	scList, err = k.lister.scLister.List(labels.NewSelector())
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	return
}
