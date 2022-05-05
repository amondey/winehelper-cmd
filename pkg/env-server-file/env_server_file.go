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

type EnvServerFile struct {
	UserToken string `json:"usertoken"`
	WebPort   uint16 `json:"web_port"`
}

func (esf *EnvServerFile) WriteFile(filepath string) error {
	markerDataFile, err := json.Marshal(esf)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath, markerDataFile, 0600)
	if err != nil {
		return err
	}
	return nil
}

func (esf *EnvServerFile) ReadFile(filepath string) error {
	markerDataFile, err := ioutil.ReadFile(filepath)
	if err != nil {
		fmt.Println("ReadFile")
		return err
	}
	err = json.Unmarshal(markerDataFile, &esf)
	if err != nil {
		fmt.Println("Unmarshal")
		return err
	}
	return nil
}

func (esf *EnvServerFile) Write() (string, error) {
	filepath, err := GetEnvFileName()
	if err != nil {
		return "", err
	}
	return filepath, esf.WriteFile(filepath)
}

func (esf *EnvServerFile) Read() error {
	filepath, err := GetEnvFileName()
	if err != nil {
		fmt.Println("GetEnvFileName")
		return err
	}
	return esf.ReadFile(filepath)
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
