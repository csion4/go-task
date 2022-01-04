package script

import "os"

// 创建临时shell脚本
func CreateTempShell(scriptDir string, scripts string) string {

	file, err := os.Create(scriptDir + "/temp.bat")
	if err != nil {
		panic(err)
	}

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
