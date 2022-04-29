package cluster

import (
	"fmt"
	"github.com/pkg/sftp"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
	"io"
	"os"
	"time"
)

const worker = "taskCluster"

func Track(ip string, userName string, password string) string  {
	// 发送worker client
	sshClient, err := ssh.Dial("tcp", ip + ":22", &ssh.ClientConfig{
		User:            userName,
		Auth:            []ssh.AuthMethod{ssh.Password(password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	})
	if err != nil {
		panic(err)
	}
	defer sshClient.Close()

	var targetPath string
	if userName == "root" {
		targetPath = "/root"
	} else {
		targetPath = "/home/" + userName
	}

	sendTaskCluster(sshClient, targetPath)

	// 启动服务
	session, err := sshClient.NewSession()
	if err != nil {
		panic(err)
	}
	defer session.Close()

	if err := session.Run("chmod 777 taskCluster && ./taskCluster"); err != nil {
		fmt.Println("Failed to run: " + err.Error())
	}

	return getPort(sshClient, targetPath)
}

func sendTaskCluster(sshClient *ssh.Client, targetPath string) {
	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		panic(err)
	}
	defer sftpClient.Close()

	// 发送客户端包
	wd, _ := os.Getwd()
	src, _ := os.Open(wd + "/cluster/" + worker)
	defer src.Close()
	dst, err := sftpClient.OpenFile(targetPath + "/" + worker, os.O_CREATE|os.O_RDWR)
	if err != nil {
		panic(err)
	}
	defer dst.Close()
	_, err = io.Copy(dst, src)
	if err != nil {
		panic(err)
	}

	// 发送配置文件
	dst, err = sftpClient.OpenFile(targetPath + "/" + worker + ".conf", os.O_CREATE|os.O_RDWR)
	if err != nil {
		panic(err)
	}
	defer dst.Close()
	_, err = dst.Write([]byte("MNode=" + viper.GetString("task.worker.MNode") +
		"\nTaskHome=" + viper.GetString("task.worker.TaskHome") +
		"\nAuth=" + viper.GetString("task.worker.Auth")))
	if err != nil {
		panic(err)
	}
}

func getPort(sshClient *ssh.Client, targetPath string) string  {
	time.Sleep(2e9)
	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		panic(err)
	}
	defer sftpClient.Close()
	// 解析日志获取port
	port, err := sftpClient.Open(targetPath + "/" + worker + ".port")
	if err != nil {
		panic(err)
	}
	buff, _ := io.ReadAll(port)
	return string(buff)
}
