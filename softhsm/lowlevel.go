package main

import (
	"encoding/base64"
	"fmt"
	"os"

	"github.com/miekg/pkcs11"
)

// Rename to main...
func lowmain() {
	p := getCtx()
	session := getSession(p)
	defer finishSession(p, session)

	digest(p, session)

	pbk, pvk := generateRSAKeyPair(p, session, "TestToken2", true)

	fmt.Println(pvk)
	fmt.Println()
	fmt.Println(pbk)

	signature := sign(p, session, pvk)

	fmt.Println(base64.StdEncoding.EncodeToString(signature))
}

func sign(p *pkcs11.Ctx, session pkcs11.SessionHandle, pvk pkcs11.ObjectHandle) []byte {
	err := p.SignInit(session, []*pkcs11.Mechanism{pkcs11.NewMechanism(pkcs11.CKM_SHA256_RSA_PKCS, nil)}, pvk)
	if err != nil {

	}
	sig, e := p.Sign(session, []byte("Sign me!"))
	if e != nil {
		panic(fmt.Sprintf("failed to sign: %s\n", e))
	}
	return sig
}

func digest(p *pkcs11.Ctx, session pkcs11.SessionHandle) {
	err := p.DigestInit(session, []*pkcs11.Mechanism{pkcs11.NewMechanism(pkcs11.CKM_SHA_1, nil)})
	if err != nil {
		panic(err)
	}
	hash, err := p.Digest(session, []byte("this is a string"))
	if err != nil {
		panic(err)
	}

	for _, d := range hash {
		fmt.Printf("%x", d)
	}
	fmt.Println()
}

func getCtx() *pkcs11.Ctx {
	lib := "/usr/lib64/libsofthsm2.so"
	if x := os.Getenv("SOFTHSM_LIB"); x != "" {
		lib = x
	}
	p := pkcs11.New(lib)
	if p == nil {
		panic("Failed to initialize pkcs11")
	}
	return p
}

func getSession(p *pkcs11.Ctx) pkcs11.SessionHandle {
	err := p.Initialize()
	if err != nil {
		panic(err)
	}

	slots, err := p.GetSlotList(true)
	if err != nil {
		panic(err)
	}

	session, err := p.OpenSession(slots[0], pkcs11.CKF_SERIAL_SESSION|pkcs11.CKF_RW_SESSION)
	if err != nil {
		panic(err)
	}

	err = p.Login(session, pkcs11.CKU_USER, "1234")
	if err != nil {
		panic(err)
	}

	return session
}

func finishSession(p *pkcs11.Ctx, session pkcs11.SessionHandle) {
	p.Logout(session)
	p.CloseSession(session)
	p.Finalize()
	p.Destroy()
}

/*
Purpose: Generate RSA keypair with a given name and persistence.
Inputs: test object
	context
	session handle
	tokenLabel: string to set as the token labels
	tokenPersistent: boolean. Whether or not the token should be
			session based or persistent. If false, the
			token will not be saved in the HSM and is
			destroyed upon termination of the session.
Outputs: creates persistent or ephemeral tokens within the HSM.
Returns: object handles for public and private keys. Fatal on error.
*/
func generateRSAKeyPair(p *pkcs11.Ctx, session pkcs11.SessionHandle, tokenLabel string, tokenPersistent bool) (pkcs11.ObjectHandle, pkcs11.ObjectHandle) {
	/*
		inputs: test object, context, session handle
			tokenLabel: string to set as the token labels
			tokenPersistent: boolean. Whether or not the token should be
					session based or persistent. If false, the
					token will not be saved in the HSM and is
					destroyed upon termination of the session.
		outputs: creates persistent or ephemeral tokens within the HSM.
		returns: object handles for public and private keys.
	*/

	publicKeyTemplate := []*pkcs11.Attribute{
		pkcs11.NewAttribute(pkcs11.CKA_CLASS, pkcs11.CKO_PUBLIC_KEY),
		pkcs11.NewAttribute(pkcs11.CKA_KEY_TYPE, pkcs11.CKK_RSA),
		pkcs11.NewAttribute(pkcs11.CKA_TOKEN, tokenPersistent),
		pkcs11.NewAttribute(pkcs11.CKA_VERIFY, true),
		pkcs11.NewAttribute(pkcs11.CKA_PUBLIC_EXPONENT, []byte{1, 0, 1}),
		pkcs11.NewAttribute(pkcs11.CKA_MODULUS_BITS, 2048),
		pkcs11.NewAttribute(pkcs11.CKA_LABEL, tokenLabel),
	}
	privateKeyTemplate := []*pkcs11.Attribute{
		pkcs11.NewAttribute(pkcs11.CKA_TOKEN, tokenPersistent),
		pkcs11.NewAttribute(pkcs11.CKA_SIGN, true),
		pkcs11.NewAttribute(pkcs11.CKA_LABEL, tokenLabel),
		pkcs11.NewAttribute(pkcs11.CKA_SENSITIVE, true),
		pkcs11.NewAttribute(pkcs11.CKA_EXTRACTABLE, true),
	}
	pbk, pvk, e := p.GenerateKeyPair(session,
		[]*pkcs11.Mechanism{pkcs11.NewMechanism(pkcs11.CKM_RSA_PKCS_KEY_PAIR_GEN, nil)},
		publicKeyTemplate, privateKeyTemplate)
	if e != nil {
		panic(fmt.Sprintf("failed to generate keypair: %s\n", e))
	}

	return pbk, pvk
}
