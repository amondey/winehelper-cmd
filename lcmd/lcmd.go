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

func run_lscmd() {
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

	Srv_data := env_server_file.Env_server_file{}
	err := Srv_data.Read()
	if err != nil {
		run_lscmd()
		time.Sleep(2 * time.Second)
		err := Srv_data.Read()
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
	use_bash := false
	if cmd[0] == 34 {
		if cmd[len(cmd)-1] == 34 {
			cmd = cmd[1 : len(cmd)-1]
			use_bash = true
		}
	}

	encoded_cmd := base64.RawStdEncoding.EncodeToString([]byte(cmd))

	url := fmt.Sprintf("http://localhost:%d/exec?token=%s&cmd=%s",
		Srv_data.Web_port,
		Srv_data.Usertoken,
		encoded_cmd)

	if use_bash {
		url += "&bash=true"
	}

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("error\n")
		os.Exit(1)
	}

	result, _ := ioutil.ReadAll(resp.Body)

	result_json := cmd_query.Cmd_result{}
	json.Unmarshal(result, &result_json)
	fmt.Println(string(result_json.Cmd_stdout))

	//url = fmt.Sprintf("http://localhost:%d/quit", marker_data.Web_port)
	//http.Get(url)
	os.Exit(int(result_json.Error_code))
}
