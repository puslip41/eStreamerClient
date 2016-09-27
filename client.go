package main

import (
	"crypto/tls"
	"log"
	"golang.org/x/crypto/pkcs12"
	"io"
	"io/ioutil"
	"encoding/pem"
)

func main() {

	log.Println("client: start")
	
	data, err := ioutil.ReadFile("certs/192.168.10.232.pkcs12")
	if err != nil {
		log.Fatalf("client: readfile: %s", err)
	}

	log.Println("client: success to readfile")

	blocks, err := pkcs12.ToPEM(data, "igloosec")
	if err != nil {
		log.Fatalf("client: topem: %s", err)
	}

	var pemData []byte
	for _, b := range blocks {
		pemData = append(pemData, pem.EncodeToMemory(b)...)
	}
	
	cert, err := tls.X509KeyPair(pemData, pemData)
	if err != nil {
		log.Fatalf("client: x509keypair: %s", err)
	}


	log.Println("client: success decode pkcs12")
	config := tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}
	conn, err := tls.Dial("tcp", "192.168.100.105:8302", &config)
	if err != nil {
		log.Fatalf("client: dial: %s", err)
	}

	defer conn.Close()
	log.Println("client: connected to: ", conn.RemoteAddr())

	state := conn.ConnectionState()
	log.Println("client: handshake: ", state.HandshakeComplete)

	message := "Hello\n"
	n, err := io.WriteString(conn, message)
	if err != nil {
		log.Fatalf("client: write: %s", err)
	}
	log.Printf("client: wrote %q (%d bytes)", message, n)

	reply := make([]byte, 256)
	n, err = conn.Read(reply)
	log.Printf("client: read %q (%d bytes)", string(reply[:n]), n)
	
	log.Print("Client: exiting")
}
