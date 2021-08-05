package k8s

import (
	"context"
	"fmt"
	myappv1 "github.com/coolsun/cloud-app/k8s/crd/pkg/apis/app/v1"
	utils "github.com/coolsun/cloud-app/utils"
	"github.com/coolsun/cloud-app/utils/log"
	"github.com/pkg/errors"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"path"
)

func (k *Kube) updateDeployment(namespace string, deploy *appv1.Deployment) (err error) {
	_, err = k.kubeClient.AppsV1().Deployments(namespace).Update(context.TODO(), deploy, metav1.UpdateOptions{})
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	return
}

func (k *Kube) buildConfigmapVolumeName(name string) string {
	return name + "-configmap-volume"
}

func (k *Kube) buildConfigmapVolume(name string) corev1.Volume {
	return corev1.Volume{
		Name: k.buildConfigmapVolumeName(name),
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: name,
				},
			},
		},
	}
}

func (k *Kube) buildConfigmapVolumeMount(name, configMountPath string) corev1.VolumeMount {
	return corev1.VolumeMount{
		Name:      k.buildConfigmapVolumeName(name),
		MountPath: configMountPath,
	}
}

func (k *Kube) updateDeploymentMountConfigmap(namespace, appName, configMountPath string) (err error) {
	deploy, err := k.getDeployment(namespace, appName)
	if err != nil {
		return
	}
	volume := k.buildConfigmapVolume(appName)
	volumeMount := k.buildConfigmapVolumeMount(appName, configMountPath)
	deploy.Spec.Template.Spec.Volumes = append(deploy.Spec.Template.Spec.Volumes, volume)
	for i, v := range deploy.Spec.Template.Spec.Containers {
		if v.Name == string(myappv1.Main) {
			deploy.Spec.Template.Spec.Containers[i].VolumeMounts = append(deploy.Spec.Template.Spec.Containers[i].VolumeMounts, volumeMount)
		}
	}
	return k.updateDeployment(namespace, deploy)
}

func (k *Kube) getDeployment(namespace, name string) (deploy *appv1.Deployment, err error) {
	deploy, err = k.lister.deployLister.Deployments(namespace).Get(name)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	return
}

func (k *Kube) deleteDeployment(namespace, name string) (err error) {
	err = k.kubeClient.AppsV1().Deployments(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	return
}

func (k *Kube) updateDeploymentVolumes(namespace, name string, newVolumeName string) (err error) {
	deploy, err := k.lister.deployLister.Deployments(namespace).Get(name)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	oldVolumes := deploy.Spec.Template.Spec.Volumes
	for i, v := range oldVolumes {
		if v.VolumeSource.PersistentVolumeClaim != nil {
			deploy.Spec.Template.Spec.Volumes[i].VolumeSource.PersistentVolumeClaim.ClaimName = newVolumeName
		}
	}
	deploy.Spec.Template.Spec.Volumes = oldVolumes
	_, err = k.kubeClient.AppsV1().Deployments(namespace).Update(context.TODO(), deploy, metav1.UpdateOptions{})
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	return
}
func (k *Kube) createDeployment(app *myappv1.App) (err error) {
	deployment, err := k.buildDeployment(app)
	if err != nil {
		return
	}
	d, exist := k.CheckResourceExist(Deployment, app.Namespace, app.Name)
	if exist {
		oldDeployment := d.(*appv1.Deployment)
		deployment.Spec.Template.Spec.Volumes = oldDeployment.Spec.Template.Spec.Volumes
		_, err = k.kubeClient.AppsV1().Deployments(app.Namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})
		if err != nil {
			err = errors.WithStack(err)
			return
		}
	} else {
		_, err = k.kubeClient.AppsV1().Deployments(app.Namespace).Create(context.TODO(), deployment, metav1.CreateOptions{})
		if err != nil {
			err = errors.WithStack(err)
			return
		}
	}
	return
}

//使用 ObjectMeta属性中的 OwnerRefrence 来做资源关联，有两个特性：
//1.Owner 资源被删除，被 Own 的资源会被级联删除，这利用了 K8s 的 GC；
//1.被 Own 的资源对象的事件变更可以触发 Owner 对象的 Reconcile 方法；
func (k *Kube) buildObjectMeta(app *myappv1.App) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      app.Name,
		Namespace: app.Namespace,
		Labels:    k.buildLabels(app.Name),
		OwnerReferences: []metav1.OwnerReference{{
			APIVersion: fmt.Sprintf("%v/%v", myappv1.GroupName, myappv1.Version),
			Kind:       Kind,
			Name:       app.Name,
			UID:        app.ObjectMeta.UID,
		}},
	}
}

func (k *Kube) buildDeployment(app *myappv1.App) (d *appv1.Deployment, err error) {
	containers, err := k.buildContainers(app)
	if err != nil {
		return
	}
	initContainers, err := k.buildInitContainer(app)
	if err != nil {
		return
	}
	d = &appv1.Deployment{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: k.buildObjectMeta(app),
		Spec: appv1.DeploymentSpec{
			Replicas: utils.GetInt32Pointer(app.Spec.Replicas),
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
			Strategy:                k.buildStrategy(app),
			MinReadySeconds:         0,
			RevisionHistoryLimit:    nil,
			Paused:                  false,
			ProgressDeadlineSeconds: nil,
		},
	}
	return
}

func (k *Kube) buildStrategy(app *myappv1.App) (strategy appv1.DeploymentStrategy) {
	if app.Spec.Strategy == StrategyRecreate {
		strategy = appv1.DeploymentStrategy{
			Type: appv1.RecreateDeploymentStrategyType,
		}
	}
	return
}

//亲和性，这块按如下规则自动处理
//节点亲和性
//1.把pod调度到app名称所在的节点，软性的；
//2.把pod调度到app小类所在的节点，软性的；
//3.把pod调度到app大类所在的节点，软性的；
//pod反亲和
//1.同名的pod不要在同一节点，软性的；
//2.同一app小类的节点尽量不要在同一节点；
//2.同一app大类的节点尽量不要在同一节点；
//上述规则都用软性的，防止因节点数不足导致pod调度失败的问题
func (k *Kube) buildAffinity(app *myappv1.App) (affinity *corev1.Affinity) {
	affinity = &corev1.Affinity{
		NodeAffinity: &corev1.NodeAffinity{
			PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{
				{
					Weight: 100,
					Preference: corev1.NodeSelectorTerm{
						MatchExpressions: []corev1.NodeSelectorRequirement{{
							Key:      CloudAppName,
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{app.Name},
						}},
					},
				},
				{
					Weight: 90,
					Preference: corev1.NodeSelectorTerm{
						MatchExpressions: []corev1.NodeSelectorRequirement{{
							Key:      CloudAppLittleCategory,
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{app.Spec.AppLittleCategory},
						}},
					},
				},
				{
					Weight: 80,
					Preference: corev1.NodeSelectorTerm{
						MatchExpressions: []corev1.NodeSelectorRequirement{{
							Key:      CloudAppBigCategory,
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{app.Spec.AppBigCategory},
						}},
					},
				},
			},
		},
		PodAntiAffinity: &corev1.PodAntiAffinity{
			PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
				{
					Weight: 100,
					PodAffinityTerm: corev1.PodAffinityTerm{
						LabelSelector: &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{
									Key:      CloudAppName,
									Operator: metav1.LabelSelectorOpIn,
									Values:   []string{app.Name},
								},
							},
						},
						TopologyKey: "kubernetes.io/hostname",
					},
				},
				{
					Weight: 90,
					PodAffinityTerm: corev1.PodAffinityTerm{
						LabelSelector: &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{
									Key:      CloudAppLittleCategory,
									Operator: metav1.LabelSelectorOpIn,
									Values:   []string{app.Spec.AppLittleCategory},
								},
							},
						},
						TopologyKey: "kubernetes.io/hostname",
					},
				},
				{
					Weight: 80,
					PodAffinityTerm: corev1.PodAffinityTerm{
						LabelSelector: &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{
									Key:      CloudAppBigCategory,
									Operator: metav1.LabelSelectorOpIn,
									Values:   []string{app.Spec.AppBigCategory},
								},
							},
						},
						TopologyKey: "kubernetes.io/hostname",
					},
				},
			},
		},
	}
	return
}

func (k *Kube) buildLabels(appName string) (l map[string]string) {
	l = make(map[string]string)
	l[CloudAppName] = appName
	return
}

func (k *Kube) buildInitContainer(app *myappv1.App) (ic []corev1.Container, err error) {
	ic = make([]corev1.Container, 0, 0)
	for mapK, mapV := range app.Spec.Pod {
		if mapK == myappv1.Init && mapV != nil {
			c1 := corev1.Container{
				Name:            string(mapK),
				Image:           k.buildContainerImage(app)[mapK],
				Command:         mapV.Command,
				Args:            mapV.Args,
				Ports:           k.buildContainerPorts(app)[mapK],
				Env:             k.buildContainerEnv(app)[mapK],
				Resources:       k.buildContainerResources(app)[mapK],
				VolumeMounts:    k.buildContainerVolumeMounts(app)[mapK],
				ImagePullPolicy: corev1.PullIfNotPresent,
			}
			ic = append(ic, c1)
		}
	}
	return
}

func (k *Kube) buildContainers(app *myappv1.App) (cs []corev1.Container, err error) {
	cs = make([]corev1.Container, 0, 0)
	for mapK, mapV := range app.Spec.Pod {
		if mapK == myappv1.Init || mapV == nil {
			continue
		}
		c1 := corev1.Container{
			Name:            string(mapK),
			Image:           k.buildContainerImage(app)[mapK],
			Command:         mapV.Command,
			Args:            mapV.Args,
			Ports:           k.buildContainerPorts(app)[mapK],
			Env:             k.buildContainerEnv(app)[mapK],
			Resources:       k.buildContainerResources(app)[mapK],
			VolumeMounts:    k.buildContainerVolumeMounts(app)[mapK],
			ImagePullPolicy: corev1.PullIfNotPresent,
		}
		cs = append(cs, c1)
	}
	return
}

func (k *Kube) buildContainerResources(app *myappv1.App) (rs map[myappv1.ContainerType]corev1.ResourceRequirements) {
	rs = make(map[myappv1.ContainerType]corev1.ResourceRequirements, 0)
	for mapK, mapV := range app.Spec.Pod {
		if mapV == nil {
			continue
		}
		reqCpu, _ := resource.ParseQuantity("100m")
		reqMem, _ := resource.ParseQuantity("128Mi")
		limitCpu, err := resource.ParseQuantity(fmt.Sprintf("%v", mapV.CPU))
		if err != nil {
			log.Errorf("%+v", err)
			continue
		}
		limitMem, err := resource.ParseQuantity(fmt.Sprintf("%vGi", mapV.Mem))
		if err != nil {
			log.Errorf("%+v", err)
			continue
		}
		obj := corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    reqCpu,
				corev1.ResourceMemory: reqMem,
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    limitCpu,
				corev1.ResourceMemory: limitMem,
			},
		}
		rs[mapK] = obj
	}
	return
}

func (k *Kube) buildVolumes(app *myappv1.App) (vols []corev1.Volume) {
	vols = make([]corev1.Volume, 0, 0)
	if app.Spec.Pvc != nil {
		obj := corev1.Volume{
			Name: app.Name,
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: app.Name,
				},
			},
		}
		vols = append(vols, obj)
	}

	//检查有没有同名称的configmap存在，如果有的话也挂载进去
	_, exist := k.CheckResourceExist(ConfigMap, app.Namespace, app.Name)
	if exist {
		vols = append(vols, k.buildConfigmapVolume(app.Name))
	}
	return
}

func (k *Kube) buildContainerVolumeMounts(app *myappv1.App) (volms map[myappv1.ContainerType][]corev1.VolumeMount) {
	volms = make(map[myappv1.ContainerType][]corev1.VolumeMount, 0)

	for mapK, mapV := range app.Spec.Pod {
		if mapV == nil {
			continue
		}
		obj := make([]corev1.VolumeMount, 0, 0)
		for _, v := range mapV.VolumeMount {
			vm := corev1.VolumeMount{
				Name:      app.Name,
				MountPath: v,
				SubPath:   path.Base(v),
			}
			obj = append(obj, vm)
		}
		if mapK == myappv1.Main {
			//检查有没有同名称的configmap存在，如果有的话也挂载进去
			_, exist := k.CheckResourceExist(ConfigMap, app.Namespace, app.Name)
			if exist {
				obj = append(obj, k.buildConfigmapVolumeMount(app.Name, app.Spec.Config.Path))
			}
		}
		volms[mapK] = obj
	}

	return
}

func (k *Kube) buildContainerPorts(app *myappv1.App) (ports map[myappv1.ContainerType][]corev1.ContainerPort) {
	ports = make(map[myappv1.ContainerType][]corev1.ContainerPort, 0)
	for mapK, mapV := range app.Spec.Pod {
		if mapV == nil {
			continue
		}
		obj := make([]corev1.ContainerPort, 0, 0)
		for i, v := range mapV.Ports {
			p := corev1.ContainerPort{
				Name:          fmt.Sprintf("port%v", i+1),
				ContainerPort: int32(v.GetPort()),
				Protocol:      v.GetProtocol(),
			}
			obj = append(obj, p)
		}
		ports[mapK] = obj
	}

	return
}

func (k *Kube) buildContainerImage(app *myappv1.App) (image map[myappv1.ContainerType]string) {
	image = make(map[myappv1.ContainerType]string, 0)
	for mapK, mapV := range app.Spec.Pod {
		if mapV == nil {
			continue
		}
		image[mapK] = path.Join(mapV.Image + ":" + mapV.Tag)
	}
	return
}

func (k *Kube) buildContainerEnv(app *myappv1.App) (env map[myappv1.ContainerType][]corev1.EnvVar) {
	env = make(map[myappv1.ContainerType][]corev1.EnvVar, 0)
	for mapK, mapV := range app.Spec.Pod {
		if mapV == nil {
			continue
		}
		obj := make([]corev1.EnvVar, 0, 0)
		for _, v := range mapV.Env {
			e := corev1.EnvVar{
				Name:  v.GetName(),
				Value: v.GetValue(),
			}
			obj = append(obj, e)
		}
		env[mapK] = obj
	}

	return
}
