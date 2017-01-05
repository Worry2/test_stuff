package main

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"flag"
	"fmt"
	"log"
	"math/big"
	"os"
	"strconv"
	"time"

	"github.com/tahkapaa/test_stuff/randata"
)

var (
	host         = flag.String("host", "", "Comma-separated hostnames and IPs to generate a certificate for")
	validFrom    = flag.String("start-date", "", "Creation date formatted as Jan 1 15:04:05 2011")
	validFor     = flag.Duration("duration", 365*24*time.Hour, "Duration that certificate is valid for")
	isCA         = flag.Bool("ca", false, "whether this cert should be its own Certificate Authority")
	rsaBits      = flag.Int("rsa-bits", 2048, "Size of RSA key to generate. Ignored if --ecdsa-curve is set")
	ecdsaCurve   = flag.String("ecdsa-curve", "", "ECDSA curve to use to generate a key. Valid values are P224, P256, P384, P521")
	serialNumber = int64(100000)

	notBefore time.Time
	notAfter  time.Time
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

func getKeyID(bytes []byte) []byte {
	h := sha1.New()
	h.Write(bytes)
	return h.Sum(nil)
}

func createRandomCertificate(parent *x509.Certificate, serialNumber int64) (cert []byte, err2 error) {

	var priv interface{}
	var err error
	priv, err = rsa.GenerateKey(rand.Reader, *rsaBits)
	if err != nil {
		log.Fatalf("failed to generate private key: %s", err)
	}

	serial := big.NewInt(serialNumber)
	template := x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			Organization: []string{randata.GetRandomWord()},
			Country:      []string{randata.GetRandomCountry()},
			CommonName:   randata.GetThreeWords(),
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		SubjectKeyId:          getKeyID(pemBlockForKey(priv).Bytes),
	}

	if parent == nil {
		template.IsCA = true
		template.KeyUsage |= x509.KeyUsageCertSign
		template.AuthorityKeyId = getKeyID(pemBlockForKey(priv).Bytes)
		parent = &template
	} else {
		template.AuthorityKeyId = parent.SubjectKeyId
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, parent, publicKey(priv), priv)
	if err != nil {
		log.Fatalf("Failed to create certificate: %s", err)
	}

	fileName := strconv.FormatInt(serialNumber, 10) + ".pem"
	certOut, err := os.Create("certs/" + fileName)
	if err != nil {
		log.Fatalf(fmt.Sprintf("failed to open %s for writing: %s", fileName, err))
	}
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	certOut.Close()

	keyFilename := strconv.FormatInt(serialNumber, 10) + "_key.pem"
	keyOut, err := os.OpenFile("keys/"+keyFilename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Print(fmt.Sprintf("failed to open %s for writing: %s", fileName, err))
		return
	}

	pem.Encode(keyOut, pemBlockForKey(priv))
	keyOut.Close()
	//log.Print(fmt.Sprintf("written %s\n", keyFilename))

	return derBytes, nil
}

func createOneRandomCertificate(caCert *x509.Certificate, serial int64) <-chan string {
	c := make(chan string)
	go func() {
		certBytes, err := createRandomCertificate(caCert, serial)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create certificate: %s\n", err)
			os.Exit(1)
		}
		cert, err := x509.ParseCertificate(certBytes)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse certificate: %s\n", err)
			os.Exit(1)
		}
		sendCertificate(cert)
		c <- fmt.Sprintf("Created certificate: %d", serial)
	}()
	return c
}

func sendCertificate(cert *x509.Certificate) {
	fmt.Printf("CN: %s\n", cert.Subject.CommonName)
	fmt.Printf("C: %s\n", cert.Subject.Country[0])
	fmt.Printf("O: %s\n", cert.Subject.Organization[0])
	fmt.Printf("Authority Key ID: % x\n", cert.AuthorityKeyId)
	fmt.Printf("Subject Key ID: % x\n", cert.SubjectKeyId)
}

func createNRandomCertificates(caCert *x509.Certificate, startSerial int64, n int64) <-chan string {

	c := make(chan string)
	go func() {
		for i := startSerial; i < startSerial+n; i++ {
			createRandomCertificate(caCert, i)
			c <- fmt.Sprintf("Created certificate: %d", i)
		}
	}()
	return c
}

func fanIn(input1, input2 <-chan string) <-chan string {
	c := make(chan string)
	go func() {
		for {
			select {
			case s := <-input1:
				c <- s
			case s := <-input2:
				c <- s
			}
		}
	}()
	return c
}

func main() {
	flag.Parse()

	os.Mkdir("certs", 0755)
	os.Mkdir("keys", 0755)

	var err error
	if len(*validFrom) == 0 {
		notBefore = time.Now()
	} else {
		notBefore, err = time.Parse("Jan 2 15:04:05 2006", *validFrom)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse creation date: %s\n", err)
			os.Exit(1)
		}
	}

	randata.Initialize()
	notAfter = notBefore.Add(*validFor)

	caCertData, err := createRandomCertificate(nil, serialNumber)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create ca certificate: %s\n", err)
		os.Exit(1)
	}
	caCert, err := x509.ParseCertificate(caCertData)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse ca certificate: %s\n", err)
		os.Exit(1)
	}
	serialNumber++

	c1 := createOneRandomCertificate(caCert, serialNumber)
	serialNumber++
	c2 := createOneRandomCertificate(caCert, serialNumber)
	serialNumber++
	c3 := createOneRandomCertificate(caCert, serialNumber)
	serialNumber++
	c4 := createOneRandomCertificate(caCert, serialNumber)
	serialNumber++
	/*	c5 := createOneRandomCertificate(caCert, serialNumber)
		serialNumber +=1
		c6 := createOneRandomCertificate(caCert, serialNumber)
		serialNumber +=1
	*/
	timeOut := time.After(2 * time.Second)

	for {
		select {
		case s := <-c1:
			fmt.Println(s)
			c1 = createOneRandomCertificate(caCert, serialNumber)
			serialNumber++
		case s := <-c2:
			fmt.Println(s)
			c2 = createOneRandomCertificate(caCert, serialNumber)
			serialNumber++
		case s := <-c3:
			fmt.Println(s)
			c3 = createOneRandomCertificate(caCert, serialNumber)
			serialNumber++
		case s := <-c4:
			fmt.Println(s)
			c4 = createOneRandomCertificate(caCert, serialNumber)
			serialNumber++
			/*			case s := <- c5:
							fmt.Println(s)
							c5 = createOneRandomCertificate(caCert, serialNumber)
							serialNumber +=1
						case s := <- c6:
							fmt.Println(s)
							c6 = createOneRandomCertificate(caCert, serialNumber)
							serialNumber +=1*/
		case <-timeOut:
			fmt.Println("Aika loppui")
			return
		}
	}
	/*
		c := fanIn(createNRandomCertificates(caCert, serialNumber, 1000),
				createNRandomCertificates(caCert, serialNumber+10000, 1000))
		for i := 0; i < 1000; i++ {
			fmt.Printf("MSG: %s\n", <-c)
		}
	*/

}
