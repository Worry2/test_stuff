package main

import (
	"fmt"

	_ "crypto/sha256"

	"github.com/ThalesIgnite/crypto11"
)

func main() {
	_, err := crypto11.ConfigureFromFile("config.softhsm2")
	if err != nil {
		panic(err)
	}

	// pvk, err := crypto11.FindKeyPairOnSlot(0, nil, []byte("TestToken"))
	// if err != nil {
	// 	panic(err)
	// }

	// signingKey, ok := pvk.(*crypto11.PKCS11PrivateKeyRSA)
	// if !ok {
	// 	panic("Failed to type assert to *crypto11.PKCS11PrivateKeyRSA")
	// }

	// id, label, err := signingKey.Identify()
	// if err != nil {
	// 	panic(fmt.Sprintf("Failed to identify key: %v", err))
	// }
	// fmt.Printf("Key Label: '%s', ID: '%d'\n", label, id)

	// if err := signingKey.Validate(); err != nil {
	// 	panic(fmt.Sprintf("Failed to validate key: %v", err))
	// }

	// plaintext := []byte("sign me with RSA")
	// h := crypto.SHA256.New()
	// h.Write(plaintext)
	// hash := h.Sum([]byte{})

	// sig, err := signingKey.Sign(rand.Reader, hash, crypto.SHA256)
	// if err != nil {
	// 	panic(fmt.Sprintf("Failed to sign: %v", err))
	// }

	// fmt.Println("Signature: ", base64.StdEncoding.EncodeToString(sig))

	// TODO: Verify
	fmt.Println("The End")
}
