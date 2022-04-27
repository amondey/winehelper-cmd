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
var Env_file_name string

func get_free_port() (uint16, bool) {

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

	ln.Close()

	return uint16(port), false
}

func quitme(_ http.ResponseWriter, _ *http.Request) {
	os.Remove(Env_file_name)
	os.Exit(0)
}

func checkcmd_and_exec(w http.ResponseWriter, req *http.Request, Srv_data *env_server_file.Env_server_file) {
	token := req.URL.Query().Get("token")
	if token == "" {
		fmt.Fprintf(w, "cmd_error_access1")
		return
	}

	if token != Srv_data.Usertoken {
		fmt.Fprintf(w, "cmd_error_access2")
		return
	}
	//Из урла достаем параметр cmd
	cmd := req.URL.Query().Get("cmd")
	if cmd == "" {
		fmt.Fprintf(w, "cmd_error")
		fmt.Println("cmd_error_url.query")
		return
	}
	//Декодируем его из base64
	cmdtmp, err := base64.RawStdEncoding.DecodeString(cmd)
	if err != nil {
		fmt.Fprintf(w, "cmd_error")
		fmt.Printf("cmd_error_encode=%s", cmd)
		return
	}
	cmd = string(cmdtmp)
	fmt.Printf("decoded_cmd=%s\n", cmd)
	//Разбиваем строку на аргументы
	cmd_splitted := csv.NewReader(strings.NewReader(cmd))
	cmd_splitted.Comma = ' ' // space
	args, err := cmd_splitted.Read()
	if err != nil {
		fmt.Fprintf(w, "cmd_error")
		return
	}

	//fmt.Printf(string(args[0]))
	//fmt.Printf(cmd)
	//Исполняем и возвращяем результат из stdio и exit code
	test_bash := req.URL.Query().Get("bash")

	var cmdtest *exec.Cmd

	if test_bash == "" {
		cmdtest1 := exec.Command(args[0], args[1:]...)
		cmdtest = cmdtest1
	} else {
		cmdtest1 := exec.Command("/bin/bash", "-c", cmd)
		cmdtest = cmdtest1
	}

	out, _ := cmdtest.CombinedOutput()
	res := cmd_query.Cmd_result{}
	res.Cmd_stdout = string(out)
	res.Error_code = uint(cmdtest.ProcessState.ExitCode())

	res_json, _ := json.Marshal(res)
	fmt.Fprintf(w, "%s", string(res_json))
}

func main() {
	var Srv_data env_server_file.Env_server_file

	//Генерим uuid
	usertoken_b, _ := ioutil.ReadFile("/proc/sys/kernel/random/uuid")
	//Ищём свободный порт
	_web_port, is_err := get_free_port()
	if is_err {
		fmt.Printf("can't open tcp port\n")
		os.Exit(1)
	}

	//Пишем файл в окружении пользователя
	//$XDG_RUNTIME_DIR/lscmd
	//Убираем символ конца строки 0x10
	Srv_data.Usertoken = string(usertoken_b[:len(usertoken_b)-1])
	Srv_data.Web_port = _web_port
	var err error
	Env_file_name, err = Srv_data.Write()
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
		os.Remove(Env_file_name)
		os.Exit(0)
	}()

	//Запускаем веб сервер
	http.HandleFunc("/exec", func(w http.ResponseWriter, req *http.Request) {
		checkcmd_and_exec(w, req, &Srv_data)
	})
	http.HandleFunc("/quit", quitme)

	fmt.Printf("%v\n", Srv_data.Web_port)
	fmt.Println(string(Srv_data.Usertoken))
	http.ListenAndServe(":"+fmt.Sprint(Srv_data.Web_port), nil)

	defer func() {
		os.Remove(Env_file_name)
	}()
}
