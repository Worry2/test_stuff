package main

import (
	"crypto"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io/ioutil"

	"crypto/rand"
	"crypto/rsa"
	_ "crypto/sha256"
	"crypto/x509"

	"github.com/ThalesIgnite/crypto11"
)

func main() {
	ctx, err := crypto11.ConfigureFromFile("config.softhsm2")
	if err != nil {
		panic(err)
	}
	defer ctx.Close()

	id := []byte("id")
	label := []byte("label")

	s, err := ctx.FindKeyPair(id, label)
	if err != nil {
		fmt.Println(err)
	}

	if s == nil {
		sd, err := ctx.GenerateRSAKeyPairWithLabel(id, label, 2048)
		if err != nil {
			panic(err)
		}
		s = sd
	}

	if d, ok := s.(crypto.Decrypter); ok {
		fmt.Println("Public key for keypair:")
		fmt.Println(d.Public().(*rsa.PublicKey).N)

		asn1, err := x509.MarshalPKIXPublicKey(d.Public())
		if err != nil {
			panic(err)
		}

		pubBytes := pem.EncodeToMemory(&pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: asn1,
		})

		fmt.Println("Public key PEM (written in key.pub):")
		fmt.Println(string(pubBytes))

		ioutil.WriteFile("key.pub", pubBytes, 0644)
	}

	tbs := []byte("sign me with RSA")

	h := crypto.SHA256.New()
	h.Write(tbs)
	hash := h.Sum([]byte{})
	ioutil.WriteFile("tbs.txt", tbs, 0644)

	sig, err := s.Sign(rand.Reader, hash, crypto.SHA256)
	if err != nil {
		panic(fmt.Sprintf("Failed to sign: %v", err))
	}

	fmt.Println("Base64 encoded signature: ", base64.StdEncoding.EncodeToString(sig))
	ioutil.WriteFile("sig.bin", sig, 0644)

	// TODO: Verify
	fmt.Println("The End")
}
