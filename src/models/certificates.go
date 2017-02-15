/*
______________________________________________________________________________

 OWLSO - Overwatch Link and Service Observer.
______________________________________________________________________________

MIT License

Copyright (c) 2014-2016 Marc Gauthier

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

______________________________________________________________________________


This file contain function to create security certificate to allow the
backend to function with HTTPS and WSS (Secure Websocket).

The certificate are created and saved in the folder ./private  and are valid
for 1 year.  If the files are deleted the OWLSO server will simply
recreate new files.

The certificate that OWLSO create are self-signed.

You can created or provide your own certificate files by simply overwritting
the existing file in the private folder.  The server expect the
following files server.key and server.crt

______________________________________________________________________________


Revision:


	01 Nov 2016 - Clean code, audit.



______________________________________________________________________________

*/

package models

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"flag"
	"fmt"
	"math/big"
	"net"
	"os"
	"time"

	"github.com/antigloss/go/logger"
)

func publicKey(priv interface{}) interface{} {

	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	default:
		return nil
	}

}

func pemBlockForKey(priv interface{}) *pem.Block {

	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(k)}
	case *ecdsa.PrivateKey:
		b, err := x509.MarshalECPrivateKey(k)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to marshal ECDSA private key: %v", err)
			os.Exit(2)
		}
		return &pem.Block{Type: "EC PRIVATE KEY", Bytes: b}
	default:
		return nil
	}

}

/*CreateCertificates this function is use to generate SSL certificates,
-Organization: OWLSO
-Not Before: Now
-Not After: Now + 1 Year
-Hostname: https://www.owlso.net
-Self Signed
Save certificate in private/server.crt
Save Private key in private/server.key
*/
func CreateCertificates() error {

	flag.Parse()

	var priv interface{}
	var err error

	priv, err = rsa.GenerateKey(rand.Reader, 2048)

	if err != nil {
		return err
	}

	var notBefore time.Time
	notBefore = time.Now()

	notAfter := notBefore.Add(365 * 24 * time.Hour)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"OWLSO"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	h := "https://www.owlso.net"

	if ip := net.ParseIP(h); ip != nil {
		template.IPAddresses = append(template.IPAddresses, ip)
	} else {
		template.DNSNames = append(template.DNSNames, h)
	}

	// whether this cert should be its own Certificate Authority
	template.IsCA = true
	template.KeyUsage |= x509.KeyUsageCertSign

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, publicKey(priv), priv)
	if err != nil {
		return err
	}

	certOut, err := os.Create("server.crt")
	if err != nil {
		return err
	}
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	certOut.Close()
	logger.Info("written server.crt")

	keyOut, err := os.OpenFile("server.key", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	pem.Encode(keyOut, pemBlockForKey(priv))
	keyOut.Close()

	logger.Info("written server.key")

	return nil
}
