package cmdquery

type Cmd_result struct {
	Error_code uint   `json:"error_code"`
	Cmd_stdout string `json:"cmd_stdout"`
}
