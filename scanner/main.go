package main

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"

	"github.com/go-ini/ini"
)

var (
	nodeid string
	salt   string
	key    string
	server string
)

func main() {
	cfg, err := ini.Load("./city-scanner.config")
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		os.Exit(1)
	}

	nodeid = cfg.Section("Node").Key("id").String()
	salt = cfg.Section("Security").Key("salt").String()
	server = cfg.Section("Server").Key("server_url").String()
	keyString := cfg.Section("Security").Key("key").String()

	block, _ := pem.Decode([]byte(keyString))
	if block == nil {
		fmt.Println("failed to parse PEM block containing the key")
	}

	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		fmt.Println(err)
	}

	/*
		s := strings.Split(string(e.PublicKeyPEM), "\n")
		for i, st := range s {
			fmt.Println("Key"+strconv.Itoa(i)+"=", st)
		}
		var newKey = strings.Join(s, "\n")
		fmt.Println(newKey)

		block, _ := pem.Decode([]byte(newKey))

		if block == nil || block.Type != "RSA PUBLIC KEY" {

			log.Fatal("failed to decode PEM block containing public key")

		}

		pub, err := x509.ParsePKIXPublicKey(block.Bytes)

		if err != nil {

			log.Fatal(err)

		}

		fmt.Println(pub)
	*/

	fmt.Println(publicKey)

	/*
		proximity.Run()
		for {

		}*/
}
