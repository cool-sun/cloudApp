package k8s

import (
	"context"
	"fmt"
	myappv1 "github.com/coolsun/cloud-app/k8s/crd/pkg/apis/app/v1"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (k *Kube) revertServiceType(namespace, name string, isOpen bool) (err error) {
	svc, err := k.kubeClient.CoreV1().Services(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	if !isOpen && svc.Spec.Type == corev1.ServiceTypeClusterIP {
		return
	}
	if isOpen && svc.Spec.Type == corev1.ServiceTypeNodePort {
		return
	}
	svcType := svc.Spec.Type
	if svcType == corev1.ServiceTypeClusterIP {
		svcType = corev1.ServiceTypeNodePort
	} else if svcType == corev1.ServiceTypeNodePort {
		svcType = corev1.ServiceTypeClusterIP
	}
	svc.Spec.Type = svcType
	_, err = k.kubeClient.CoreV1().Services(namespace).Update(context.TODO(), svc, metav1.UpdateOptions{})
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	return
}

func (k *Kube) deleteService(namespace, name string) (err error) {
	err = k.kubeClient.CoreV1().Services(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	return
}

func getPorts(old, new []corev1.ServicePort) (out []corev1.ServicePort) {
	out = make([]corev1.ServicePort, 0, 0)
	for _, n := range new {
		var p *corev1.ServicePort
		for _, o := range old {
			if n.Port == o.Port {
				p = &o
				break
			}
		}
		if p != nil {
			out = append(out, *p)
		} else {
			out = append(out, n)
		}
	}
	return
}

func (k *Kube) createHeadlessService(app *myappv1.App) (err error) {
	svc, err := k.buildService(app)
	if err != nil {
		return
	}
	svc.Name = k.buildHeadlessServiceName(app)
	svc.Spec.ClusterIP = "None"
	return k.createServiceCommon(svc)
}

func (k *Kube) createServiceCommon(svc *corev1.Service) (err error) {
	o, exist := k.CheckResourceExist(Service, svc.Namespace, svc.Name)
	if exist {
		oldSvc := o.(*corev1.Service)
		oldSvc.Spec.Ports = getPorts(oldSvc.Spec.Ports, svc.Spec.Ports)
		_, err = k.kubeClient.CoreV1().Services(svc.Namespace).Update(context.TODO(), oldSvc, metav1.UpdateOptions{})
		if err != nil {
			err = errors.WithStack(err)
			return
		}
	} else {
		_, err = k.kubeClient.CoreV1().Services(svc.Namespace).Create(context.TODO(), svc, metav1.CreateOptions{})
		if err != nil {
			err = errors.WithStack(err)
			return
		}
	}
	return
}

func (k *Kube) createService(app *myappv1.App) (err error) {
	svc, err := k.buildService(app)
	if err != nil {
		return
	}
	return k.createServiceCommon(svc)
}

func (k *Kube) buildService(app *myappv1.App) (svc *corev1.Service, err error) {
	svc = &corev1.Service{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: k.buildObjectMeta(app),
		Spec: corev1.ServiceSpec{
			Ports:    k.buildServicePorts(app),
			Selector: k.buildLabels(app.Name),
			Type:     k.buildServiceType(app),
		},
	}
	return
}

func (k *Kube) buildServiceType(app *myappv1.App) corev1.ServiceType {
	if app.Spec.InternetAccess {
		return corev1.ServiceTypeNodePort
	}
	return corev1.ServiceTypeClusterIP
}

func (k *Kube) buildServicePorts(app *myappv1.App) (ports []corev1.ServicePort) {
	ports = make([]corev1.ServicePort, 0, 0)
	for _, mapV := range app.Spec.Pod {
		if mapV == nil {
			continue
		}
		for i, v := range mapV.Ports {
			obj := corev1.ServicePort{
				Name:       fmt.Sprintf("%v-port-%v", app.Name, i),
				Protocol:   v.GetProtocol(),
				Port:       int32(v.GetPort()),
				TargetPort: intstr.IntOrString{IntVal: int32(v.GetPort())},
			}
			ports = append(ports, obj)
		}
	}
	return
}
