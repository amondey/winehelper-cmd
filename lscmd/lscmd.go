package main

import (
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	cmd_query "winehelper-cmd/pkg/cmd_query"
	env_server_file "winehelper-cmd/pkg/env-server-file"
)

//Обьявляем глобально, т.к. надо этот файл удалять из разных мест
var (
	EnvFileName string
	srvData     env_server_file.EnvServerFile
)

func getFreePort() (uint16, bool) {

	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		fmt.Printf("can't ResolveTCPAddr: %v\n", err)
		return 0, true
	}

	ln, err := net.ListenTCP("tcp", addr)
	if err != nil {
		fmt.Printf("can't ListenTCP: %v\n", err)
		return 0, true
	}
	port := ln.Addr().(*net.TCPAddr).Port

	if err = ln.Close(); err != nil {
		fmt.Printf("can't close listener: %v\n", err)
		return 0, true
	}

	return uint16(port), false
}

func quitme(_ http.ResponseWriter, _ *http.Request) {
	if err := os.Remove(EnvFileName); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

func checkCmdAndExec(w http.ResponseWriter, req *http.Request, srvData *env_server_file.EnvServerFile) {
	token := req.URL.Query().Get("token")
	if token == "" {
		if _, err := fmt.Fprintf(w, "cmd_error_access1"); err != nil {
			fmt.Println(err)
		}
		return
	}

	if token != srvData.UserToken {
		if _, err := fmt.Fprintf(w, "cmd_error_access2"); err != nil {
			fmt.Println(err)
		}
		return
	}
	//Из хедера достаем параметр cmd
	cmd := req.Header.Get("lcmd")

	if cmd == "" {
		if _, err := fmt.Fprintf(w, "cmd_error"); err != nil {
			fmt.Println(err)
		}
		fmt.Println("cmd_error_url.query")
		return
	}

	//Декодируем его из base64
	cmdtmp, err := base64.RawStdEncoding.DecodeString(cmd)
	if err != nil {
		if _, err = fmt.Fprintf(w, "cmd_error"); err != nil {
			fmt.Println(err)
		}
		if _, err = fmt.Printf("cmd_error_encode=%s", cmd); err != nil {
			fmt.Println(err)
		}
		return
	}
	cmd = string(cmdtmp)
	fmt.Printf("lcmd:%s\n", cmd)
	//Разбиваем строку на аргументы
	cmdSplitted := csv.NewReader(strings.NewReader(cmd))
	cmdSplitted.Comma = ' ' // space
	args, err := cmdSplitted.Read()
	if err != nil {
		if _, err = fmt.Fprintf(w, "cmd_error"); err != nil {
			fmt.Println(err)
		}
		return
	}

	//fmt.Printf(string(args[0]))
	//fmt.Printf(cmd)
	//Исполняем и возвращяем результат из stdio и exit code
	testBash := req.URL.Query().Get("bash")
	var cmdtest *exec.Cmd
	if _, echo := fmt.Printf("bash: %s\n", testBash); err != nil {
		fmt.Println(echo)
	}

	if testBash == "" {
		cmdtest1 := exec.Command(args[0], args[1:]...)
		cmdtest = cmdtest1
	} else {
		cmdtest1 := exec.Command("/bin/bash", "-c", cmd)
		cmdtest = cmdtest1
	}

	out, _ := cmdtest.CombinedOutput()

	res := cmd_query.CmdResult{}
	res.CmdStdout = string(out)

	res.ErrorCode = uint(cmdtest.ProcessState.ExitCode())

	resJson, _ := json.Marshal(res)
	if _, err = fmt.Fprintf(w, "%s", string(resJson)); err != nil {
		fmt.Println(err)
	}
}

func main() {
	//Генерим uuid
	usertokenB, _ := ioutil.ReadFile("/proc/sys/kernel/random/uuid")
	//Ищём свободный порт
	webPort, isErr := getFreePort()
	if isErr {
		fmt.Printf("can't open tcp port\n")
		os.Exit(1)
	}

	//Пишем файл в окружении пользователя
	//$XDG_RUNTIME_DIR/lscmd
	//Убираем символ конца строки 0x10
	srvData.UserToken = string(usertokenB[:len(usertokenB)-1])
	srvData.WebPort = webPort
	var err error
	EnvFileName, err = srvData.Write()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	//При закрытии, надо удалить файл
	sigs := make(chan os.Signal, 2)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		fmt.Println(sig)
		if err := os.Remove(EnvFileName); err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	}()

	//Запускаем веб сервер
	http.HandleFunc("/exec", func(w http.ResponseWriter, req *http.Request) {
		checkCmdAndExec(w, req, &srvData)
	})
	http.HandleFunc("/quit", quitme)

	fmt.Printf("%v\n", srvData.WebPort)
	fmt.Println(string(srvData.UserToken))
	if err = http.ListenAndServe(":"+fmt.Sprint(srvData.WebPort), nil); err != nil {
		panic(err)
	}

	defer func() {
		if err := os.Remove(EnvFileName); err != nil {
			os.Exit(1)
		}
	}()
}
