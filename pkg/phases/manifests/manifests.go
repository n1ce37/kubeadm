package manifests

import (
	"fmt"
	"github.com/n1ce37/kubeadm/pkg/util/maps"
	"net"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/n1ce37/kubeadm/pkg/constants"
)

const (
	etcdClientPort          = 2379
	etcdPeerPort            = 2380
	etcdInitialClusterToken = "kube-etcd-cluster"
)

type Config struct {
	AdvertiseAddress string
	Machines         map[string]net.IP

	SvcNet net.IPNet

	CertDir string
	ConfDir string
}

func GetManifests() map[string][]byte {
	apiserverPod := buildPod(getAPIServerContainer())

	return nil
}

func getEtcdContainers(cfg Config) map[string]v1.Container {
	// https://github.com/etcd-io/etcd/blob/master/Documentation/op-guide/configuration.md
	baseArgs := map[string]string{
		"data-dir":              constants.EtcdDataDir,
		"initial-cluster-token": etcdInitialClusterToken,
		"cert-file":             filepath.Join(cfg.CertDir, constants.EtcdServerCert),
		"key-file":              filepath.Join(cfg.CertDir, constants.EtcdServerKey),
		"trusted-ca-file":       filepath.Join(cfg.CertDir, constants.EtcdCACert),
		"client-cert-auth":      "true",
		"peer-cert-file":        filepath.Join(cfg.CertDir, constants.EtcdPeerCert),
		"peer-key-file":         filepath.Join(cfg.CertDir, constants.EtcdPeerKey),
		"peer-trusted-ca-file":  filepath.Join(cfg.CertDir, constants.EtcdCACert),
		"peer-client-cert-auth": "true",
		"initial-cluster-state": "new",
	}

	containers := make(map[string]v1.Container, len(cfg.Machines))
	for k, v := range cfg.Machines {
		args := maps.DeepCopy(baseArgs)
		args["name"] = k
		args["initial-advertise-peer-urls"] = getEtcdPeerURL(v)
		args["listen-peer-urls"] = getEtcdPeerURL(v)
		args["listen-client-urls"] = getEtcdClientURL(v)
		args["advertise-client-urls"] = getEtcdClientURL(v)
		args["initial-cluster"] = getEtcdInitialCluster(cfg.Machines)
		containers[k] = v1.Container{
			Name:    constants.Etcd,
			Command: buildCommand(constants.Etcd, args),
		}
	}

	return containers
}

func getAPIServerContainer(cfg Config) v1.Container {
	// https://kubernetes.io/docs/reference/command-line-tools-reference/kube-apiserver/
	args := map[string]string{
		"advertise-address":        cfg.AdvertiseAddress,
		"allowed-privileged":       "true",
		"anonymous-auth":           "false",
		"apiserver-count":          strconv.Itoa(len(cfg.Machines)),
		"enable-admission-plugins": "NodeRestriction",
		"service-cluster-ip-range": cfg.SvcNet.String(),
		// TODO constants
		"client-ca-file":             filepath.Join(cfg.CertDir, "sa.pub"),
		"tls-cert-file":              filepath.Join(cfg.CertDir, "apiserver.crt"),
		"tls-private-key-file":       filepath.Join(cfg.CertDir, "apiserver.key"),
		"kubelet-client-certificate": filepath.Join(cfg.CertDir, "apiserver-kubelet-client.crt"),
		"kubelet-client-key":         filepath.Join(cfg.CertDir, "apiserver-kubelet-client.key"),
		"secret-port":                        strconv.Itoa(constants.APIServerPort),
		"kubelet-preferred-address-types":    "InternalIP,ExternalIP,Hostname",
		"requestheader-username-headers":     "X-Remote-User",
		"requestheader-group-headers":        "X-Remote-Group",
		"requestheader-extra-headers-prefix": "X-Remote-Extra-",
		"requestheader-client-ca-file":       filepath.Join(cfg.CertDir, "front-proxy-ca.crt"),
		"requestheader-allowed-names":        "front-proxy-client",
		"proxy-client-cert-file":             "front-proxy-client.crt",
		"proxy-client-key-file":              "front-proxy-client.key",
	}

	return v1.Container{
		Name:    constants.KubeAPIServer,
		Command: buildCommand(constants.KubeAPIServer, args),
	}
}

func getControllerManagerContainer(cfg Config) v1.Container {
	// https://kubernetes.io/docs/reference/command-line-tools-reference/kube-controller-manager/
	kubeConf := filepath.Join(cfg.ConfDir, constants.GetKubeConf(constants.KubeControllerManager)
	args := map[string]string{
		"leader-elect":                     "true",
		"kubeconfig":                       kubeConf,
		"authentication-kubeconfig":        kubeConf,
		"authorization-kubeconfig":         kubeConf,
		"client-ca-file":                   filepath.Join(cfg.CertDir, constants.CACert),
		"requestheader-client-ca-file":     filepath.Join(cfg.CertDir, constants.FrontProxyCert),
		"root-ca-file":                     filepath.Join(cfg.CertDir, constants.CACert),
		"service-account-private-key-file": filepath.Join(cfg.CertDir, constants.ServiceAccountPrivateKey),
		"cluster-signing-cert-file":        filepath.Join(cfg.CertDir, constants.CACert),
		"cluster-signing-key-file":         filepath.Join(cfg.CertDir, constants.CACert),
		"use-service-account-credentials":  "true",
		"controllers":                      "*,bootstrapsigner,tokencleaner",
	}

	return v1.Container{
		Name:    constants.KubeControllerManager,
		Command: buildCommand(constants.KubeControllerManager, args),
	}
}

func getSchedulerContainer(cfg Config) v1.Container {
	kubeConf := filepath.Join(cfg.ConfDir, constants.GetKubeConf(constants.KubeScheduler))
	args := map[string]string{
		"leader-elect":              "true",
		"kubeconfig":                kubeConf,
		"authentication-kubeconfig": kubeConf,
		"authorization-kubeconfig":  kubeConf,
	}
	return v1.Container{
		Name:    constants.KubeScheduler,
		Command: buildCommand(constants.KubeScheduler, args),
	}
}

func buildCommand(baseCommand string, args map[string]string) []string {
	keys := make([]string, 0, len(args))
	for k, _ := range args {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	command := make([]string, 0, len(args)+1)
	command = append(command, baseCommand)
	for _, v := range keys {
		command = append(command, fmt.Sprintf("--%s=%s", v, args[v]))
	}

	return command
}

func buildPod(container v1.Container) v1.Pod {
	return v1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      container.Name,
			Namespace: metav1.NamespaceSystem,
			Labels:    map[string]string{"component": container.Name},
		},
		Spec: v1.PodSpec{
			Containers:        []v1.Container{container},
			PriorityClassName: "system-cluster-critical",
			HostNetwork:       true,
			// TODO
			//Volumes: []
		},
	}
}

func getEtcdClientURL(ip net.IP) string {
	return fmt.Sprintf("https://%s:%s", ip.String(), strconv.Itoa(etcdClientPort))
}

func getEtcdPeerURL(ip net.IP) string {
	return fmt.Sprintf("https://%s:%s", ip.String(), strconv.Itoa(etcdPeerPort))
}

func getEtcdInitialCluster(machines map[string]net.IP) string {
	nodes := make([]string, 0, len(machines))

	for k, v := range machines {
		nodes = append(nodes, fmt.Sprintf("%s=%s", k, getEtcdPeerURL(v)))
	}

	return strings.Join(nodes, ",")
}
