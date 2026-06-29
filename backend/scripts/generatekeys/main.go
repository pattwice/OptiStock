package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	keyDir := "keys"
	if len(os.Args) > 1 {
		keyDir = os.Args[1]
	}
	if err := os.MkdirAll(keyDir, 0o755); err != nil {
		panic(err)
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	privatePath := filepath.Join(keyDir, "private.pem")
	publicPath := filepath.Join(keyDir, "public.pem")

	if err := writePrivateKey(privatePath, privateKey); err != nil {
		panic(err)
	}
	if err := writePublicKey(publicPath, &privateKey.PublicKey); err != nil {
		panic(err)
	}

	fmt.Printf("JWT keys written to %s\n", keyDir)
}

func writePrivateKey(path string, key *rsa.PrivateKey) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return pem.Encode(file, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
}

func writePublicKey(path string, key *rsa.PublicKey) error {
	der, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return pem.Encode(file, &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: der,
	})
}
