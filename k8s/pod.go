package k8s

import (
	"bytes"
	"context"
	"fmt"
	myappv1 "github.com/coolsun/cloud-app/k8s/crd/pkg/apis/app/v1"
	"github.com/coolsun/cloud-app/model"
	"github.com/coolsun/cloud-app/utils/log"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
	"path"
	"strings"
)

func (k *Kube) deleteAppPods(namespace, name string) (err error) {
	err = k.kubeClient.CoreV1().Pods(namespace).DeleteCollection(context.TODO(), metav1.DeleteOptions{}, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%v=%v", CloudAppName, name),
	})
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	return
}

func (k *Kube) getDefaultConfigFromPod(namespace, appName, configDir string, configFiles []string) (config map[string]string, err error) {
	pods, err := k.getAppPods(namespace, appName)
	if err != nil {
		return
	}
	pod := pods[0]
	config = make(map[string]string, 0)
	for _, v := range configFiles {
		completePath := path.Join(configDir, v)
		s1, _, err := k.execInPod(namespace, pod.Name, string(myappv1.Main), "cat "+completePath)
		if err != nil {
			log.Errorf("%+v", err)
			continue
		}
		config[v] = s1
	}

	return
}

//获取某个app的所有pod
func (k *Kube) getAppPods(namespace, appName string) (pods []corev1.Pod, err error) {
	podList, err := k.kubeClient.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%v=%v", CloudAppName, appName),
	})
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	items := podList.Items
	if len(items) == 0 {
		err = errors.New(model.E10017)
		return
	}
	pods = items
	return
}

func (k *Kube) execInPod(namespace, podName, containerName, command string) (string, string, error) {
	cmd := []string{
		"sh",
		"-c",
		command,
	}
	const tty = false
	req := k.kubeClient.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).SubResource("exec").Param("container", containerName)
	req.VersionedParams(
		&corev1.PodExecOptions{
			Command: cmd,
			Stdin:   false,
			Stdout:  true,
			Stderr:  true,
			TTY:     tty,
		},
		scheme.ParameterCodec,
	)

	var stdout, stderr bytes.Buffer
	exec, err := remotecommand.NewSPDYExecutor(restConfig, "POST", req.URL())
	if err != nil {
		return "", "", err
	}
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  nil,
		Stdout: &stdout,
		Stderr: &stderr,
	})
	if err != nil {
		return "", "", err
	}
	return strings.TrimSpace(stdout.String()), strings.TrimSpace(stderr.String()), err
}
