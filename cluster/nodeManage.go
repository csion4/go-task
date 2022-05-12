package cluster

import (
	"com.csion/tasks/common"
	"com.csion/tasks/tLog"
	"fmt"
	"github.com/pkg/sftp"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
	"io"
	"net/http"
	"os"
	"time"
)

const worker = "taskCluster"
var log = tLog.GetTLog()

// 添加node
func Track(ip string, userName string, password string) string  {
	// 发送worker client
	sshClient, err := ssh.Dial("tcp", ip + ":22", &ssh.ClientConfig{
		User:            userName,
		Auth:            []ssh.AuthMethod{ssh.Password(password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	})
	log.Panic2("节点连接异常", err)
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
	log.Panic2("节点连接异常", err)
	defer session.Close()

	if err := session.Run("chmod 777 taskCluster && ./taskCluster"); err != nil {
		fmt.Println("Failed to run: " + err.Error())
	}

	return getPort(sshClient, targetPath)
}

// node埋点
func sendTaskCluster(sshClient *ssh.Client, targetPath string) {
	sftpClient, err := sftp.NewClient(sshClient)
	log.Panic2("节点连接异常", err)
	defer sftpClient.Close()

	// 发送客户端包
	wd, _ := os.Getwd()
	src, _ := os.Open(wd + "/cluster/" + worker)
	defer src.Close()
	dst, err := sftpClient.OpenFile(targetPath + "/" + worker, os.O_CREATE|os.O_RDWR)
	log.Panic2("节点连接异常", err)
	defer dst.Close()
	_, err = io.Copy(dst, src)
	log.Panic2("节点连接异常", err)

	// 发送配置文件
	dst, err = sftpClient.OpenFile(targetPath + "/" + worker + ".conf", os.O_CREATE|os.O_RDWR)
	log.Panic2("节点连接异常", err)
	defer dst.Close()
	_, err = dst.Write([]byte("MNode=" + viper.GetString("task.worker.MNode") +
		"\nTaskHome=" + viper.GetString("task.worker.TaskHome") +
		"\nAuth=" + viper.GetString("task.worker.Auth")))
	log.Panic2("节点连接异常", err)
}

// 获取node服务port
func getPort(sshClient *ssh.Client, targetPath string) string  {
	time.Sleep(2e9)
	sftpClient, err := sftp.NewClient(sshClient)
	log.Panic2("节点连接异常", err)
	defer sftpClient.Close()
	// 解析日志获取port
	port, err := sftpClient.Open(targetPath + "/" + worker + ".port")
	log.Panic2("节点连接异常", err)
	buff, _ := io.ReadAll(port)
	return string(buff)
}

// 对工作节点进行探测，todo：两种策略，每一个node添加一个，或者是所有的node公用一个协程；不管使用哪种，这里都不应该使用固定的ip、port入参，而是使用id查询的方式，因为节点信息会变
func NodeProbe(id int, ip string, port int) {
	go func() {
		time.Sleep(time.Second * 10)
		var i int
		log.Error("数据操作异常", common.GetDb().Raw("select count(1) from worker_nodes where id = ? and status = 1", id).Scan(&i).Error)
		if i == 0 {
			return
		}
		if ping(ip, port, 3) != nil {
			log.Error("数据操作异常", common.GetDb().Exec("update worker_nodes set node_status = 2 where id = ").Error)
		}
	}()
}

func ping(ip string, port int, i int) error {
	// 发送任务
	r, err := http.Get(fmt.Sprintf("http://%s:%d/task", ip, port))
	if err != nil {
		if i == 0 {
			return err
		}
		return ping(ip, port, i - 1)
	}
	defer r.Body.Close()
	return nil
}
