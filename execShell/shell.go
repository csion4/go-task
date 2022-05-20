package execShell

import (
	"bufio"
	"com.csion/tasks/tLog"
	"golang.org/x/text/encoding/simplifiedchinese"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

type Charset string

const (
	UTF8    = Charset("UTF-8")
	GB18030 = Charset("GB18030")
)

var log = tLog.GetTLog()

//执行shell脚本
func ExecShell(cmd string, dir string, file *os.File) {
	_, err := file.Write([]byte("[script]: " + cmd + " \n"))
	log.Panic2("日志写入异常", err)
	log.Debug("执行脚本", cmd)

	var command *exec.Cmd
	if strings.Contains(os.Getenv("os"), "Windows"){
		command = exec.Command("cmd", "/C", cmd)
	} else {
		command = exec.Command("/bin/sh", "-c", cmd)
	}
	command.Dir = dir

	errPipe, err := command.StderrPipe()
	checkErr("获取脚本执行结果异常", err, file)
	defer errPipe.Close()

	pipe, err := command.StdoutPipe()
	checkErr("获取脚本执行结果异常", err, file)
	defer pipe.Close()

	checkErr("脚本执行异常", command.Start(), file)

	errOut, err := ioutil.ReadAll(errPipe)
	checkErr("获取脚本执行结果异常", err, file)
	if len(errOut) > 0 {
		_, e := file.Write([]byte("【ERROR】脚本执行异常" + string(errOut) + " \n"))
		log.Panic2("日志写入异常", e)
	}

	reader := bufio.NewReader(pipe)
	for {
		line, _, err := reader.ReadLine()
		if err == io.EOF {
			return
		} else if err != nil {
			checkErr("脚本执行异常", err, file)
		}

		_, err = file.Write(append(line, ' ', '\n'))
		if err != nil {
			log.Panic2("日志写入异常", err)
		}

	}



	//reader := bufio.NewReader(pipe)
	// writer, file := logToFile("log1.log")
	//for ;; {
	//	line, _, err := reader.ReadLine()
	//	if err == io.EOF {
	//		_ = writer.Flush()
	//		_ = file.Close()
	//		return
	//	}
	//	byte2String := convertByte2String(line, "GB18030")
	//	fmt.Println(byte2String)
	//	_, err = writer.WriteString(byte2String + "\n")
	//	if err != nil {
	//		panic(err)
	//	}
	//	// _ = writer.Flush() // 这里不用做成实时刷新到file中
	//}

}

// 异常校验，结果写入到执行日志和系统日志中
func checkErr(s string, err error, logFile *os.File) {
	if err != nil {
		_, e := logFile.Write([]byte("【ERROR】 " + s + err.Error() + " \n"))
		log.Panic2("日志写入异常", e)
		log.Panic2(s, err)
	}
}

func convertByte2String(byte []byte, charset Charset) string {
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

func logToFile (fileName string) (*bufio.Writer, *os.File) {
	file, err := os.OpenFile("E:\\tasks\\log\\test2\\"+fileName, os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}

	return bufio.NewWriter(file), file

}
