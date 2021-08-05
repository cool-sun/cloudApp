package k8s

import (
	"context"
	"fmt"
	"github.com/coolsun/cloud-app/model"
	helmclient "github.com/mittwald/go-helm-client"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/repo"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

func (k *Kube) uninstallRelease(namespace, releaseName string) (err error) {
	client, err := getClient(namespace)
	if err != nil {
		return
	}
	err = client.UninstallRelease(&helmclient.ChartSpec{
		ReleaseName: releaseName,
	})
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	return
}

func getClient(namespace string) (client helmclient.Client, err error) {
	client, err = helmclient.NewClientFromRestConf(&helmclient.RestConfClientOptions{
		Options: &helmclient.Options{
			Namespace: namespace,
		},
		RestConfig: restConfig,
	})
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	return
}
func (k *Kube) installOrUpgradeChart(namespace, repoName, repoURL, releaseName, chartName, version, values string) (err error) {
	client, err := getClient(namespace)
	if err != nil {
		return
	}
	err = client.AddOrUpdateChartRepo(repo.Entry{
		Name: repoName,
		URL:  repoURL,
	})
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	err = client.UpdateChartRepos()
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	err = client.InstallOrUpgradeChart(context.TODO(), &helmclient.ChartSpec{
		ReleaseName: releaseName,
		ChartName:   chartName,
		Namespace:   namespace,
		Version:     version,
		ValuesYaml:  values,
	})
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	return
}

func (k *Kube) getReleaseValues(namespace, releaseName string) (m map[string]interface{}, err error) {
	client, err := getClient(namespace)
	if err != nil {
		return
	}
	m, err = client.GetReleaseValues(releaseName, true)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	return
}

func (k *Kube) getReleaseEndPoint(namespace, releaseName string) (endPoint *model.EndPoint, err error) {
	endPoint = &model.EndPoint{
		In:  make([]string, 0),
		Out: make([]string, 0),
	}
	requirement, err := labels.NewRequirement("app.kubernetes.io/instance", selection.Equals, []string{releaseName})
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	ret, err := k.lister.svcLister.Services(namespace).List(labels.NewSelector().Add(*requirement))
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	for _, svc := range ret {
		for _, v := range svc.Spec.Ports {
			endPoint.In = append(endPoint.In, fmt.Sprintf("%v.%v.svc:%v", svc.Name, namespace, v.Port))
		}
		if svc.Spec.Type == corev1.ServiceTypeNodePort {
			for _, v := range svc.Spec.Ports {
				endPoint.Out = append(endPoint.Out, fmt.Sprintf("%v:%v", k.cfg.MasterIp, v.NodePort))
			}
		}
	}
	return
}
