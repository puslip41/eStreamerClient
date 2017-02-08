package configuration

import (
	"os"
	"fmt"
	"encoding/json"
	"log"
)

const CONFIG_FILE_NAME = `conf/configuration.json`
//const CONFIG_FILE_NAME = `D:/SourceCode/Go/src/github.com/puslip41/eStreamerClient/conf/configuration.json`

type Configuration struct {
	ServerIP string `json:"server ip"`
	ServiecePort int `json:"server port"`
	Pkcs12FileName string `json:"pkcs12 filename"`
	Pkcs12Password string `json:"pkcs12 password"`
	ExportDirectory string `json:"export directory"`
	LogLevel string `json:"log level"`
	RequestFlags string `json:"request flags"`
}

func ReadConfiguration() (Configuration, error) {
	config := Configuration{}
	var err error

	file, err := os.OpenFile(CONFIG_FILE_NAME, os.O_RDONLY, os.FileMode(644))
	if err != nil {
		return Configuration{}, err
	}

	json := json.NewDecoder(file)
	err = json.Decode(&config)
	if err != nil {
		return Configuration{}, err
	}

	return config, nil
}

func WriteDefaultConfiguration() {
	config := Configuration{}

	config.ServerIP = "192.168.1.1"
	config.ServiecePort = 8302
	config.Pkcs12FileName = `/conf/cert/estreamer.pkcs12`
	config.Pkcs12Password = "password"
	config.ExportDirectory = `raw_logs`

	b, err := json.Marshal(config)
	if err != nil {
		log.Println("cannot open config file:", err)
	} else {
		file, _ := os.OpenFile(fmt.Sprintf(`conf/%s`, CONFIG_FILE_NAME), os.O_CREATE|os.O_WRONLY, os.FileMode(644))
		file.Write(b)
	}
}
