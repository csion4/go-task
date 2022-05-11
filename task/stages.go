package task

import (
	"com.csion/tasks/execShell"
	"os"
)

// git交互下代码
func Git(url string, branch string, workDir string, file *os.File){
	execShell.ExecShell("git init", workDir, file)
	execShell.ExecShell("git remote add origin " + url, workDir, file)
	execShell.ExecShell("git fetch origin", workDir, file)
	execShell.ExecShell("git checkout -b " + branch + " origin/" + branch, workDir, file)
}

// 执行脚本
func ExecScript(scripts string, scriptDir string, workDir string, file *os.File){
	filePath := execShell.CreateTempShell(scriptDir, scripts, file)
	execShell.ExecShell(filePath, workDir, file)
	execShell.DelFile(filePath)
}

// http调用
func HttpInvoke(url string, param string, t string){

}

func GetTaskStageChan(taskId string, recordId string) chan int {
	ch := make(chan int)
	chanMap[taskId + recordId] = ch  // 暂时不用缓冲，缓冲可以避免某些异常，但是可能导致gc失败，暂时不用缓冲
	return ch
}
