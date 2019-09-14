package constants

import "fmt"

const (
	// components
	Etcd                  = "etcd"
	KubeAPIServer         = "kube-apiserver"
	KubeControllerManager = "kube-controller-manager"
	KubeScheduler         = "kube-scheduler"

	CACert                   = "ca.crt"
	FrontProxyCert           = "front-proxy-ca.crt"
	ServiceAccountPrivateKey = "sa.key"
	EtcdCACert               = "etcd-ca.crt"
	EtcdCAKey                = "etcd-ca.key"
	EtcdServerCert           = "etcd-server.crt"
	EtcdServerKey            = "etcd-server.key"
	EtcdPeerCert             = "etcd-peer.crt"
	EtcdPeerKey              = "etcd-peer.key"

	// etcd
	EtcdDataDir    = "/var/lib/etcd/data"
	EtcdMetricPort = 2381

	// kubernetes
	APIServerPort = 6443
)

func GetKubeConf(componetName string) string {
	return fmt.Sprintf("%s.conf", componetName)
}
