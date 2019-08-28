package main

import (
	"crypto"
	"encoding/base64"
	"fmt"

	"crypto/rand"
	"crypto/rsa"
	_ "crypto/sha256"

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
	}

	plaintext := []byte("sign me with RSA")
	h := crypto.SHA256.New()
	h.Write(plaintext)
	hash := h.Sum([]byte{})

	sig, err := s.Sign(rand.Reader, hash, crypto.SHA256)
	if err != nil {
		panic(fmt.Sprintf("Failed to sign: %v", err))
	}

	fmt.Println("Signature: ", base64.StdEncoding.EncodeToString(sig))

	// TODO: Verify
	fmt.Println("The End")
}
