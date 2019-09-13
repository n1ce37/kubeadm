package certs

import (
	"crypto"
	"crypto/rsa"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math"
	"math/big"
	"net"
	"time"

	certutil "k8s.io/client-go/util/cert"
)

const (
	rsaPrivateKeyBlockType = "RSA PRIVATE KEY"
  certificateBlockType = "CERTIFICATE"

	rsaKeySize = 2048
)

type Config struct {
	InternalAdvertiseAddress net.IP
	ExternalAdvertiseAddress net.IP

	Etcds map[string]net.IP

	SvcNet net.IPNet
}

// TODO more log and config struct define
func CreatePKIAssets(cfg Config) (map[string][]byte, error) {
	certGroupSpecList, err := getCertGroupSpecList(cfg)
	if err != nil {
		return nil, err
	}

	pkis := map[string][]byte{}
	for _, vcertGroupSpec := range certGroupSpecList {
		caKey, caCert, err := newCaPrivateKeyAndCert(vcertGroupSpec.ca.config)
		if err != nil {
			return nil, err
		}
		caKeyBytes, caCertBytes, err := encodeKeyAndCertPEM(caKey, caCert)
		if err != nil {
			return nil, err
		}
		// TODO name change
		pkis[vcertGroupSpec.ca.name] = caKeyBytes
		pkis[vcertGroupSpec.ca.name] = caCertBytes

		for _, vcert := range vcertGroupSpec.subCerts {
			key, cert, err := newPrivateKeyAndCert(vcert.config, caKey, caCert)
			if err != nil {
				return nil, err
			}
			keyBytes, certBytes, err := encodeKeyAndCertPEM(key, cert)
			if err != nil {
				return nil, err
			}
			pkis[vcert.name] = keyBytes
			pkis[vcert.name] = certBytes
		}
	}

	return pkis, nil
}

func newPrivateKey() (crypto.Signer, error){
	return rsa.GenerateKey(rand.Reader, rsaKeySize)
}

func newCaCert(cfg certutil.Config, key crypto.Signer) (*x509.Certificate, error) {
	return certutil.NewSelfSignedCACert(cfg, key)
}

func newCert(cfg certutil.Config, key crypto.Signer, caCert *x509.Certificate, caKey crypto.Signer) (*x509.Certificate, error) {
	serial, err := rand.Int(rand.Reader, new(big.Int).SetInt64(math.MaxInt64))
	if err != nil {
		return nil, err
	}

	certTmpl := x509.Certificate{
		Subject: pkix.Name{
			CommonName: cfg.CommonName,
			Organization: cfg.Organization,
		},
		DNSNames: cfg.AltNames.DNSNames,
		IPAddresses: cfg.AltNames.IPs,
		SerialNumber: serial,
		NotBefore: caCert.NotBefore,
		// TODO time limit
		NotAfter: time.Now().Add(368 * 100 * time.Hour).UTC(),
		KeyUsage: x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: cfg.Usages,
	}
	certDERBytes, err := x509.CreateCertificate(rand.Reader, &certTmpl, caCert, key.Public(), caKey)
	if err != nil {
		return nil, err
	}
	return x509.ParseCertificate(certDERBytes)
}

func newCaPrivateKeyAndCert(cfg certutil.Config) ( crypto.Signer, *x509.Certificate,error) {
	key, err := newPrivateKey()
	if err != nil {
		return nil, nil, err
	}
	cert, err := newCaCert(cfg, key)
	if err != nil {
		return nil, nil, err
	}

	return key, cert, nil
}

func newPrivateKeyAndCert(cfg certutil.Config, caKey crypto.Signer,  caCert *x509.Certificate) ( crypto.Signer,  *x509.Certificate, error) {
	key, err := newPrivateKey()
	if err != nil {
		return nil, nil, err
	}
	cert, err := newCert(cfg, key, caCert, caKey)
	if err != nil {
		return nil, nil, err
	}

	return  key, cert, nil
}

func encodeKeyAndCertPEM(key crypto.Signer, cert *x509.Certificate) ([]byte, []byte, error) {
	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		// TODO more detail
		return nil, nil, errors.New("not rsa private key")
	}
	keyBlock := &pem.Block{
		Type: rsaPrivateKeyBlockType,
		Bytes: x509.MarshalPKCS1PrivateKey(rsaKey),
	}
	keyByte := pem.EncodeToMemory(keyBlock)

	certBlock := &pem.Block{
		Type: certificateBlockType,
		Bytes: cert.Raw,
	}
	certByte := pem.EncodeToMemory(certBlock)

	return keyByte, certByte, nil
}