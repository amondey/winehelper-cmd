package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
	cmd_query "winehelper-cmd/pkg/cmd_query"
	env_server_file "winehelper-cmd/pkg/env-server-file"
)

func runLscmd() {
	cmd := exec.Command("start", "/unix", "/usr/sbin/winehelper_lscmd")
	err := cmd.Run()
	if err != nil {
		fmt.Printf("error run lscmd\n")
		os.Exit(1)
	}
}

func main() {

	if len(os.Args) < 2 {
		fmt.Printf("error args\n")

		os.Exit(1)
	}

	srvData := env_server_file.EnvServerFile{}
	err := srvData.Read()
	if err != nil {
		runLscmd()
		time.Sleep(2 * time.Second)
		err := srvData.Read()
		if err != nil {
			fmt.Println(err.Error())
			fmt.Printf("error\n")
			os.Exit(1)
		}
	}

	//заново формируем переданные аргументы, как raw строку
	cmd := ""
	for i, s := range os.Args {
		if i > 0 {
			stradd := ""
			if i == 1 {
				stradd = ""
			} else {
				stradd = " "
			}

			if strings.Contains(s, " ") {
				cmd = cmd + stradd + "\"" + s + "\""
			} else {
				cmd = cmd + stradd + s
			}
		}
	}

	//Если команда завернута в " то используем bash
	//К примеру "ls -la|grep test", мы вырежем ковычки: ls -la|grep test
	useBash := false
	if cmd[0] == 34 {
		if cmd[len(cmd)-1] == 34 {
			cmd = cmd[1 : len(cmd)-1]
			useBash = true
		}
	}

	encodedCmd := base64.RawStdEncoding.EncodeToString([]byte(cmd))

	url := fmt.Sprintf("http://localhost:%d/exec?token=%s",
		srvData.WebPort,
		srvData.UserToken,
	)

	if useBash {
		url += "&bash=true"
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("error 2\n")
		os.Exit(1)
	}

	req.Header.Add("lcmd", encodedCmd)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("error 3\n")
		os.Exit(1)
	}

	result, _ := ioutil.ReadAll(resp.Body)

	resultJson := cmd_query.CmdResult{}
	if err = json.Unmarshal(result, &resultJson); err != nil {
		os.Exit(-1)
	}
	fmt.Println(string(resultJson.CmdStdout))

	//url = fmt.Sprintf("http://localhost:%d/quit", marker_data.Web_port)
	//http.Get(url)
	os.Exit(int(resultJson.ErrorCode))
}
