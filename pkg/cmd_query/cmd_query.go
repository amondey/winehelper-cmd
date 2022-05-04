package cmdquery

type CmdResult struct {
	ErrorCode uint   `json:"error_code"`
	CmdStdout string `json:"cmd_stdout"`
}
