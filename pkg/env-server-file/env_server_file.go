package env_server_file

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
)

type Env_server_file struct {
	Usertoken string `json:"usertoken"`
	Web_port  uint16 `json:"web_port"`
}

func (esf *Env_server_file) Write_file(filepath string) error {
	marker_data_file, err := json.Marshal(esf)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath, marker_data_file, 0600)
	if err != nil {
		return err
	}
	return nil
}

func (esf *Env_server_file) Read_file(filepath string) error {
	marker_data_file, err := ioutil.ReadFile(filepath)
	if err != nil {
		fmt.Println("ReadFile")
		return err
	}
	err = json.Unmarshal([]byte(marker_data_file), &esf)
	if err != nil {
		fmt.Println("Unmarshal")
		return err
	}
	return nil
}

func (esf *Env_server_file) Write() (string, error) {
	filepath, err := GetEnvFileName()
	if err != nil {
		return "", err
	}
	return filepath, esf.Write_file(filepath)
}

func (esf *Env_server_file) Read() error {
	filepath, err := GetEnvFileName()
	if err != nil {
		fmt.Println("GetEnvFileName")
		return err
	}
	return esf.Read_file(filepath)
}

func GetEnvFileName() (string, error) {
	filepath := os.Getenv("XDG_RUNTIME_DIR")
	if filepath == "" {
		return "", errors.New("XDG_RUNTIME_DIR not set")
	}

	if runtime.GOOS == "windows" {
		tmp := strings.Replace(filepath, "/", "\\", -1)
		filepath = "Z:" + tmp
		filepath += "\\lscmd"
	} else {
		filepath += "/lscmd"
	}
	return filepath, nil
}
