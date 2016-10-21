package main

import (
	"crypto/tls"
	"log"
	"golang.org/x/crypto/pkcs12"
	"io/ioutil"
	"encoding/pem"
	"fmt"
	"github.com/puslip41/eStreamerClient/message"
	"time"
	"io"
	"sync"
)

/*
func main1() {
	conn, err := connectServer()
	if err != nil {
		PrintError(err, "cannot connect server: %s:%s", "192.168.100.105", "8302")
	} else {
		defer conn.Close()
		log.Println("client: connected to: ", conn.RemoteAddr())

		requestMessage := getRequestMessage()


		n, err := conn.Write(requestMessage.Marshal())
		if err != nil {
			PrintError(err, "cannot send request message")
		} else {
			log.Printf("send request message: %d, % x", n, requestMessage.Marshal())
		}

		headerMessage := make([]byte, 16)
		for {
			_, err := conn.Read(headerMessage)
			if err != nil {
				if err == io.EOF {
					PrintError(err, "disconnect session")
					break;
				} else {
					PrintError(err, "read error that message")
					time.Sleep(100)
				}
			} else {
				if headerMessage == nil {
					log.Println("receive null message")
					time.Sleep(100)
				} else {
					log.Println("receive message:", headerMessage)
					time.Sleep(100)
				}
			}

			header := message.UnmarshalHeader(headerMessage)
			contentMessage := make([]byte, header.MessageLength)

			_, err = conn.Read(contentMessage)
			if err != nil {
			} else {
				//log := message.RawMessage{Header:header, Content:contentMessage}.String()
				switch header.MessageType {
				case message.NULL_MESSAGE:
					content = ""
				case message.ERROR_MESSAGE:
					content = ""
				case message.EVENT_DATA:
					content = ""
				case message.SINGLE_HOST_DATA:
					content = ""
				case message.MULTIPLE_HOST_DATA:
					content = ""
				case message.STREAMING_INFORMATION:
					content = ""
				case message.MESSAGE_BUNDLE:
				default:
					PrintError(nil, "unknown message type: %d", header.MessageType)
				}
			}
		}
	}
}
*/

var wg sync.WaitGroup
func main() {
	conn := connectServer()
	if conn != nil {
		defer conn.Close()

		wg.Add(3)

		receiveQueue := receive(conn)

		writeQueue := process(receiveQueue)

		write(writeQueue)

		wg.Wait()
	}
}

func readRawMessage(conn *tls.Conn) (*message.RawMessage, error) {
	conn.ConnectionState()
	return nil, nil
}

func receive(conn *tls.Conn) <- chan *message.RawMessage {
	c := make(chan *message.RawMessage, 1024)

	go func() {
		defer close(c)
		isConnected := true
		for {
			if isConnected {
				rawData, err := readRawMessage(conn)
				if err != nil {
					if err == io.EOF {
						isConnected = false
						conn.Close()
					} else {

					}
				} else {
					c <- rawData
				}
			} else {
				conn := connectServer()
				if conn != nil {
					isConnected = true
				} else {
					time.Sleep(1000)
				}
			}
		}
		wg.Done()
	} ()

	return c
}

func process(receiveQueue <- chan *message.RawMessage) <- chan string {
	c := make(chan string, 1024)

	go func() {
		defer close(c)
		for v := range receiveQueue {
			log.Println(v)
			c <- v.String()
			time.Sleep(1000)
		}
		wg.Done()
	} ()

	return c
}

func write(writeQueue <- chan string) {
	go func () {
		for v := range writeQueue {
			log.Printf("WRITE: %s", v)
			time.Sleep(1000)
		}
		wg.Done()
	}()
}

func PrintError(err error, format string, args ...interface{}) {
	log.Printf("ERROR: %s\n%s", fmt.Sprintf(format, args...), err.Error())
}

func connectServer() *tls.Conn {
	ip := "192.168.100.105"
	port := "8302"
	pkcs12FileName := `D:\SourceCode\Go\src\github.com\puslip41\eStreamerClient\certs\192.168.10.232.pkcs12`
	pkcs12Password := "igloosec"

	certPEMBlock, keyPEMBlock, err := readPkcs12(pkcs12FileName, pkcs12Password)
	if err != nil {
		PrintError(err, "cannot extract pem block")

		return nil
	}

	cert, err := tls.X509KeyPair(certPEMBlock, keyPEMBlock)
	if err != nil {
		PrintError(err, "cannot make certification and key")
		return nil
	}

	config := tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}

	conn, err := tls.Dial("tcp", ip + ":" + port, &config)
	if err != nil {
		PrintError(err, "cannot connect server to %s:%s", ip, port)
		return nil
	} else {
		log.Printf("INFO: success connect server to %s:%s", ip, port)
	}

	return conn
}

// pkcs12File : certs/192.168.10.232.pkcs12
// string : igloosec
func readPkcs12(pkcs12File, password string) ([]byte, []byte, error) {
	data, err := ioutil.ReadFile(pkcs12File)
	if err != nil {
		PrintError(err, "cannot read pkcs12: %s", pkcs12File)
		return nil, nil, err
	}

	blocks, err := pkcs12.ToPEM(data, password)
	if err != nil {
		PrintError(err, "cannot convert pem block")
		return nil, nil, err
	}

	var pemData []byte
	for _, b := range blocks {
		pemData = append(pemData, pem.EncodeToMemory(b)...)
	}

	return pemData, pemData, nil
}

func getRequestMessage() message.RequestMessage {
	return message.RequestMessage{
		Header:message.MessageHeader{
			HeaderVersion:1,
			MessageType:message.EVENT_STREAM_REQUEST,
			MessageLength:8,
		},
		InitialTimestamp: time.Now(),
		RequestFlags: message.BIT0 | message.BIT1 | message.BIT5,
	}
}
