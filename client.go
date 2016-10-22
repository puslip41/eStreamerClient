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

var wg sync.WaitGroup
func main() {
	conn := connectServer()
	if conn != nil {
		defer conn.Close()

		log.Println("client: connected to: ", conn.RemoteAddr())

		requestMessage := getRequestMessage()

		n, err := conn.Write(requestMessage.Marshal())
		if err != nil {
			PrintError(err, "cannot send request message")
		} else {
			log.Printf("send request message: %d, % x", n, requestMessage.Marshal())
		}

		wg.Add(3)

		receiveQueue := receive(conn)

		writeQueue := process(receiveQueue)

		write(writeQueue)

		wg.Wait()
	}
}

func readRawMessage(conn *tls.Conn) (message.RawMessage, error) {
	messageHeader := make([]byte, 8)

	_, err := conn.Read(messageHeader)
	if err != nil {
		PrintError(err, "cannot receive message header")
	} else {
		header := message.UnmarshalHeader(messageHeader)

		//log.Printf("Header Version : %d, Type : %d, Message Length : %d\r\n", header.HeaderVersion, header.MessageType, header.MessageLength)

		if header.MessageLength > 0 {
			messageBody := make([]byte, header.MessageLength)

			_, err = conn.Read(messageBody)
			if err != nil {
				PrintError(err, "cannot receive message body")
			} else {
				return message.RawMessage{Header:header, Content:messageBody}, nil
			}
		}
	}

	return message.RawMessage{}, nil
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
					PrintError(err, "cannot read estreamer message")
					if err == io.EOF {
						isConnected = false
						conn.Close()
					} else {

					}
				} else {
					if err = sendNullMessage(conn); err != nil {
						PrintError(err, "cannot send null message")
					}

					c <- &rawData
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

func process(receiveQueue <- chan *message.RawMessage) <- chan fmt.Stringer {
	c := make(chan fmt.Stringer, 1024)

	go func() {
		defer close(c)
		for v := range receiveQueue {
			if v.Header.MessageType == message.EVENT_DATA {
				recordHeader := message.UnmarshalRecordHeader(v.Content[:16])
				if recordHeader.RecordType == 2 {
					c <- message.UnmarshalPacketRecord(v.Header, recordHeader, v.Content[16:])
				} else  {
					log.Printf("undefined event data: RecordType:%d, RecordLength:%d, Data:% X\r\n", recordHeader.RecordType, recordHeader.RecordLength, v.Content[16:])
				}
			} else {
				log.Printf("undefine data: %d\r\n", v.Header.MessageType)
			}
		}
		wg.Done()
	} ()

	return c
}

func write(writeQueue <- chan fmt.Stringer) {
	go func () {
		for v := range writeQueue {
			//log.Printf("WRITE: %s", v.String())
			v.String()
		}
		wg.Done()
	}()
}

func PrintError(err error, format string, args ...interface{}) {
	log.Printf("ERROR: %s\n%s", fmt.Sprintf(format, args...), err.Error())
}

func connectServer() *tls.Conn {
	ip := "192.168.100.197"
	port := "8302"
	pkcs12FileName := `D:\SourceCode\Go\src\github.com\puslip41\eStreamerClient\certs\192.168.10.172.pkcs12`
	pkcs12Password := "cisco123"

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
// string : Cisco123
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
		//RequestFlags: message.BIT0 | message.BIT2 | message.BIT20 | message.BIT24 | message.BIT22 | message.BIT5 | message.BIT18 | message.BIT21 | message.BIT23,//4.9.0.x
		//RequestFlags: message.BIT0 | message.BIT2 | message.BIT20 | message.BIT25 | message.BIT22 | message.BIT5 | message.BIT26 | message.BIT21 | message.BIT23,//4.9.1.x
		//RequestFlags: message.BIT0 | message.BIT2 | message.BIT20 | message.BIT28 | message.BIT29 | message.BIT27 | message.BIT5 | message.BIT26 | message.BIT21 | message.BIT23,//4.10.x
		//RequestFlags: message.BIT0 | message.BIT2 | message.BIT20 | message.BIT30 | message.BIT30 | message.BIT27 | message.BIT5 | message.BIT30 | message.BIT30 | message.BIT23, //5.0+, 5.1
		//RequestFlags: message.BIT0 | message.BIT30 | message.BIT20 | message.BIT30 | message.BIT30 | message.BIT27 | message.BIT5 | message.BIT30 | message.BIT30 | message.BIT30 | message.BIT30 | message.BIT23, //5.1.1+
		RequestFlags: message.BIT0 | message.BIT2 | message.BIT23, //5.1.1+
	}
}

func sendNullMessage(conn *tls.Conn) error {
	nullMessage := message.MessageHeader{HeaderVersion:1, MessageType:0, MessageLength:0}

	_, err := conn.Write(nullMessage.Marshal())

	return err
}
