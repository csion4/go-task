package execShell

import (
	"os"
	"strings"
)

// 创建临时shell脚本
func CreateTempShell(scriptDir string, scripts string, file *os.File) string {

	var tempFile string
	if strings.Contains(os.Getenv("os"), "Windows") {
		tempFile = "/temp.bat"
	} else {
		tempFile = "/temp.sh"
	}

	file, err := os.Create(scriptDir + tempFile)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	_, _ = file.WriteString(scripts)
	_ = file.Close()

	return file.Name()
}
func DelFile(file string){
	err := os.Remove(file)
	if err != nil {
		panic(err)
	}
}
