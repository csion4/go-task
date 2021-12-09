package task

import (
	"com.csion/tasks/execShell"
	"com.csion/tasks/script"
)


// git交互下代码
func Git(url string, branch string, workDir string){
	execShell.ExecShell("git init & git remote add origin " + url, workDir)
	execShell.ExecShell("git fetch origin", workDir)
	execShell.ExecShell("git checkout -b " + branch + " origin/" + branch, workDir)
}

// 执行脚本
func ExecScript(scripts string, scriptDir string, workDir string){
	filePath := script.CreateTempShell(scriptDir, scripts)
	execShell.ExecShell(filePath, workDir)
	script.DelFile(filePath)
}

// http调用
func HttpInvoke(url string, param string, t string){

}
