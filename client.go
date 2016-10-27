package main

import (
	"crypto/tls"
	"golang.org/x/crypto/pkcs12"
	"io/ioutil"
	"encoding/pem"
	"fmt"
	"time"
	"io"
	"sync"
	"os"

	"github.com/puslip41/eStreamerClient/message"
	"github.com/puslip41/eStreamerClient/configuration"
)

const FILE_END_FLAG = "END_FILE"
const TIME_HOUR_PATTERN = "2006010215"

var wg sync.WaitGroup

func main() {
	config, err := configuration.ReadConfiguration()
	if err != nil {
		fmt.Println("cannot read config file:", err)
		os.Exit(-1)
	}

	initializeDirectories(config)

	configuration.InitializeLogger(config.LogLevel)
	defer configuration.CloseLogger()

	conn := connectServer(config)
	if conn != nil {
		defer conn.Close()

		configuration.WriteInfo("success connect server to %s", conn.RemoteAddr())

		sendRequestMessage(conn)

		wg.Add(3)

		receiveQueue := receive(conn, config)

		writeQueue := process(receiveQueue)

		write(config.ExportDirectory, writeQueue)

		wg.Wait()
	}
}


func IsExistDirectory(directory string) bool {
	_, err := os.Open(directory)
	if err != nil {
		return false
	}

	return true
}

func initializeDirectories(config configuration.Configuration) {
	if IsExistDirectory(config.ExportDirectory) == false {
		os.MkdirAll(config.ExportDirectory, os.FileMode(644))
	}
}

func readRawMessage(conn *tls.Conn) (message.RawMessage, error) {
	messageHeader := make([]byte, 8)

	_, err := conn.Read(messageHeader)
	if err != nil {
		configuration.WriteError(err, "cannot receive message header")
		return message.RawMessage{}, err
	} else {
		header := message.UnmarshalHeader(messageHeader)

		if header.MessageLength > 0 {
			messageBody := make([]byte, header.MessageLength)

			_, err = conn.Read(messageBody)
			if err != nil {
				configuration.WriteError(err, "cannot receive message body")
				return message.RawMessage{}, err
			} else {
				return message.RawMessage{Header:header, Content:messageBody}, nil
			}
		}
	}

	return message.RawMessage{}, nil
}

func sendRequestMessage(conn *tls.Conn) error {
	requestMessage := getRequestMessage()
	_, err := conn.Write(requestMessage.Marshal())
	if err != nil {
		configuration.WriteError(err, "cannot send request message")
	} else {
		configuration.WriteInfo("send request message")
	}

	return err
}

func receive(conn *tls.Conn, config configuration.Configuration) <- chan *message.RawMessage {
	c := make(chan *message.RawMessage, 1024)

	go func() {
		defer close(c)
		isConnected := true
		var failedCount time.Duration
		for {
			if isConnected {
				rawData, err := readRawMessage(conn)
				if err != nil {
					failedCount++
					if err == io.EOF {
						isConnected = false
						conn.Close()
						configuration.WriteInfo("disconnect server")
						time.Sleep(1000*1000*1000 * failedCount)
					} else {
						configuration.WriteError(err, "cannot read estreamer message")
					}
				} else {
					if err = sendNullMessage(conn); err != nil {
						configuration.WriteError(err, "cannot send null message")
					}
					failedCount = 0

					c <- &rawData
				}
			} else {
				conn := connectServer(config)
				if conn != nil {
					isConnected = true
					configuration.WriteInfo("reconnected to server: %s:%d", config.ServerIP, config.ServiecePort)
					sendRequestMessage(conn)
				} else {
					failedCount++
					configuration.WriteError(nil, "cannot reconnect to server: %s:%d", config.ServerIP, config.ServiecePort)
					time.Sleep(1000*1000*1000 * failedCount)
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
			if v.Header.MessageType == message.EVENT_DATA {
				recordHeader := message.UnmarshalRecordHeader(v.Content[:16])
				if recordHeader.RecordType == 2 {
					record := message.UnmarshalPacketRecord(v.Header, recordHeader, v.Content[16:])

					intrusionEvent := message.GetIntrusionEvent(record.DeviceID, record.EventID)
					if intrusionEvent != nil {
						c <- fmt.Sprintf("%s packet_size=%d packet_data=%X", intrusionEvent.String(), record.PacketLength, record.PacketData)
					} else {
						configuration.WriteError(nil, "cannot find intrusion event record: device id:%d, event id:%d", record.DeviceID, record.EventID)
					}

				} else if recordHeader.RecordType == 104 {
					v := message.UnmarshalIntrusionEventRecordIPv4(v.Header, recordHeader, v.Content[16:])
					message.PutIntrusionEvent(v.DeviceID, v.EventID, v)
				} else if recordHeader.RecordType == 105 {
					v := message.UnmarshalIntrusionEventRecordIPv6(v.Header, recordHeader, v.Content[16:])
					message.PutIntrusionEvent(v.DeviceID, v.EventID, v)
				} else {
					configuration.WriteWarning("undefined event data: RecordType:%3d, RecordLength:%3d, Data:% X\r\n", recordHeader.RecordType, recordHeader.RecordLength, v.Content[16:])
				}

				configuration.WriteDebug("### Record Storage Count: %d", message.GetCountInRecordStroage())
			} else {
				configuration.WriteWarning("undefine data: % x, %d, % x\r\n", v.Header, len(v.Content), v.Content)
			}
		}
		wg.Done()
	}()

	return c
}

func createLogFile(time time.Time, directory string) (*os.File, error) {
	filename := fmt.Sprintf(`%s\%s.log`, directory, time.Format(TIME_HOUR_PATTERN))

	return os.OpenFile(filename, os.O_CREATE | os.O_WRONLY | os.O_APPEND, os.FileMode(644))
}

func write(exportDirectory string, writeQueue <- chan string) {
	go func () {
		var file *os.File
		var err error
		beforeHour := ""

		for v := range writeQueue {
			if beforeHour != time.Now().Format(TIME_HOUR_PATTERN) {
				currentTime := time.Now()

				if file != nil {
					file.WriteString(FILE_END_FLAG)
					file.Close()
				}

				file, err = createLogFile(currentTime, exportDirectory)
				if err != nil {
					configuration.WriteError(err, "cannot create log file:")
				} else {
					beforeHour = currentTime.Format(TIME_HOUR_PATTERN)
				}
			}
			file.WriteString(v+"\n")
		}
		wg.Done()
	}()
}

func connectServer(config configuration.Configuration) *tls.Conn {
	ip := config.ServerIP
	port := config.ServiecePort
	pkcs12FileName := config.Pkcs12FileName
	pkcs12Password := config.Pkcs12Password

	certPEMBlock, keyPEMBlock, err := readPkcs12(pkcs12FileName, pkcs12Password)
	if err != nil {
		configuration.WriteError(err, "cannot extract pem block")

		return nil
	}

	cert, err := tls.X509KeyPair(certPEMBlock, keyPEMBlock)
	if err != nil {
		configuration.WriteError(err, "cannot make certification and key")
		return nil
	}

	tlsConfig := tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}

	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", ip, port), &tlsConfig)
	if err != nil {
		configuration.WriteError(err, "cannot connect server to %s:%d", ip, port)
		return nil
	}

	return conn
}

// pkcs12File : certs/192.168.10.232.pkcs12
// string : Cisco123
func readPkcs12(pkcs12File, password string) ([]byte, []byte, error) {
	data, err := ioutil.ReadFile(pkcs12File)
	if err != nil {
		configuration.WriteError(err, "cannot read pkcs12: %s", pkcs12File)
		return nil, nil, err
	}

	blocks, err := pkcs12.ToPEM(data, password)
	if err != nil {
		configuration.WriteError(err, "cannot convert pem block")
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
