package main

import (
	"crypto/tls"
	"log"
	"golang.org/x/crypto/pkcs12"
	"io/ioutil"
	"encoding/pem"
	"encoding/binary"
	"fmt"
)

type RequestMessage struct {
	HeaderVersion int16
	MessageType int16
	MessageLength int32
	InitialTimestamp int32
	RequestFlags int32
}

func ConvertSendBytes(message RequestMessage) []byte {
	bytes := make([]byte, 0, 16)

	byteValue := make([]byte, 2)
	binary.BigEndian.PutUint16(byteValue, uint16(message.HeaderVersion))
	fmt.Println("%d -> %x", message.HeaderVersion, byteValue)
	bytes = append(bytes, byteValue[0], byteValue[1])

	byteValue = make([]byte, 2)
	binary.BigEndian.PutUint16(byteValue, uint16(message.MessageType))
	fmt.Println("%d -> %x", message.MessageType, byteValue)
	bytes = append(bytes, byteValue[0], byteValue[1])

	byteValue = make([]byte, 4)
	binary.BigEndian.PutUint32(byteValue, uint32(message.MessageLength)) 
	fmt.Println("%d -> %x", message.MessageLength, byteValue)
	bytes = append(bytes, byteValue[0], byteValue[1], byteValue[2], byteValue[3])

	byteValue = make([]byte, 4)
	binary.BigEndian.PutUint32(byteValue, uint32(message.InitialTimestamp)) 
	fmt.Println("%d -> %x", message.InitialTimestamp, byteValue)
	bytes = append(bytes, byteValue[0], byteValue[1], byteValue[2], byteValue[3])

	byteValue = make([]byte, 4)
	binary.BigEndian.PutUint32(byteValue, uint32(message.RequestFlags)) 
	fmt.Println("%d -> %x", message.RequestFlags, byteValue)
	bytes = append(bytes, byteValue[0], byteValue[1], byteValue[2], byteValue[3])

	return bytes
}

func main() {

	msg := RequestMessage{HeaderVersion: 1, MessageType: 2, MessageLength: 4, InitialTimestamp: 8, RequestFlags: 16}

	requestMessage := ConvertSendBytes(msg)

	log.Printf("cliend: send: %x", requestMessage) 

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

	n, err := conn.Write(requestMessage)
	if err != nil {
		log.Fatalf("client: write: %s", err)
	}
	log.Printf("client: wrote %q (%d bytes)", requestMessage, n)

	reply := make([]byte, 256)
	n, err = conn.Read(reply)
	log.Printf("client: read %q (%d bytes)", string(reply[:n]), n)
	
	log.Print("Client: exiting")
}
