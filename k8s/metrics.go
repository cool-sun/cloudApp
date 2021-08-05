package k8s

import (
	"context"
	"fmt"
	"github.com/coolsun/cloud-app/utils/log"
	json "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	metrics_v1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

func (k *Kube) getMetric(namespace, appName string) (podMetrics []metrics_v1beta1.PodMetrics, err error) {
	//获取一个app的所有pod监控信息

	//podMetricsList, err := k.MetricsClient.MetricsV1beta1().PodMetricses(namespace).List(context.TODO(), metav1.ListOptions{
	//	LabelSelector: fmt.Sprintf("%v=%v", CloudAppName, name),
	//})
	//if err != nil {
	//	err = errors.WithStack(err)
	//	return
	//}
	//podMetrics = podMetricsList.Items
	return
}

func (k *Kube) createByYaml(appName string) {
	podMetricsList, err := k.metricsClient.MetricsV1beta1().PodMetricses("default").List(context.TODO(), metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=%s", "redis"),
	})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(podMetricsList)

	config, err := inClusterConnect()
	if err != nil {
		if k.cfg.KubeConfig == "" {
			config, err = outClusterConnect()
		} else {
			config, err = outClusterConnect(k.cfg.KubeConfig)
		}
		if err != nil {
			err = errors.WithStack(err)
			return
		}
	}
	log.Error(doSSA(context.TODO(), config))
}

const deploymentYAML = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  namespace: default
spec:
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:latest
`

func doSSA(ctx context.Context, cfg *rest.Config) error {
	var decUnstructured = yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	// 1. Prepare a RESTMapper to find GVR
	dc, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return err
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))

	// 2. Prepare the dynamic client
	dyn, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return err
	}

	// 3. Decode YAML manifest into unstructured.Unstructured
	obj := &unstructured.Unstructured{}
	_, gvk, err := decUnstructured.Decode([]byte(deploymentYAML), nil, obj)
	if err != nil {
		return err
	}

	// 4. Find GVR
	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return err
	}

	// 5. Obtain REST interface for the GVR
	var dr dynamic.ResourceInterface
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		// namespaced resources should specify the namespace
		dr = dyn.Resource(mapping.Resource).Namespace(obj.GetNamespace())
	} else {
		// for cluster-wide resources
		dr = dyn.Resource(mapping.Resource)
	}

	// 6. Marshal object into JSON
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	// 7. Create or Update the object with SSA
	//     types.ApplyPatchType indicates SSA.
	//     FieldManager specifies the field owner ID.
	_, err = dr.Patch(context.TODO(), obj.GetName(), types.ApplyPatchType, data, metav1.PatchOptions{
		FieldManager: "sample-controller",
	})

	return err
}
