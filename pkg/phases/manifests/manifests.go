package manifests

import (
	"fmt"
	"path/filepath"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetManifests() map[string][]byte {
	apiserverPod := buildPod(getAPIServerContainer())

	return nil
}

func getEtcdContainers() []v1.Container {
	args := map[string]string{
		"name": fmt.Sprint("etcd-%s", "ip"),
		"listen-client-urls": fmt.Sprintf("%s,%s", "", ""),
		"advertise-client-urls":       "",
		"listen-peer-urls":            "",
		"initial-advertise-peer-urls": "",
		"data-dir":                    "",
		"cert-file":                   "",
		"key-file":                    "",
		"trusted-ca-file":             "",
		"client-cert-auth":            "true",
		"peer-cert-file":              "",
		"peer-key-file":               "",
		"peer-trusted-ca-file":        "",
		"peer-client-cert-auth":       "true",
		"snapshot-count":              "10000",
		"listen-metrics-urls":         fmt.Sprintf("http://127.0.0.1:%d", ""),
	}

	//return v1.Container{
	//	TODO constants
		//Name: "etcd",
		//Image: "",
		//Command:buildCommand("etcd", args),
	//}
	return []v1.Container{}
}

func getAPIServerContainer() v1.Container {
	args := map[string]string{
		"advertise-address": "xx",
		"insecure-port": "0",
		"enable-admisson-plugins": "",
		"service-cluster-ip-range": "",
		"client-ca-file": "",
		"tls-cert-file": "",
		"tls-private-key-file": "",
		"kubelet-client-certificate": "",
		"kubelet-client-key": "",
		"secret-port": "",
		"allow-privileged": "true",
		"kubelet-preferred-address-types": "",
		"requestheader-username-headers":     "X-Remote-User",
		"requestheader-group-headers":        "X-Remote-Group",
		"requestheader-extra-headers-prefix": "X-Remote-Extra-",
		"requestheader-client-ca-file":       "",
		"requestheader-allowed-names":        "front-proxy-client",
		"proxy-client-cert-file":             "",
		"proxy-client-key-file":"",
	}

	return v1.Container{
		Name: "kube-apiserver",
		Command: buildCommand("", args),
	}
}

func getControllerManagerContainer() v1.Container {
	return v1.Container{
		Name: "kube-controller-manager",
	}
}

func getSchedulerContainer() v1.Container {
	return v1.Container{
		Name: "kube-scheduler",
	}
}

func buildCommand(baseCommand string, args map[string]string) []string {
	command := make([]string, 0, len(args)+1)

	command = append(command, baseCommand)
	for k, v := range args {
		command = append(command, fmt.Sprintf("--%s=%s", k, v))
	}

	return command
}

func buildPod(container v1.Container) v1.Pod {
	return v1.Pod{
		TypeMeta:   metav1.TypeMeta{
			APIVersion: "v1",
			Kind: "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:                       container.Name,
			Namespace: metav1.NamespaceSystem,
			Labels: map[string]string{"component": container.Name},
		},
		Spec:       v1.PodSpec{
			Containers: []v1.Container{container},
			PriorityClassName: "system-cluster-critical",
			HostNetwork: true,
			// TODO
			//Volumes: []
		},
	}
}