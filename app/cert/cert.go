package cert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"golang.org/x/net/idna"
	"math/big"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	certDir       = "certs"
	pathSeparator = string(os.PathSeparator)
)

var mu sync.Mutex

func GetCert(serverName string) (*tls.Certificate, error) {
	cert, key, err := CreateKeyPair(serverName)
	if err != nil {
		return nil, err
	}

	var tlsCert tls.Certificate
	if tlsCert, err = tls.LoadX509KeyPair(cert, key); err != nil {
		return nil, err
	}

	return &tlsCert, nil
}

func CreateKeyPair(commonName string) (certFile, keyFile string, err error) {
	mu.Lock()
	defer mu.Unlock()

	commonName, err = idna.ToASCII(commonName)
	if err != nil {
		return "", "", err
	}
	commonName = strings.ToLower(commonName)

	destDir := certDir + pathSeparator + commonName + pathSeparator

	certFile = destDir + "cert.pem"
	keyFile = destDir + "key.pem"

	if _, err = tls.LoadX509KeyPair(certFile, keyFile); err == nil {
		return certFile, keyFile, nil
	}

	lastWeek := time.Now().AddDate(0, 0, -7)
	notBefore := lastWeek
	notAfter := lastWeek.AddDate(2, 0, 0)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return "", "", nil
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization:       []string{"mitm-proxy certificate"},
			OrganizationalUnit: []string{"mitm-proxy certificate"},
			CommonName:         commonName,
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  false,
	}

	if ip := net.ParseIP(commonName); ip != nil {
		template.IPAddresses = append(template.IPAddresses, ip)
	} else {
		template.DNSNames = append(template.DNSNames, commonName)
	}

	rootCA, err := tls.LoadX509KeyPair("ca.crt", "ca.key")
	if err != nil {
		return "", "", err
	}

	if rootCA.Leaf, err = x509.ParseCertificate(rootCA.Certificate[0]); err != nil {
		return "", "", err
	}

	template.AuthorityKeyId = rootCA.Leaf.SubjectKeyId

	var key *rsa.PrivateKey
	if key, err = rsa.GenerateKey(rand.Reader, 2048); err != nil {
		return "", "", err
	}
	template.SubjectKeyId = func(n *big.Int) []byte {
		h := sha1.New()
		h.Write(n.Bytes())
		return h.Sum(nil)
	}(key.N)

	var bytes []byte
	if bytes, err = x509.CreateCertificate(rand.Reader, &template, rootCA.Leaf, &key.PublicKey, rootCA.PrivateKey); err != nil {
		return "", "", err
	}

	if err = os.MkdirAll(destDir, 0755); err != nil {
		return "", "", err
	}

	certOut, err := os.Create(certFile)
	if err != nil {
		return "", "", err
	}
	defer certOut.Close()

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: bytes}); err != nil {
		return "", "", err
	}

	keyOut, err := os.OpenFile(keyFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return "", "", err
	}
	defer keyOut.Close()

	if err := pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}); err != nil {
		return "", "", nil
	}

	return certFile, keyFile, nil
}
