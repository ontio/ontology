package link

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"testing"
)

var _tlsConfig *tls.Config

func TestTLSDial(t *testing.T) {
	CertPath := "./user1-cert.pem"
	KeyPath := "./user1-cert-key.pem"
	CAPath := "./ca.pem"

	clientCertPool := x509.NewCertPool()

	cacert, err := ioutil.ReadFile(CAPath)
	cert, err := tls.LoadX509KeyPair(CertPath, KeyPath)
	if err != nil {
		t.Error("ReadFile err:", err)
		return
	}

	ok := clientCertPool.AppendCertsFromPEM(cacert)
	if !ok {
		t.Fatalf("failed to parse root certificate")
	}

	conf := &tls.Config{
		RootCAs:      clientCertPool,
		Certificates: []tls.Certificate{cert},
		//InsecureSkipVerify: true,
	}
	println(conf)
}
