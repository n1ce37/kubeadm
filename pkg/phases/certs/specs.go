package certs

import (
	"crypto/x509"
	"net"

	certutil "k8s.io/client-go/util/cert"
)

type certSpec struct {
	name   string
	config certutil.Config
}

type certGroupSpec struct {
	ca       certSpec
	subCerts []certSpec
}

func getCertGroupSpecList(cfg Config) ([]certGroupSpec, error) {
	apiserverAltNames, err := getAPIServerAltNames(cfg)
	if err != nil {
		return nil, err
	}
	etcdAltNames, err := getEtcdAltNames(cfg)
	if err != nil {
		return nil, err
	}

	return []certGroupSpec{
		// common kube
		{
			ca: certSpec{
				name: "ca",
				config: certutil.Config{
					CommonName: "kubernetes",
				},
			},
			subCerts: []certSpec{
				{
					name: "apiserver",
					config: certutil.Config{
						CommonName: "kube-apiserver",
						Usages:     []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
						AltNames:   apiserverAltNames,
					},
				},
				{
					name: "apiserver-kubelet-client",
					config: certutil.Config{
						CommonName:   "kube-apiserver-kubelet-client",
						Organization: []string{"system:masters"},
						Usages:       []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
					},
				},
			},
		},

		// front-proxy
		{
			ca: certSpec{
				name: "front-proxy-ca",
				config: certutil.Config{
					CommonName: "front-proxy-ca",
				},
			},
			subCerts: []certSpec{
				{
					name: "front-proxy-client",
					config: certutil.Config{
						CommonName: "front-proxy-client",
						Usages:     []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
					},
				},
			},
		},

		// etcd
		{
			ca: certSpec{
				name: "etcd-ca",
				config: certutil.Config{
					CommonName: "etcd-ca",
				},
			},
			subCerts: []certSpec{
				{
					name: "etcd-server",
					config: certutil.Config{
						Usages:   []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
						AltNames: etcdAltNames,
					},
				},
				{
					name: "etcd-peer",
					config: certutil.Config{
						Usages: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
					},
				},
				{
					name: "etcd-healthcheck-client",
					config: certutil.Config{
						CommonName:   "kube-etcd-healthcheck-client",
						Organization: []string{"system:masters"},
						Usages:       []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
					},
				},
				{
					name: "apiserver-etcd-client",
					config: certutil.Config{
						CommonName:   "kube-apiserver-etcd-client",
						Organization: []string{"system:masters"},
						Usages:       []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
					},
				},
			},
		},
	}, nil
}

func getAPIServerAltNames(cfg Config) (certutil.AltNames, error) {
	//_, svcSubnet, err := net.ParseCIDR("")
	//if err != nil {
	//	return nil, err
	//}

	//internalAPIServerVirtualIP, err := ipallocator.GetIndexedIP(svcSubnet, 1)
	//if err != nil {
	//	return nil, err
	//}

	altNames := certutil.AltNames{
		DNSNames: []string{
			"kubernetes",
			"kubernetes.default",
			"kubernetes.default.svc",
			// TODO
			//fmt.Sprintf("kubernetes.default.svc.%s", )
		},
		IPs: []net.IP{
			// TODO
			//internalAPIServerVirtualIP,
			cfg.InternalAdvertiseAddress,
			cfg.ExternalAdvertiseAddress,
		},
	}

	return altNames, nil
}

// TODO
func getEtcdAltNames(cfg Config) (certutil.AltNames, error) {
	ips := make([]net.IP, 0, len(cfg.Etcds))
	return certutil.AltNames{
		IPs: ips,
	}, nil
}

// TODO
func getEtcdPeerAltNames() {

}
