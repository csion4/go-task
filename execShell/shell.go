package execShell

import (
	"bufio"
	"golang.org/x/text/encoding/simplifiedchinese"
	"io"
	"os"
	"os/exec"
)

type Charset string

const (
	UTF8    = Charset("UTF-8")
	GB18030 = Charset("GB18030")
)

//执行shell脚本
func ExecShell(cmd string, dir string, file *os.File) {
	_, err := file.Write([]byte("[script]: " + cmd + " \n"))
	if err != nil {
		panic(err)
	}
	command := exec.Command("cmd", "/C", cmd)
	command.Dir = dir

	pipe, err1 := command.StdoutPipe()
	if err1 != nil {
		panic(err1)
	}
	defer pipe.Close()

	if err2 := command.Start(); err2 != nil {
		panic(err2)
	}

	reader := bufio.NewReader(pipe)
	for ;; {
		line, _, err := reader.ReadLine()
		if err == io.EOF {
			return
		} else if err != nil {
			panic(err)
		}

		_, err = file.Write(append(line, ' ', '\n'))
		if err != nil {
			panic(err)
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
	file, err := os.OpenFile("E:\\tasks\\log\\test2\\"+fileName, os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		panic(err)
	}

	return bufio.NewWriter(file), file

}
