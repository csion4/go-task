package cluster

import (
	"com.csion/tasks/common"
	"com.csion/tasks/dto"
	"com.csion/tasks/tLog"
	"fmt"
	"github.com/pkg/sftp"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

const worker = "taskCluster"
var log = tLog.GetTLog()

// 添加node
func Track(ip string, userName string, password string, taskHome string) string  {
	// 发送worker client
	sshClient, err := ssh.Dial("tcp", ip + ":22", &ssh.ClientConfig{
		User:            userName,
		Auth:            []ssh.AuthMethod{ssh.Password(password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout: time.Second * 3,
	})
	log.Panic2("节点连接异常", err)
	defer sshClient.Close()

	var targetPath string
	if userName == "root" {
		targetPath = "/root"
	} else {
		targetPath = "/home/" + userName
	}

	sendTaskCluster(sshClient, targetPath, taskHome)

	// 启动服务
	session, err := sshClient.NewSession()
	log.Panic2("节点连接异常", err)
	defer session.Close()

	if err := session.Start("./taskCluster"); err != nil {
		log.Panic2("worker节点服务启动异常", err)
	}

	return getPort(sshClient, targetPath)
}

// node埋点
func sendTaskCluster(sshClient *ssh.Client, targetPath string, taskHome string) {
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
		"\nTaskHome=" + taskHome +
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
	port, err := sftpClient.OpenFile(targetPath + "/" + worker + ".port", os.O_CREATE|os.O_RDWR)
	log.Panic2("节点连接异常", err)
	buff, _ := io.ReadAll(port)
	return string(buff)
}

// 对工作节点进行探测，todo：两种策略，每一个node添加一个，或者是所有的node公用一个协程；不管使用哪种，这里都不应该使用固定的ip、port入参，而是使用id查询的方式，因为节点信息会变；- 考虑到每个worker node在异常时需要尝试恢复，所以选择每个node都有单独的监控协程作用与监控和异常恢复
func NodeProbe(id int) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				return
			}
		}()
		for {
			time.Sleep(time.Second * 5)
			var n dto.WorkerNode
			log.Panic2("数据操作异常", common.GetDb().Raw("select id, ip, port, user_name, password, task_home from worker_nodes where id = ? and status = 1", id).Scan(&n).Error)
			if n.Id == 0 {
				return
			}
			if ping(n.Ip, n.Port, 3) != nil {
				log.Panic2("数据操作异常", common.GetDb().Exec("update worker_nodes set node_status = 2 where id = ?", id).Error)
				// 恢复节点
				for {
					time.Sleep(time.Second * 5)
					port := CheckNode(n.Ip, n.UserName, n.Password, n.TaskHome)
					if port != "" {
						p, _ := strconv.Atoi(port)
						if p != n.Port {
							log.Panic2("数据操作异常", common.GetDb().Exec("update worker_nodes set port = ? where id = ?", p, n.Id).Error)
						}
						log.Panic2("数据操作异常", common.GetDb().Exec("update worker_nodes set node_status = 1 where id = ?", n.Id).Error)
						break
					}
				}
			}
		}
	}()
}

func CheckNode(ip string, userName string, password string, taskHome string) string {
	defer func() {
		if err := recover();err != nil {
			return
		}
	}()
	return Track(ip, userName, password, taskHome)
}

func ping(ip string, port int, i int) error {
	// 发送任务
	client := http.Client{
		Timeout: time.Second * 2,
	}
	r, err := client.Get(fmt.Sprintf("http://%s:%d/ping", ip, port))
	if err != nil {
		if i == 1 {
			return err
		}
		return ping(ip, port, i - 1)
	}
	defer r.Body.Close()
	return nil
}
