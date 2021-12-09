package execShell

import (
	"bufio"
	"fmt"
	"golang.org/x/text/encoding/simplifiedchinese"
	"io"
	"os/exec"
)

type Charset string

const (
	UTF8    = Charset("UTF-8")
	GB18030 = Charset("GB18030")
)

//执行shell脚本
func ExecShell(cmd string, dir string) {

	// command := exec.Command("cmd", "/C", "ping www.baidu.com -n 5")
	//command := exec.Command("cmd", "/C", "E:/工作文件/联通数科/project/temp/temp.bat")
	command := exec.Command("cmd", "/C", cmd)

	// command.Dir = "E:/工作文件/联通数科/project/inner_source_scm/"
	command.Dir = dir

	pipe, err1 := command.StdoutPipe()
	if err1 != nil {
		panic(err1)
	}
	if err2 := command.Start(); err2 != nil {
		panic(err2)
	}

	reader := bufio.NewReader(pipe)
	for ;; {
		line, _, err := reader.ReadLine()
		if err == io.EOF {
			return
		}
		byte2String := ConvertByte2String(line, "GB18030")
		fmt.Println(byte2String)
	}
}

func ConvertByte2String(byte []byte, charset Charset) string {
	var str string
	switch charset {
	case GB18030:
		var decodeBytes, _ = simplifiedchinese.GB18030.NewDecoder().Bytes(byte)
		str = string(decodeBytes)
	case UTF8:
		fallthrough
	default:
		str = string(byte)
	}
	return str
}
